package users

import (
	"errors"
	"time"
)

// ErrNotFound is returned when a named user, role, or permission does not exist.
var ErrNotFound = errors.New("not found")

// ErrLastAdmin is returned when an operation would leave the system with no
// active admin (demoting or disabling the last active admin, or deleting the
// admin role while a user still holds it).
var ErrLastAdmin = errors.New("cannot remove the last active admin")

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
