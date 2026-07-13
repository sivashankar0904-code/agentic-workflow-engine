-- dag_roles: which roles may see/act on which *specific* DAG. A
-- resource-level grant, distinct from the action-level role_permissions in
-- 06_roles.sql: role_permissions says "editors can do dag.create at all";
-- dag_roles says "this particular DAG is visible to ops and editor, not
-- viewer." A DAG with no rows here means no non-admin role can see it
-- (admin bypasses both this and role_permissions).
CREATE TABLE IF NOT EXISTS dag_roles (
    dag_id  BIGINT NOT NULL REFERENCES dag_registry(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (dag_id, role_id)
);
