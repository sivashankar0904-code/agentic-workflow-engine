package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store persists users, roles, permissions, and services. Mirrors
// dag.Store's shape: one Store type wrapping a pool, one method per
// operation, transactional replace for many-to-many rewrites.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore returns a Store backed by pool.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// ─── Users ──────────────────────────────────────────────────────────────

// GetByUsername returns the user named username, with its role name joined in.
func (s *Store) GetByUsername(ctx context.Context, username string) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.username, u.password_hash, u.role_id, r.name, u.is_active,
		       coalesce(u.email, ''), u.created_at, u.updated_at
		FROM users u JOIN roles r ON r.id = u.role_id
		WHERE u.username = $1`, username).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.RoleID, &u.RoleName, &u.IsActive,
			&u.Email, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	return u, nil
}

// List returns every user, with role names joined in, ordered by username.
func (s *Store) List(ctx context.Context) ([]User, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.id, u.username, u.password_hash, u.role_id, r.name, u.is_active,
		       coalesce(u.email, ''), u.created_at, u.updated_at
		FROM users u JOIN roles r ON r.id = u.role_id
		ORDER BY u.username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.RoleID, &u.RoleName,
			&u.IsActive, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// Create inserts a new user with roleName and passwordHash, for onboarding
// (POST /users): admin-chosen username/role/temporary password.
func (s *Store) Create(ctx context.Context, username, passwordHash, roleName, email string) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (username, password_hash, role_id, email)
		SELECT $1, $2, r.id, nullif($4, '')
		FROM roles r WHERE r.name = $3
		RETURNING id, username, password_hash, role_id, is_active, coalesce(email, ''), created_at, updated_at`,
		username, passwordHash, roleName, email).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.RoleID, &u.IsActive, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	u.RoleName = roleName
	return u, nil
}

// SetRole changes username's role to roleName. Refuses to demote the last
// active admin.
func (s *Store) SetRole(ctx context.Context, username, roleName string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	if err := guardLastAdmin(ctx, tx, username, roleName, nil); err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `
		UPDATE users SET role_id = (SELECT id FROM roles WHERE name = $2), updated_at = now()
		WHERE username = $1`, username, roleName)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return tx.Commit(ctx)
}

// SetActive enables or disables username. Refuses to disable the last active admin.
func (s *Store) SetActive(ctx context.Context, username string, active bool) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	if err := guardLastAdmin(ctx, tx, username, "", &active); err != nil {
		return err
	}

	tag, err := tx.Exec(ctx,
		`UPDATE users SET is_active = $2, updated_at = now() WHERE username = $1`, username, active)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return tx.Commit(ctx)
}

// SetPassword sets username's password to the bcrypt hash of newPasswordHash.
func (s *Store) SetPassword(ctx context.Context, username, passwordHash string) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE users SET password_hash = $2, updated_at = now() WHERE username = $1`,
		username, passwordHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// guardLastAdmin rejects a role change or active-flag change that would
// leave zero active admins. newRoleName is "" when only active is changing;
// newActive is nil when only role is changing. Must run inside tx so the
// count and the update are atomic.
func guardLastAdmin(ctx context.Context, tx pgx.Tx, username, newRoleName string, newActive *bool) error {
	var currentRole string
	var currentActive bool
	err := tx.QueryRow(ctx, `
		SELECT r.name, u.is_active FROM users u JOIN roles r ON r.id = u.role_id
		WHERE u.username = $1`, username).Scan(&currentRole, &currentActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	// Only admin accounts can trigger the guard, and only if the change
	// would actually take them out of the active-admin set.
	if currentRole != "admin" || !currentActive {
		return nil
	}
	willStayAdmin := newRoleName == "" || newRoleName == "admin"
	willStayActive := newActive == nil || *newActive
	if willStayAdmin && willStayActive {
		return nil
	}

	var activeAdmins int
	err = tx.QueryRow(ctx, `
		SELECT count(*) FROM users u JOIN roles r ON r.id = u.role_id
		WHERE r.name = 'admin' AND u.is_active = true`).Scan(&activeAdmins)
	if err != nil {
		return err
	}
	if activeAdmins <= 1 {
		return ErrLastAdmin
	}
	return nil
}

// ─── Roles & permissions ───────────────────────────────────────────────

// ListRoles returns every role.
func (s *Store) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name FROM roles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Role
	for rows.Next() {
		var r Role
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// CreateRole creates a new, permission-less role.
func (s *Store) CreateRole(ctx context.Context, name string) (Role, error) {
	var r Role
	err := s.pool.QueryRow(ctx,
		`INSERT INTO roles (name) VALUES ($1) RETURNING id, name`, name).
		Scan(&r.ID, &r.Name)
	if err != nil {
		return Role{}, fmt.Errorf("create role: %w", err)
	}
	return r, nil
}

// DeleteRole deletes the role named name. Refuses if any user still holds it.
func (s *Store) DeleteRole(ctx context.Context, name string) error {
	var inUse int
	err := s.pool.QueryRow(ctx, `
		SELECT count(*) FROM users u JOIN roles r ON r.id = u.role_id WHERE r.name = $1`,
		name).Scan(&inUse)
	if err != nil {
		return err
	}
	if inUse > 0 {
		return fmt.Errorf("%w: %d user(s) still have this role", ErrLastAdmin, inUse)
	}

	tag, err := s.pool.Exec(ctx, `DELETE FROM roles WHERE name = $1`, name)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListPermissions returns every registered permission across all services,
// with dotted names, ordered for stable grouping (service, feature_group,
// feature, action).
func (s *Store) ListPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.id, p.service_id, s.key, p.feature_group, p.feature, p.action
		FROM permissions p JOIN services s ON s.id = p.service_id
		ORDER BY s.key, p.feature_group, p.feature, p.action`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Permission
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.ID, &p.ServiceID, &p.ServiceKey, &p.FeatureGroup, &p.Feature, &p.Action); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// PermissionsForRole returns the dotted permission names granted to the
// named role — the query backing GET /authz/:userId.
func (s *Store) PermissionsForRole(ctx context.Context, roleName string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT s.key || '.' || p.feature_group || '.' || p.feature || '.' || p.action
		FROM role_permissions rp
		JOIN roles r ON r.id = rp.role_id
		JOIN permissions p ON p.id = rp.permission_id
		JOIN services s ON s.id = p.service_id
		WHERE r.name = $1
		ORDER BY 1`, roleName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// SetRolePermissions replaces roleName's permission set with exactly
// permissionIDs — the matrix's per-checkbox "Save changes".
func (s *Store) SetRolePermissions(ctx context.Context, roleName string, permissionIDs []int64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	var roleID int64
	if err := tx.QueryRow(ctx, `SELECT id FROM roles WHERE name = $1`, roleName).Scan(&roleID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, roleID); err != nil {
		return fmt.Errorf("clear role_permissions: %w", err)
	}
	for _, pid := range permissionIDs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`,
			roleID, pid); err != nil {
			return fmt.Errorf("grant permission %d: %w", pid, err)
		}
	}
	return tx.Commit(ctx)
}

// GrantFeatureGroup grants roleName every permission under featureGroup for
// the given service — the group-level bulk-grant convenience.
func (s *Store) GrantFeatureGroup(ctx context.Context, roleName string, serviceID int64, featureGroup string) error {
	return s.grantMatching(ctx, roleName, `service_id = $2 AND feature_group = $3`, serviceID, featureGroup)
}

// GrantFeature grants roleName every permission under featureGroup/feature
// for the given service.
func (s *Store) GrantFeature(ctx context.Context, roleName string, serviceID int64, featureGroup, feature string) error {
	return s.grantMatching(ctx, roleName,
		`service_id = $2 AND feature_group = $3 AND feature = $4`, serviceID, featureGroup, feature)
}

func (s *Store) grantMatching(ctx context.Context, roleName, where string, args ...interface{}) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	var roleID int64
	if err := tx.QueryRow(ctx, `SELECT id FROM roles WHERE name = $1`, roleName).Scan(&roleID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	queryArgs := append([]interface{}{roleID}, args...)
	_, err = tx.Exec(ctx, `
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT $1, id FROM permissions WHERE `+where+`
		ON CONFLICT DO NOTHING`, queryArgs...)
	if err != nil {
		return fmt.Errorf("grant matching permissions: %w", err)
	}
	return tx.Commit(ctx)
}

// ─── Services ───────────────────────────────────────────────────────────

// RegisterService creates a new service registration, returning its
// generated api_key.
func (s *Store) RegisterService(ctx context.Context, key, name, apiKey string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO services (key, name, api_key) VALUES ($1, $2, $3)`,
		key, name, apiKey)
	if err != nil {
		return fmt.Errorf("register service: %w", err)
	}
	return nil
}

// ServiceByAPIKey looks up a service by its api_key, for RequireServiceKey.
func (s *Store) ServiceByAPIKey(ctx context.Context, apiKey string) (int64, string, error) {
	var id int64
	var key string
	err := s.pool.QueryRow(ctx, `SELECT id, key FROM services WHERE api_key = $1`, apiKey).
		Scan(&id, &key)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, "", ErrNotFound
	}
	if err != nil {
		return 0, "", err
	}
	return id, key, nil
}

// RegisterPermission idempotently declares one permission for serviceID —
// called by a service on every boot to (re)assert its catalog.
func (s *Store) RegisterPermission(ctx context.Context, serviceID int64, featureGroup, feature, action, description string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO permissions (service_id, feature_group, feature, action, description)
		VALUES ($1, $2, $3, $4, nullif($5, ''))
		ON CONFLICT (service_id, feature_group, feature, action) DO NOTHING`,
		serviceID, featureGroup, feature, action, description)
	if err != nil {
		return fmt.Errorf("register permission: %w", err)
	}
	return nil
}

// ─── DAG resource-level grants ─────────────────────────────────────────

// DagRolesFor returns the role names granted access to dagID.
func (s *Store) DagRolesFor(ctx context.Context, dagID int64) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT r.name FROM dag_roles dr JOIN roles r ON r.id = dr.role_id
		WHERE dr.dag_id = $1 ORDER BY r.name`, dagID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// SetDagRoles replaces dagID's role grants with exactly roleIDs.
func (s *Store) SetDagRoles(ctx context.Context, dagID int64, roleIDs []int64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	if _, err := tx.Exec(ctx, `DELETE FROM dag_roles WHERE dag_id = $1`, dagID); err != nil {
		return fmt.Errorf("clear dag_roles: %w", err)
	}
	for _, rid := range roleIDs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO dag_roles (dag_id, role_id) VALUES ($1, $2)`, dagID, rid); err != nil {
			return fmt.Errorf("grant role %d on dag %d: %w", rid, dagID, err)
		}
	}
	return tx.Commit(ctx)
}
