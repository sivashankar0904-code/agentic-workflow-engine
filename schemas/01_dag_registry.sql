-- dag_registry: one row per workflow DAG. Multiple DAGs coexist and are
-- addressed by name. Nodes and edges reference a DAG via dag_id.
CREATE TABLE IF NOT EXISTS dag_registry (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
