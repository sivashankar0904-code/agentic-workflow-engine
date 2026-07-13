package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"controlplane/internal/dag"
	"controlplane/internal/server/handlers"
	"controlplane/internal/server/middleware"
	"controlplane/internal/server/routeperms"
	"controlplane/internal/users"
)

// New builds the Gin engine with CORS, a health check, session/login, the
// DAG registry API, user/role management, and the cross-service authz
// oracle — all gated by RBAC except /health and /login.
func New(dagStore *dag.Store, userStore *users.Store, jwtSecret string) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "control-plane"})
	})

	d := handlers.NewDAG(dagStore, userStore)
	u := handlers.NewUsers(userStore)
	ro := handlers.NewRoles(userStore)
	sess := handlers.NewSession(userStore, jwtSecret)
	az := handlers.NewAuthz(userStore)

	// The two truly public routes: no token required, so they're never
	// folded into routeperms.Routes (which would force every entry to
	// carry a third "requires auth?" bit alongside Permission).
	r.POST("/login", sess.Login)

	requireAuth := middleware.RequireAuth(jwtSecret, userStore)
	for _, e := range routeperms.Routes(d, u, ro, sess) {
		r.Handle(e.Method, e.Path, requireAuth, middleware.RequirePermission(e.Permission), e.Handler)
	}

	// Service-to-service: a separate credential (api_key, no role), so
	// these are never part of routeperms.Routes — a different auth
	// dimension, not a special-cased row in the same table.
	requireServiceKey := middleware.RequireServiceKey(userStore)
	r.GET("/authz/:userId", requireServiceKey, az.Get)
	r.POST("/services/:key/permissions", requireServiceKey, az.RegisterPermissions)

	// Service registration needs an admin-equivalent human caller. It
	// reuses routeperms.RoleCreate but isn't itself a routeperms.Routes
	// entry (it's service-registry, not DAG/user/role CRUD), so it's
	// wired explicitly.
	r.POST("/services", requireAuth, middleware.RequirePermission(routeperms.RoleCreate), az.RegisterService)

	return r
}
