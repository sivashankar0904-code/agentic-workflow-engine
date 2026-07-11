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
//	dag_registry  — one row per named DAG
//	nodes         — services (graph vertices), unique by (dag_id, name)
//	edges         — routing rules (graph edges), referencing nodes by id
//
// The YAML wire schema (DAG) is mapped to/from this graph. A DAG's routing
// source and each rule's target are topic strings; they are resolved to nodes
// by matching nodes.topic.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore returns a Store backed by pool.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Save creates or replaces the DAG stored under name. Nodes and edges are
// rewritten atomically in a single transaction.
func (s *Store) Save(ctx context.Context, name string, d DAG) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	// Upsert the registry row, keeping updated_at fresh, and get its id.
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

	// Insert nodes, tracking node ids by topic (targets/source are topics).
	nodeIDByTopic := make(map[string]int64, len(d.Services))
	for _, svc := range d.Services {
		var nodeID int64
		err := tx.QueryRow(ctx, `
			INSERT INTO nodes (dag_id, name, host, port, topic)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`,
			dagID, svc.Name, svc.Host, svc.Port, svc.Topic).Scan(&nodeID)
		if err != nil {
			return fmt.Errorf("insert node %q: %w", svc.Name, err)
		}
		nodeIDByTopic[svc.Topic] = nodeID
	}

	// Insert edges: source topic -> target topic, resolved to node ids.
	fromID, ok := nodeIDByTopic[d.Routing.Source]
	if !ok && len(d.Routing.Rules) > 0 {
		return fmt.Errorf("routing source %q has no matching service topic", d.Routing.Source)
	}
	for _, rule := range d.Routing.Rules {
		toID, ok := nodeIDByTopic[rule.Target]
		if !ok {
			return fmt.Errorf("rule target %q has no matching service topic", rule.Target)
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO edges (dag_id, from_node_id, to_node_id, condition_field, condition_contains)
			VALUES ($1, $2, $3, $4, $5)`,
			dagID, fromID, toID, rule.Condition.Field, rule.Condition.Contains)
		if err != nil {
			return fmt.Errorf("insert edge -> %q: %w", rule.Target, err)
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

	// Nodes → services, plus id→topic for edge reassembly.
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, host, port, topic FROM nodes WHERE dag_id = $1 ORDER BY id`, dagID)
	if err != nil {
		return DAG{}, err
	}
	var d DAG
	topicByNodeID := make(map[int64]string)
	for rows.Next() {
		var id int64
		var svc Service
		if err := rows.Scan(&id, &svc.Name, &svc.Host, &svc.Port, &svc.Topic); err != nil {
			rows.Close()
			return DAG{}, err
		}
		topicByNodeID[id] = svc.Topic
		d.Services = append(d.Services, svc)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return DAG{}, err
	}

	// Edges → routing rules; source is the shared from-node topic.
	erows, err := s.pool.Query(ctx, `
		SELECT from_node_id, to_node_id, condition_field, condition_contains
		FROM edges WHERE dag_id = $1 ORDER BY id`, dagID)
	if err != nil {
		return DAG{}, err
	}
	defer erows.Close()
	for erows.Next() {
		var fromID, toID int64
		var rule Rule
		if err := erows.Scan(&fromID, &toID, &rule.Condition.Field, &rule.Condition.Contains); err != nil {
			return DAG{}, err
		}
		d.Routing.Source = topicByNodeID[fromID]
		rule.Target = topicByNodeID[toID]
		d.Routing.Rules = append(d.Routing.Rules, rule)
	}
	if err := erows.Err(); err != nil {
		return DAG{}, err
	}

	return d, nil
}

// List returns the names of all stored DAGs.
func (s *Store) List(ctx context.Context) ([]string, error) {
	rows, err := s.pool.Query(ctx, `SELECT name FROM dag_registry ORDER BY name`)
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
