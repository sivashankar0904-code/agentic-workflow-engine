package dag

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when a DAG does not exist.
var ErrNotFound = errors.New("dag not found")

// ListEntry is one row of the DAG registry list: the stable id the UI and
// engines address a DAG by, its human name, and its active lifecycle flag.
type ListEntry struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

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

// Save creates or replaces the DAG stored under name, returning its id.
// Nodes and edges are rewritten atomically in a single transaction. The
// active flag is left untouched for existing DAGs and defaults to false for
// new ones. Upload stays name-addressed (a DAG is a named workflow), but the
// generated id is what the UI and engines address it by thereafter.
func (s *Store) Save(ctx context.Context, name string, d DAG) (int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	var dagID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO dag_registry (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET updated_at = now()
		RETURNING id`, name).Scan(&dagID)
	if err != nil {
		return 0, fmt.Errorf("upsert dag_registry: %w", err)
	}

	// Replace the graph: cascading delete of nodes clears edges too.
	if _, err := tx.Exec(ctx, `DELETE FROM nodes WHERE dag_id = $1`, dagID); err != nil {
		return 0, fmt.Errorf("clear nodes: %w", err)
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
			return 0, fmt.Errorf("insert node %q: %w", n.Name, err)
		}
		nodeIDByName[n.Name] = nodeID
	}

	for _, n := range d.Nodes {
		fromID := nodeIDByName[n.Name]
		for pos, route := range n.Routes {
			toID, ok := nodeIDByName[route.To]
			if !ok {
				return 0, fmt.Errorf("route target %q has no matching node", route.To)
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
				return 0, fmt.Errorf("insert edge %q -> %q: %w", n.Name, route.To, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return dagID, nil
}

// Get reassembles the DAG with the given id from its nodes and edges.
func (s *Store) Get(ctx context.Context, id int64) (DAG, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM dag_registry WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return DAG{}, err
	}
	if !exists {
		return DAG{}, ErrNotFound
	}
	return s.assemble(ctx, id)
}

// List returns all stored DAGs as {id, name, active}. If activeOnly is true,
// only DAGs with active = true are returned.
func (s *Store) List(ctx context.Context, activeOnly bool) ([]ListEntry, error) {
	query := `SELECT id, name, active FROM dag_registry`
	if activeOnly {
		query += ` WHERE active = true`
	}
	query += ` ORDER BY name`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []ListEntry{}
	for rows.Next() {
		var e ListEntry
		if err := rows.Scan(&e.ID, &e.Name, &e.Active); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// ListForRole returns the stored DAGs visible to roleName: every DAG that
// has a dag_roles grant for roleName. If activeOnly is true, the result is
// further restricted to active = true. Filters in SQL (not post-filter) so
// unauthorized DAGs are simply absent from the list, rather than fetched and
// then hidden.
func (s *Store) ListForRole(ctx context.Context, roleName string, activeOnly bool) ([]ListEntry, error) {
	query := `
		SELECT DISTINCT r.id, r.name, r.active FROM dag_registry r
		JOIN dag_roles dr ON dr.dag_id = r.id
		JOIN roles ro ON ro.id = dr.role_id
		WHERE ro.name = $1`
	if activeOnly {
		query += ` AND r.active = true`
	}
	query += ` ORDER BY r.name`

	rows, err := s.pool.Query(ctx, query, roleName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []ListEntry{}
	for rows.Next() {
		var e ListEntry
		if err := rows.Scan(&e.ID, &e.Name, &e.Active); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// Meta returns the {id, name, active} of the DAG with the given id.
func (s *Store) Meta(ctx context.Context, id int64) (ListEntry, error) {
	var e ListEntry
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, active FROM dag_registry WHERE id = $1`, id).
		Scan(&e.ID, &e.Name, &e.Active)
	if errors.Is(err, pgx.ErrNoRows) {
		return ListEntry{}, ErrNotFound
	}
	return e, err
}

// VisibleToRole reports whether roleName has a dag_roles grant on the DAG
// with the given id.
func (s *Store) VisibleToRole(ctx context.Context, id int64, roleName string) (bool, error) {
	var visible bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM dag_roles dr
			JOIN roles ro ON ro.id = dr.role_id
			WHERE dr.dag_id = $1 AND ro.name = $2
		)`, id, roleName).Scan(&visible)
	if err != nil {
		return false, err
	}
	return visible, nil
}

// SetActive flips the active lifecycle flag for the DAG with the given id.
// Returns ErrNotFound if no such DAG exists.
func (s *Store) SetActive(ctx context.Context, id int64, active bool) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE dag_registry SET active = $2, updated_at = now() WHERE id = $1`, id, active)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes the DAG with the given id (nodes/edges cascade). Returns
// ErrNotFound if no such DAG exists.
func (s *Store) Delete(ctx context.Context, id int64) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM dag_registry WHERE id = $1`, id)
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
