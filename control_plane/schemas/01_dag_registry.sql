-- dag_registry: one row per workflow DAG. Multiple DAGs coexist and are
-- addressed by name. Nodes and edges reference a DAG via dag_id.
-- `active` gates whether execution engines are served this DAG at all
-- (GET /dags?active=true); RBAC/ownership metadata can extend this table later.
CREATE TABLE IF NOT EXISTS dag_registry (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    active     BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
