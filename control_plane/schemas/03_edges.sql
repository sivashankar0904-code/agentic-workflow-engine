-- edges: the routes[] of a DAG. Each edge connects two nodes by id
-- (FK-enforced graph integrity) and carries an optional `when` predicate.
-- A route with no `when` (all three condition columns NULL) is unconditional.
-- `position` preserves declaration order within a node, since routes are
-- evaluated first-match-wins.
CREATE TABLE IF NOT EXISTS edges (
    id              BIGSERIAL PRIMARY KEY,
    dag_id          BIGINT NOT NULL REFERENCES dag_registry(id) ON DELETE CASCADE,
    from_node_id    BIGINT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    to_node_id      BIGINT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    position        INTEGER NOT NULL,
    when_field      TEXT,
    when_op         TEXT,
    when_value      TEXT
);

CREATE INDEX IF NOT EXISTS idx_edges_dag_id ON edges (dag_id);
CREATE INDEX IF NOT EXISTS idx_edges_from_node ON edges (from_node_id);
CREATE INDEX IF NOT EXISTS idx_edges_to_node ON edges (to_node_id);
