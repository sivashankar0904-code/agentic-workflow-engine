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
// required to call it ("" = any authenticated user, no specific grant), the
// handler itself, and whether a registered service (X-Service-Key) may call
// it in place of a logged-in user — true only for the DAG read routes the
// execution engines pull from.
type Entry struct {
	Method          string
	Path            string
	Permission      string
	Handler         gin.HandlerFunc
	AllowServiceKey bool
}

// Routes returns every route this service exposes to authenticated users
// (everything except /login, /health, and the service-to-service /authz +
// /services* routes, which are wired directly in server.go since they use a
// different credential — see server.go for why).
func Routes(d *handlers.DAG, u *handlers.Users, ro *handlers.Roles, sess *handlers.Session) []Entry {
	return []Entry{
		// DAG read routes accept a service key (engines pull these) or a user JWT.
		{Method: "GET", Path: "/dags", Permission: DagReadList, Handler: d.List, AllowServiceKey: true},
		{Method: "GET", Path: "/dags/:id", Permission: DagRead, Handler: d.Get, AllowServiceKey: true},
		// DAG mutations: user JWT only.
		{Method: "POST", Path: "/dags/name/:name", Permission: DagCreate, Handler: d.Upload},
		{Method: "DELETE", Path: "/dags/:id", Permission: DagDelete, Handler: d.Delete},
		{Method: "POST", Path: "/dags/:id/activate", Permission: DagPatch, Handler: d.Activate},
		{Method: "POST", Path: "/dags/:id/deactivate", Permission: DagPatch, Handler: d.Deactivate},
		{Method: "PATCH", Path: "/dags/:id/roles", Permission: DagPatch, Handler: d.SetRoles},
		{Method: "GET", Path: "/users", Permission: UserReadList, Handler: u.List},
		{Method: "POST", Path: "/users", Permission: UserCreate, Handler: u.Create},
		{Method: "PATCH", Path: "/users/:username/role", Permission: UserUpdate, Handler: u.SetRole},
		{Method: "PATCH", Path: "/users/:username/active", Permission: UserUpdate, Handler: u.SetActive},
		{Method: "GET", Path: "/roles", Permission: RoleCreate, Handler: ro.List},
		{Method: "POST", Path: "/roles", Permission: RoleCreate, Handler: ro.Create},
		{Method: "DELETE", Path: "/roles/:name", Permission: RoleCreate, Handler: ro.Delete},
		{Method: "PUT", Path: "/roles/:name/permissions", Permission: RoleCreate, Handler: ro.SetPermissions},
		{Method: "PUT", Path: "/roles/:name/permission-groups", Permission: RoleCreate, Handler: ro.SetPermissionGroup},
		{Method: "GET", Path: "/permissions", Permission: RoleCreate, Handler: ro.ListPermissions},
		{Method: "GET", Path: "/me", Permission: "", Handler: sess.Me},
		{Method: "POST", Path: "/me/password", Permission: "", Handler: sess.ChangePassword},
	}
}
