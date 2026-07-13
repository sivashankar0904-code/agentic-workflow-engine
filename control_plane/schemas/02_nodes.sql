-- nodes: the graph vertices of a DAG. Unique by name within a DAG.
-- `entry` marks the (at most one) node where ingress messages start.
-- `tools` is the optional list of logical tool names an agent node may call,
-- resolved to concrete MCP tools by agentic_orchestrator; service_orchestrator
-- ignores it.
CREATE TABLE IF NOT EXISTS nodes (
    id     BIGSERIAL PRIMARY KEY,
    dag_id BIGINT NOT NULL REFERENCES dag_registry(id) ON DELETE CASCADE,
    name   TEXT NOT NULL,
    topic  TEXT NOT NULL,
    host   TEXT NOT NULL,
    port   INTEGER NOT NULL,
    entry  BOOLEAN NOT NULL DEFAULT false,
    tools  TEXT[] NOT NULL DEFAULT '{}',
    UNIQUE (dag_id, name)
);

CREATE INDEX IF NOT EXISTS idx_nodes_dag_id ON nodes (dag_id);
