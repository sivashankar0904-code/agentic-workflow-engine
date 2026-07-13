package users

import (
	"errors"
	"time"
)

// ErrNotFound is returned when a named user, role, or permission does not exist.
var ErrNotFound = errors.New("not found")

// ErrLastAdmin is returned when an operation would leave the system with no
// active admin (demoting or disabling the last active admin).
var ErrLastAdmin = errors.New("cannot remove the last active admin")

// ErrRoleInUse is returned when deleting a role that one or more users still hold.
var ErrRoleInUse = errors.New("role is still assigned to one or more users")

// User is a Control Plane account. Every user has exactly one role;
// permissions are derived from that role via role_permissions.
type User struct {
	ID           int64
	Username     string
	PasswordHash string
	RoleID       int64
	RoleName     string
	IsActive     bool
	Email        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
