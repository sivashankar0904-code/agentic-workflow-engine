-- nodes: the services (graph vertices) of a DAG. Unique by name within a DAG.
CREATE TABLE IF NOT EXISTS nodes (
    id     BIGSERIAL PRIMARY KEY,
    dag_id BIGINT NOT NULL REFERENCES dag_registry(id) ON DELETE CASCADE,
    name   TEXT NOT NULL,
    host   TEXT NOT NULL,
    port   INTEGER NOT NULL,
    topic  TEXT NOT NULL,
    UNIQUE (dag_id, name)
);

CREATE INDEX IF NOT EXISTS idx_nodes_dag_id ON nodes (dag_id);
