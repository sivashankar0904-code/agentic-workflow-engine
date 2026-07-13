package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"controlplane/internal/dag"
	"controlplane/internal/server/middleware"
	"controlplane/internal/users"
)

// DAG handles the /dags endpoints, backed by a dag store. Action-level
// gating (can this role do dag.create/dag.delete/dag.patch at all) is
// handled entirely by the RequirePermission instance routeperms.Routes
// wires in front of each handler, before the handler runs. Each handler
// here still re-checks the dag_roles resource-level grant for the target
// DAG — is *this specific* DAG visible to the caller's role.
type DAG struct {
	store     *dag.Store
	userStore *users.Store
}

// NewDAG returns a DAG handler bound to store.
func NewDAG(store *dag.Store, userStore *users.Store) *DAG {
	return &DAG{store: store, userStore: userStore}
}

// isAdmin reports whether the caller's role holds every permission that
// exists — the bootstrap admin role, and any role granted the whole
// catalog. Cheap approximation used here: role name "admin" bypasses
// dag_roles resource-level checks, matching the bootstrap seed. A role
// other than "admin" is always subject to dag_roles.
func isAdmin(c *gin.Context) bool {
	return middleware.Role(c) == "admin"
}

// List returns the names of stored DAGs. ?active=true restricts the result
// to DAGs currently in the active lifecycle state. Admins see every DAG;
// other roles see only DAGs where their role has a dag_roles grant.
func (h *DAG) List(c *gin.Context) {
	activeOnly := c.Query("active") == "true"

	var names []string
	var err error
	if isAdmin(c) {
		names, err = h.store.List(c.Request.Context(), activeOnly)
	} else {
		names, err = h.store.ListForRole(c.Request.Context(), middleware.Role(c), activeOnly)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"dags": names})
}

// Get returns the DAG named by :name as YAML, reassembled from the graph.
// Non-admin roles without a dag_roles grant on this DAG get 404 rather than
// 403, so a guessed name doesn't leak existence.
func (h *DAG) Get(c *gin.Context) {
	name := c.Param("name")

	if !isAdmin(c) {
		visible, err := h.store.VisibleToRole(c.Request.Context(), name, middleware.Role(c))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !visible {
			c.JSON(http.StatusNotFound, gin.H{"error": "dag not found: " + name})
			return
		}
	}

	d, err := h.store.Get(c.Request.Context(), name)
	if errors.Is(err, dag.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "dag not found: " + name})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out, err := dag.ToYAML(d)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/x-yaml", out)
}

// Upload reads a YAML DAG from the request body and saves it under :name,
// creating or replacing the stored graph. New DAGs start inactive and, if
// created (not replacing an existing DAG), have no dag_roles grant yet —
// visible only to roles with the bare dag.create permission (i.e.
// effectively admin/editor) until an admin assigns dag_roles.
//
// Replacing an existing DAG additionally requires the caller's role to
// already be visible on it (or be admin) — dag.create alone doesn't let an
// editor overwrite a DAG scoped to a different role.
func (h *DAG) Upload(c *gin.Context) {
	name := c.Param("name")

	if !isAdmin(c) {
		existed, err := h.store.IDByName(c.Request.Context(), name)
		if err != nil && !errors.Is(err, dag.ErrNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if existed != 0 {
			visible, err := h.store.VisibleToRole(c.Request.Context(), name, middleware.Role(c))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if !visible {
				c.JSON(http.StatusNotFound, gin.H{"error": "dag not found: " + name})
				return
			}
		}
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	d, err := dag.FromYAML(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid YAML: " + err.Error()})
		return
	}

	if err := h.store.Save(c.Request.Context(), name, d); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to store DAG: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "saved", "name": name})
}

// Delete removes the DAG named by :name.
func (h *DAG) Delete(c *gin.Context) {
	name := c.Param("name")

	if err := h.checkResourceVisible(c, name); err != nil {
		writeDagError(c, name, err)
		return
	}

	err := h.store.Delete(c.Request.Context(), name)
	if errors.Is(err, dag.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "dag not found: " + name})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted", "name": name})
}

// Activate marks the named DAG active, making it visible to
// GET /dags?active=true and eligible for execution engines to build.
func (h *DAG) Activate(c *gin.Context) {
	h.setActive(c, true)
}

// Deactivate marks the named DAG inactive without deleting it.
func (h *DAG) Deactivate(c *gin.Context) {
	h.setActive(c, false)
}

func (h *DAG) setActive(c *gin.Context, active bool) {
	name := c.Param("name")

	if err := h.checkResourceVisible(c, name); err != nil {
		writeDagError(c, name, err)
		return
	}

	err := h.store.SetActive(c.Request.Context(), name, active)
	if errors.Is(err, dag.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "dag not found: " + name})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "name": name, "active": active})
}

type setDagRolesRequest struct {
	RoleIDs []int64 `json:"roleIds"`
}

// SetRoles replaces the DAG's dag_roles with exactly the given role IDs —
// backs RolesAccessPanel's "Save changes" for the DAG<->role grant matrix.
func (h *DAG) SetRoles(c *gin.Context) {
	name := c.Param("name")

	if err := h.checkResourceVisible(c, name); err != nil {
		writeDagError(c, name, err)
		return
	}

	var req setDagRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "roleIds is required"})
		return
	}

	dagID, err := h.store.IDByName(c.Request.Context(), name)
	if errors.Is(err, dag.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "dag not found: " + name})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.userStore.SetDagRoles(c.Request.Context(), dagID, req.RoleIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "name": name})
}

// checkResourceVisible re-checks the dag_roles resource-level grant for
// name against the caller's role, for mutation endpoints. Admins bypass it.
func (h *DAG) checkResourceVisible(c *gin.Context, name string) error {
	if isAdmin(c) {
		return nil
	}
	visible, err := h.store.VisibleToRole(c.Request.Context(), name, middleware.Role(c))
	if err != nil {
		return err
	}
	if !visible {
		return dag.ErrNotFound
	}
	return nil
}

func writeDagError(c *gin.Context, name string, err error) {
	if errors.Is(err, dag.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "dag not found: " + name})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
