package dag

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when a named DAG does not exist.
var ErrNotFound = errors.New("dag not found")

// Store persists DAGs as a relational graph across three tables:
//
//	dag_registry  — one row per named DAG, plus the active lifecycle flag
//	nodes         — graph vertices, unique by (dag_id, name)
//	edges         — routes[], referencing nodes by id, in declaration order
//
// The YAML wire schema (DAG) is mapped to/from this graph. A route's `to`
// is a node name; it is resolved to a node id by matching nodes.name.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore returns a Store backed by pool.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Save creates or replaces the DAG stored under name. Nodes and edges are
// rewritten atomically in a single transaction. The active flag is left
// untouched for existing DAGs and defaults to false for new ones.
func (s *Store) Save(ctx context.Context, name string, d DAG) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	var dagID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO dag_registry (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET updated_at = now()
		RETURNING id`, name).Scan(&dagID)
	if err != nil {
		return fmt.Errorf("upsert dag_registry: %w", err)
	}

	// Replace the graph: cascading delete of nodes clears edges too.
	if _, err := tx.Exec(ctx, `DELETE FROM nodes WHERE dag_id = $1`, dagID); err != nil {
		return fmt.Errorf("clear nodes: %w", err)
	}

	nodeIDByName := make(map[string]int64, len(d.Nodes))
	for _, n := range d.Nodes {
		var nodeID int64
		err := tx.QueryRow(ctx, `
			INSERT INTO nodes (dag_id, name, topic, host, port, entry, tools)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id`,
			dagID, n.Name, n.Topic, n.Host, n.Port, n.Entry, n.Tools).Scan(&nodeID)
		if err != nil {
			return fmt.Errorf("insert node %q: %w", n.Name, err)
		}
		nodeIDByName[n.Name] = nodeID
	}

	for _, n := range d.Nodes {
		fromID := nodeIDByName[n.Name]
		for pos, route := range n.Routes {
			toID, ok := nodeIDByName[route.To]
			if !ok {
				return fmt.Errorf("route target %q has no matching node", route.To)
			}
			var field, op, value *string
			if route.When != nil {
				field, op, value = &route.When.Field, &route.When.Op, &route.When.Value
			}
			_, err := tx.Exec(ctx, `
				INSERT INTO edges (dag_id, from_node_id, to_node_id, position, when_field, when_op, when_value)
				VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				dagID, fromID, toID, pos, field, op, value)
			if err != nil {
				return fmt.Errorf("insert edge %q -> %q: %w", n.Name, route.To, err)
			}
		}
	}

	return tx.Commit(ctx)
}

// Get reassembles the DAG stored under name from its nodes and edges.
func (s *Store) Get(ctx context.Context, name string) (DAG, error) {
	var dagID int64
	err := s.pool.QueryRow(ctx,
		`SELECT id FROM dag_registry WHERE name = $1`, name).Scan(&dagID)
	if errors.Is(err, pgx.ErrNoRows) {
		return DAG{}, ErrNotFound
	}
	if err != nil {
		return DAG{}, err
	}
	return s.assemble(ctx, dagID)
}

// List returns the names of all stored DAGs. If activeOnly is true, only
// DAGs with active = true are returned.
func (s *Store) List(ctx context.Context, activeOnly bool) ([]string, error) {
	query := `SELECT name FROM dag_registry`
	if activeOnly {
		query += ` WHERE active = true`
	}
	query += ` ORDER BY name`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

// SetActive flips the active lifecycle flag for the named DAG. Returns
// ErrNotFound if no such DAG exists.
func (s *Store) SetActive(ctx context.Context, name string, active bool) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE dag_registry SET active = $2, updated_at = now() WHERE name = $1`, name, active)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes the DAG stored under name (nodes/edges cascade). Returns
// ErrNotFound if no such DAG exists.
func (s *Store) Delete(ctx context.Context, name string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM dag_registry WHERE name = $1`, name)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) assemble(ctx context.Context, dagID int64) (DAG, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, topic, host, port, entry, tools FROM nodes WHERE dag_id = $1 ORDER BY id`, dagID)
	if err != nil {
		return DAG{}, err
	}
	var d DAG
	nameByNodeID := make(map[int64]string)
	nodeIdxByName := make(map[string]int)
	for rows.Next() {
		var id int64
		var n Node
		if err := rows.Scan(&id, &n.Name, &n.Topic, &n.Host, &n.Port, &n.Entry, &n.Tools); err != nil {
			rows.Close()
			return DAG{}, err
		}
		nameByNodeID[id] = n.Name
		nodeIdxByName[n.Name] = len(d.Nodes)
		n.Routes = []Route{}
		d.Nodes = append(d.Nodes, n)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return DAG{}, err
	}

	erows, err := s.pool.Query(ctx, `
		SELECT from_node_id, to_node_id, when_field, when_op, when_value
		FROM edges WHERE dag_id = $1 ORDER BY from_node_id, position`, dagID)
	if err != nil {
		return DAG{}, err
	}
	defer erows.Close()
	for erows.Next() {
		var fromID, toID int64
		var field, op, value *string
		if err := erows.Scan(&fromID, &toID, &field, &op, &value); err != nil {
			return DAG{}, err
		}
		route := Route{To: nameByNodeID[toID]}
		if field != nil {
			route.When = &When{Field: *field, Op: *op, Value: *value}
		}
		idx := nodeIdxByName[nameByNodeID[fromID]]
		d.Nodes[idx].Routes = append(d.Nodes[idx].Routes, route)
	}
	if err := erows.Err(); err != nil {
		return DAG{}, err
	}

	return d, nil
}
