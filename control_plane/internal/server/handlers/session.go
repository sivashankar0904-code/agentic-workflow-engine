package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"controlplane/internal/server/middleware"
	"controlplane/internal/users"
)

// Session handles login and the caller's own identity/password.
type Session struct {
	store  *users.Store
	secret string
}

// NewSession returns a Session handler bound to store, signing/verifying
// tokens with secret.
func NewSession(store *users.Store, secret string) *Session {
	return &Session{store: store, secret: secret}
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login verifies credentials and issues a JWT. Returns a generic error for
// both wrong password and unknown user, so a failed attempt never reveals
// whether the username exists.
func (h *Session) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	u, err := h.store.GetByUsername(c.Request.Context(), req.Username)
	if err != nil || !u.IsActive || !users.Compare(u.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := users.Sign(h.secret, u)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"username": u.Username,
			"role":     u.RoleName,
		},
	})
}

// Me returns the caller's identity and permission set, so the UI can
// restore session on reload without re-deriving RBAC client-side.
func (h *Session) Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"username":    middleware.Username(c),
		"role":        middleware.Role(c),
		"permissions": middleware.Permissions(c),
	})
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required"`
}

// ChangePassword is the self-service password change: any logged-in user
// (including admin, day-to-day) may change their own password by proving
// they know the current one. Distinct from the CLI-only
// reset-admin-password break-glass path, which needs no current password.
func (h *Session) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "currentPassword and newPassword are required"})
		return
	}

	username := middleware.Username(c)
	u, err := h.store.GetByUsername(c.Request.Context(), username)
	if err != nil || !users.Compare(u.PasswordHash, req.CurrentPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current password is incorrect"})
		return
	}

	hash, err := users.Hash(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}
	if err := h.store.SetPassword(c.Request.Context(), username, hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
