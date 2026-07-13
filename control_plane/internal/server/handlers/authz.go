package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"controlplane/internal/users"
)

// Authz is the cross-service authorization oracle: any registered service's
// middleware calls GET /authz/:userId to learn what a given caller may do,
// then enforces locally. Also handles service self-registration and
// permission-catalog declaration.
type Authz struct {
	store *users.Store
}

// NewAuthz returns an Authz handler bound to store.
func NewAuthz(store *users.Store) *Authz {
	return &Authz{store: store}
}

// Get returns the permission list for the user named by :userId (a
// username, despite the path segment name matching the architecture's
// GET /authz/:userId contract). Requires a valid service API key.
func (h *Authz) Get(c *gin.Context) {
	username := c.Param("userId")
	u, err := h.store.GetByUsername(c.Request.Context(), username)
	if errors.Is(err, users.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found: " + username})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	perms, err := h.store.PermissionsForRole(c.Request.Context(), u.RoleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userId":      u.Username,
		"role":        u.RoleName,
		"permissions": perms,
	})
}

type registerServiceRequest struct {
	Key  string `json:"key" binding:"required"`
	Name string `json:"name" binding:"required"`
}

// RegisterService registers a new service, returning its generated api_key.
// Requires an admin-equivalent human caller (wired outside routeperms.Routes
// in server.go, since this is service-registry, not DAG/user/role CRUD).
func (h *Authz) RegisterService(c *gin.Context) {
	var req registerServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key and name are required"})
		return
	}

	apiKey, err := generateAPIKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate api key"})
		return
	}

	if err := h.store.RegisterService(c.Request.Context(), req.Key, req.Name, apiKey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"key": req.Key, "name": req.Name, "apiKey": apiKey})
}

type registerPermissionRequest struct {
	FeatureGroup string `json:"featureGroup" binding:"required"`
	Feature      string `json:"feature" binding:"required"`
	Action       string `json:"action" binding:"required"`
	Description  string `json:"description"`
}

// RegisterPermissions lets a registered service (authenticated by its own
// api_key via RequireServiceKey) declare/update its permission catalog on
// boot — idempotent, safe to call every startup.
func (h *Authz) RegisterPermissions(c *gin.Context) {
	serviceIDVal, _ := c.Get("service.id")
	serviceID, _ := serviceIDVal.(int64)

	var reqs []registerPermissionRequest
	if err := c.ShouldBindJSON(&reqs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "expected an array of {featureGroup, feature, action}"})
		return
	}

	for _, r := range reqs {
		if err := h.store.RegisterPermission(c.Request.Context(), serviceID, r.FeatureGroup, r.Feature, r.Action, r.Description); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "count": len(reqs)})
}

func generateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
