// Package routeperms is the single source of truth for both "which routes
// exist" and "what permission each one needs." server.go registers routes
// by iterating Routes(...) once — method/path/permission/handler are never
// written a second time in a separate r.GET(...) call, which would let a
// route and its permission drift apart.
package routeperms

import (
	"github.com/gin-gonic/gin"

	"controlplane/internal/server/handlers"
)

// Dotted permission constants for control_plane's own catalog (seeded in
// control_plane/schemas/05_permissions.sql). Exported so server.go can
// reuse RoleCreate for the one route (POST /services) that needs a
// permission check but, being service-registration rather than a
// DAG/user/role CRUD action, doesn't belong in Routes itself.
const (
	DagRead      = "control_plane.dag_registry.dag.read"
	DagReadList  = "control_plane.dag_registry.dag.read_list"
	DagCreate    = "control_plane.dag_registry.dag.create"
	DagDelete    = "control_plane.dag_registry.dag.delete"
	DagPatch     = "control_plane.dag_registry.dag.patch"
	UserReadList = "control_plane.user_management.user.read_list"
	UserCreate   = "control_plane.user_management.user.create"
	UserUpdate   = "control_plane.user_management.user.update"
	RoleCreate   = "control_plane.user_management.role.create"
)

// Entry is one registered route: method, Gin path pattern, the permission
// required to call it ("" = any authenticated user, no specific grant), and
// the handler itself.
type Entry struct {
	Method     string
	Path       string
	Permission string
	Handler    gin.HandlerFunc
}

// Routes returns every route this service exposes to authenticated users
// (everything except /login, /health, and the service-to-service /authz +
// /services* routes, which are wired directly in server.go since they use a
// different credential — see server.go for why).
func Routes(d *handlers.DAG, u *handlers.Users, ro *handlers.Roles, sess *handlers.Session) []Entry {
	return []Entry{
		{"GET", "/dags", DagReadList, d.List},
		{"GET", "/dags/:name", DagRead, d.Get},
		{"POST", "/dags/:name", DagCreate, d.Upload},
		{"DELETE", "/dags/:name", DagDelete, d.Delete},
		{"POST", "/dags/:name/activate", DagPatch, d.Activate},
		{"POST", "/dags/:name/deactivate", DagPatch, d.Deactivate},
		{"PATCH", "/dags/:name/roles", DagPatch, d.SetRoles},
		{"GET", "/users", UserReadList, u.List},
		{"POST", "/users", UserCreate, u.Create},
		{"PATCH", "/users/:username/role", UserUpdate, u.SetRole},
		{"PATCH", "/users/:username/active", UserUpdate, u.SetActive},
		{"GET", "/roles", RoleCreate, ro.List},
		{"POST", "/roles", RoleCreate, ro.Create},
		{"DELETE", "/roles/:name", RoleCreate, ro.Delete},
		{"PUT", "/roles/:name/permissions", RoleCreate, ro.SetPermissions},
		{"PUT", "/roles/:name/permission-groups", RoleCreate, ro.SetPermissionGroup},
		{"GET", "/permissions", RoleCreate, ro.ListPermissions},
		{"GET", "/me", "", sess.Me},
		{"POST", "/me/password", "", sess.ChangePassword},
	}
}
