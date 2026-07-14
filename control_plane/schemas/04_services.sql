-- services: registry of every service allowed to declare permissions and
-- query /authz. The Control Plane is the single authorization authority for
-- the whole system (see docs/architecture.md); other services (e.g.
-- service_orchestrator, agentic_orchestrator) register here on their own
-- boot via POST /services + POST /services/:key/permissions.
CREATE TABLE IF NOT EXISTS services (
    id         BIGSERIAL PRIMARY KEY,
    key        TEXT NOT NULL UNIQUE,   -- stable slug, e.g. 'control_plane', 'service_orchestrator'
    name       TEXT NOT NULL,          -- display name for the UI
    api_key    TEXT NOT NULL,          -- shared secret the service presents when registering/querying authz
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- The Control Plane registers itself as service 1. Seeded directly here
-- (rather than self-registering over HTTP at boot) since it's the one
-- service guaranteed to exist before its own API is up.
INSERT INTO services (key, name, api_key)
VALUES ('control_plane', 'Control Plane', 'dev-control-plane-api-key-change-me')
ON CONFLICT (key) DO NOTHING;

-- The service_orchestrator engine, seeded with a known dev api_key it
-- presents via X-Service-Key to pull active DAGs. Production must rotate
-- this (and register other engines via POST /services instead).
INSERT INTO services (key, name, api_key)
VALUES ('service_orchestrator', 'Service Orchestrator', 'dev-orchestrator-api-key-change-me')
ON CONFLICT (key) DO NOTHING;
