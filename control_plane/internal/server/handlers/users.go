package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"controlplane/internal/users"
)

// Users handles the /users endpoints: admin onboarding and role/active
// management. Onboarding is admin-created only — no self-registration, no
// email flow. A new user's temporary password is admin-chosen and shared
// out-of-band; the user changes it via POST /me/password on first login.
type Users struct {
	store *users.Store
}

// NewUsers returns a Users handler bound to store.
func NewUsers(store *users.Store) *Users {
	return &Users{store: store}
}

// List returns every user.
func (h *Users) List(c *gin.Context) {
	list, err := h.store.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, u := range list {
		out = append(out, userJSON(u))
	}
	c.JSON(http.StatusOK, gin.H{"users": out})
}

type createUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
	Email    string `json:"email"`
}

// Create onboards a new user with an admin-chosen temporary password.
func (h *Users) Create(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username, password, and role are required"})
		return
	}

	hash, err := users.Hash(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	u, err := h.store.Create(c.Request.Context(), req.Username, hash, req.Role, req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, userJSON(u))
}

type setRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// SetRole changes a user's role. Refuses to demote the last active admin.
func (h *Users) SetRole(c *gin.Context) {
	var req setRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role is required"})
		return
	}

	username := c.Param("username")
	err := h.store.SetRole(c.Request.Context(), username, req.Role)
	writeGuardedResult(c, username, err)
}

type setActiveRequest struct {
	IsActive bool `json:"isActive"`
}

// SetActive enables/disables a user. Refuses to disable the last active admin.
func (h *Users) SetActive(c *gin.Context) {
	var req setActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "isActive is required"})
		return
	}

	username := c.Param("username")
	err := h.store.SetActive(c.Request.Context(), username, req.IsActive)
	writeGuardedResult(c, username, err)
}

func writeGuardedResult(c *gin.Context, username string, err error) {
	switch {
	case errors.Is(err, users.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found: " + username})
	case errors.Is(err, users.ErrLastAdmin):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func userJSON(u users.User) gin.H {
	return gin.H{
		"username": u.Username,
		"role":     u.RoleName,
		"isActive": u.IsActive,
		"email":    u.Email,
	}
}
