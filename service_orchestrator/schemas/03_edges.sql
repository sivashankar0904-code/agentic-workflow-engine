-- edges: the routing rules (graph edges) of a DAG. Each edge connects two nodes
-- by id (FK-enforced graph integrity) and carries the routing condition.
CREATE TABLE IF NOT EXISTS edges (
    id                 BIGSERIAL PRIMARY KEY,
    dag_id             BIGINT NOT NULL REFERENCES dag_registry(id) ON DELETE CASCADE,
    from_node_id       BIGINT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    to_node_id         BIGINT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    condition_field    TEXT NOT NULL,
    condition_contains TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_edges_dag_id ON edges (dag_id);
CREATE INDEX IF NOT EXISTS idx_edges_from_node ON edges (from_node_id);
CREATE INDEX IF NOT EXISTS idx_edges_to_node ON edges (to_node_id);
