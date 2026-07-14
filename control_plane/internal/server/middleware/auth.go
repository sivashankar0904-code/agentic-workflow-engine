package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"controlplane/internal/users"
)

const (
	ctxKeyUsername    = "auth.username"
	ctxKeyRole        = "auth.role"
	ctxKeyPermissions = "auth.permissions"
	ctxKeyIsService   = "auth.isService"
)

// RequireAuth parses the Authorization: Bearer <token> header, verifies it,
// and loads the caller's username/role/permission set into the Gin context
// for RequirePermission (and handlers) to consult. Only ever attached
// inside the routeperms.Routes registration loop — public routes (/login,
// /health) are registered outside that loop and never see this middleware.
func RequireAuth(secret string, store *users.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		token := strings.TrimPrefix(header, prefix)

		claims, err := users.Verify(secret, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		perms, err := store.PermissionsForRole(c.Request.Context(), claims.Role)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load permissions"})
			return
		}

		c.Set(ctxKeyUsername, claims.Username)
		c.Set(ctxKeyRole, claims.Role)
		c.Set(ctxKeyPermissions, perms)
		c.Next()
	}
}

// RequirePermission 403s unless the context's permission set (loaded by
// RequireAuth, which always runs immediately before it) contains name. A
// blank name means "any authenticated user, no specific grant" and always
// passes. Only ever instantiated inside the routeperms.Routes registration
// loop and the one POST /services registration — never called ad hoc
// elsewhere.
func RequirePermission(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if name == "" || IsService(c) {
			// A blank permission means "any authenticated caller"; a service
			// caller (X-Service-Key) has already been fully authorized by
			// RequireAuthOrServiceKey and carries no per-role permission set.
			c.Next()
			return
		}
		perms, _ := c.Get(ctxKeyPermissions)
		list, _ := perms.([]string)
		for _, p := range list {
			if p == name {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "missing permission: " + name})
	}
}

// RequireServiceKey validates the X-Service-Key header against
// services.api_key, for service-to-service routes (/authz/:userId,
// /services/:key/permissions). These carry no user JWT and no role, so they
// never go through RequireAuth/RequirePermission.
func RequireServiceKey(store *users.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-Service-Key")
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing X-Service-Key"})
			return
		}
		serviceID, serviceKey, err := store.ServiceByAPIKey(c.Request.Context(), key)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid service key"})
			return
		}
		c.Set("service.id", serviceID)
		c.Set("service.key", serviceKey)
		c.Next()
	}
}

// RequireAuthOrServiceKey accepts EITHER a valid X-Service-Key (a registered
// engine pulling active DAGs) OR a user JWT. Used on the DAG read routes so
// the execution engines can authenticate to the Control Plane without a user
// login, per architecture.md. A service caller is marked as a service in the
// context (IsService) and, having no role, is treated as having full read
// access to the active DAGs it lists — the same "sees everything" path an
// admin takes.
func RequireAuthOrServiceKey(secret string, store *users.Store) gin.HandlerFunc {
	auth := RequireAuth(secret, store)
	return func(c *gin.Context) {
		if key := c.GetHeader("X-Service-Key"); key != "" {
			serviceID, serviceKey, err := store.ServiceByAPIKey(c.Request.Context(), key)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid service key"})
				return
			}
			c.Set("service.id", serviceID)
			c.Set("service.key", serviceKey)
			c.Set(ctxKeyIsService, true)
			c.Next()
			return
		}
		auth(c)
	}
}

// IsService reports whether the caller authenticated with a service key
// rather than a user JWT.
func IsService(c *gin.Context) bool {
	v, _ := c.Get(ctxKeyIsService)
	b, _ := v.(bool)
	return b
}

// Username returns the authenticated caller's username, set by RequireAuth.
func Username(c *gin.Context) string {
	v, _ := c.Get(ctxKeyUsername)
	s, _ := v.(string)
	return s
}

// Role returns the authenticated caller's role name, set by RequireAuth.
func Role(c *gin.Context) string {
	v, _ := c.Get(ctxKeyRole)
	s, _ := v.(string)
	return s
}

// Permissions returns the authenticated caller's permission set, set by RequireAuth.
func Permissions(c *gin.Context) []string {
	v, _ := c.Get(ctxKeyPermissions)
	list, _ := v.([]string)
	return list
}
