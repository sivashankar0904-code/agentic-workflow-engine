-- permissions: the service_name.feature_group.feature.action catalog.
-- Stored with structured columns (not the flattened string) so grants can
-- target any level of the hierarchy via a real query, not string-prefix
-- parsing. The dotted name a client sends/receives is reconstructed by the
-- permissions_dotted view below, not stored redundantly.
CREATE TABLE IF NOT EXISTS permissions (
    id            BIGSERIAL PRIMARY KEY,
    service_id    BIGINT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    feature_group TEXT NOT NULL,  -- e.g. 'dag_registry', 'user_management'
    feature       TEXT NOT NULL,  -- e.g. 'dag', 'user', 'role'
    action        TEXT NOT NULL,  -- e.g. 'create', 'read', 'read_list', 'update', 'patch', 'delete'
    description   TEXT,
    UNIQUE (service_id, feature_group, feature, action)
);

-- Control Plane's own permission catalog, seeded at bootstrap.
INSERT INTO permissions (service_id, feature_group, feature, action, description)
SELECT s.id, p.feature_group, p.feature, p.action, p.description
FROM services s, (VALUES
    ('dag_registry',    'dag',  'create',    'Upload/create a DAG'),
    ('dag_registry',    'dag',  'read',      'Read a single DAG'),
    ('dag_registry',    'dag',  'read_list', 'List DAGs'),
    ('dag_registry',    'dag',  'update',    'Replace a DAG (full upload)'),
    ('dag_registry',    'dag',  'patch',     'Activate/deactivate a DAG, or edit its role grants'),
    ('dag_registry',    'dag',  'delete',    'Delete a DAG'),
    ('user_management',  'user', 'create',    'Create a user'),
    ('user_management',  'user', 'read_list', 'List users'),
    ('user_management',  'user', 'update',    'Change a user''s role or active state'),
    ('user_management',  'role', 'create',    'Create/edit a role and its permissions')
) AS p(feature_group, feature, action, description)
WHERE s.key = 'control_plane'
ON CONFLICT (service_id, feature_group, feature, action) DO NOTHING;

-- permissions_dotted: the service_name.feature_group.feature.action string
-- clients actually send/receive, computed at read time from the structured
-- columns above.
CREATE OR REPLACE VIEW permissions_dotted AS
SELECT p.*, s.key || '.' || p.feature_group || '.' || p.feature || '.' || p.action AS name
FROM permissions p
JOIN services s ON s.id = p.service_id;
