-- roles: not a fixed enum. Admins can create/rename/delete roles and edit
-- each role's permission set at runtime via PUT /roles/:name/permissions
-- (exact leaf grants) or PUT /roles/:name/permission-groups (bulk grant of
-- an entire feature_group/feature).
CREATE TABLE IF NOT EXISTS roles (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- role_permissions always stores individual leaf permission rows. Granting
-- "a whole feature_group" is a write-time convenience (insert every
-- matching permission row, as the seeds below do) rather than a separate
-- runtime concept, so the read path (/authz, RequirePermission) stays a
-- flat membership check with no prefix logic.
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

INSERT INTO roles (name) VALUES ('admin'), ('editor'), ('ops'), ('viewer')
ON CONFLICT (name) DO NOTHING;

-- admin: every permission that exists at bootstrap time (whole catalog —
-- the broadest possible grant, one level above even feature_group).
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

-- editor: every action on the dag_registry.dag feature except delete —
-- a feature-level grant expressed as an explicit action list.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'editor' AND p.feature_group = 'dag_registry' AND p.feature = 'dag'
  AND p.action IN ('create','read','read_list','update')
ON CONFLICT DO NOTHING;

-- ops: view + lifecycle toggle, no authoring.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'ops' AND p.feature_group = 'dag_registry' AND p.feature = 'dag'
  AND p.action IN ('read','read_list','patch')
ON CONFLICT DO NOTHING;

-- viewer: whole dag_registry feature_group, read-only actions only.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'viewer' AND p.feature_group = 'dag_registry'
  AND p.action IN ('read','read_list')
ON CONFLICT DO NOTHING;
