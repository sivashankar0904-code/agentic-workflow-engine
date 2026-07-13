package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"controlplane/internal/users"
)

// Roles handles role CRUD and both flavors of permission grant: exact leaf
// selection (SetPermissions) and the group/feature bulk-grant convenience
// (SetPermissionGroup). Also lists the permission catalog that drives the
// UI's role<->permission matrix.
type Roles struct {
	store *users.Store
}

// NewRoles returns a Roles handler bound to store.
func NewRoles(store *users.Store) *Roles {
	return &Roles{store: store}
}

// List returns every role with its granted (dotted) permission names.
func (h *Roles) List(c *gin.Context) {
	roles, err := h.store.ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]gin.H, 0, len(roles))
	for _, r := range roles {
		perms, err := h.store.PermissionsForRole(c.Request.Context(), r.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, gin.H{
			"id":          r.ID,
			"name":        r.Name,
			"permissions": perms,
		})
	}
	c.JSON(http.StatusOK, gin.H{"roles": out})
}

// ListPermissions returns the full registered permission catalog, ordered
// by service/feature_group/feature/action — drives the matrix's columns.
func (h *Roles) ListPermissions(c *gin.Context) {
	perms, err := h.store.ListPermissions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]gin.H, 0, len(perms))
	for _, p := range perms {
		out = append(out, gin.H{
			"id":           p.ID,
			"serviceId":    p.ServiceID,
			"serviceKey":   p.ServiceKey,
			"featureGroup": p.FeatureGroup,
			"feature":      p.Feature,
			"action":       p.Action,
			"name":         p.DottedName(),
		})
	}
	c.JSON(http.StatusOK, gin.H{"permissions": out})
}

type createRoleRequest struct {
	Name string `json:"name" binding:"required"`
}

// Create creates a new, permission-less role.
func (h *Roles) Create(c *gin.Context) {
	var req createRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	r, err := h.store.CreateRole(c.Request.Context(), req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": r.ID, "name": r.Name})
}

// Delete deletes a role. Refuses if any user still holds it.
func (h *Roles) Delete(c *gin.Context) {
	name := c.Param("name")
	err := h.store.DeleteRole(c.Request.Context(), name)
	switch {
	case errors.Is(err, users.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found: " + name})
	case errors.Is(err, users.ErrRoleInUse):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusOK, gin.H{"status": "deleted", "name": name})
	}
}

type setPermissionsRequest struct {
	PermissionIDs []int64 `json:"permissionIds"`
}

// SetPermissions replaces a role's permission set with exactly the given
// leaf permission IDs — the matrix's per-checkbox "Save changes".
func (h *Roles) SetPermissions(c *gin.Context) {
	var req setPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "permissionIds is required"})
		return
	}
	name := c.Param("name")
	if err := h.store.SetRolePermissions(c.Request.Context(), name, req.PermissionIDs); err != nil {
		if errors.Is(err, users.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found: " + name})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "name": name})
}

type setPermissionGroupRequest struct {
	ServiceID    int64  `json:"serviceId" binding:"required"`
	FeatureGroup string `json:"featureGroup" binding:"required"`
	Feature      string `json:"feature"`
}

// SetPermissionGroup is the group/feature bulk-grant convenience: grants a
// role every permission under a feature_group (or, if Feature is set, just
// that one feature) in one call, without enumerating leaves.
func (h *Roles) SetPermissionGroup(c *gin.Context) {
	var req setPermissionGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "serviceId and featureGroup are required"})
		return
	}
	name := c.Param("name")

	var err error
	if req.Feature != "" {
		err = h.store.GrantFeature(c.Request.Context(), name, req.ServiceID, req.FeatureGroup, req.Feature)
	} else {
		err = h.store.GrantFeatureGroup(c.Request.Context(), name, req.ServiceID, req.FeatureGroup)
	}
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found: " + name})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "name": name})
}
