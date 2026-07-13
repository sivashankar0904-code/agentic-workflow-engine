CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role_id       BIGINT NOT NULL REFERENCES roles(id),
    is_active     BOOLEAN NOT NULL DEFAULT true, -- matches dag_registry.active's convention
    email         TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Bootstrap default admin (username 'admin', password 'admin'). Password can
-- only be changed afterward via POST /me/password (self-service, requires
-- the current password) or, if locked out, the CLI-only
-- `controlplane reset-admin-password` break-glass subcommand — never a
-- second run of this seed.
INSERT INTO users (username, password_hash, role_id, is_active)
SELECT 'admin', '$2a$10$1kY0FlZ6YGfK/CYMTGm6Z.GyB0iYbVvAV2iiLy55YsYtcv6zN7ol6', r.id, true
FROM roles r WHERE r.name = 'admin'
ON CONFLICT (username) DO NOTHING;
