package handlers

import (
	"errors"
	"io"
	"net/http"
	"strconv"

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
//
// Upload is name-addressed (a DAG is a named workflow); everything else
// addresses a DAG by its stable id, which List returns alongside the name.
type DAG struct {
	store     *dag.Store
	userStore *users.Store
}

// NewDAG returns a DAG handler bound to store.
func NewDAG(store *dag.Store, userStore *users.Store) *DAG {
	return &DAG{store: store, userStore: userStore}
}

// seesAllDags reports whether the caller bypasses dag_roles resource-level
// checks and sees every DAG. The bootstrap admin role does; so does a
// service caller (a registered engine pulling active DAGs via X-Service-Key),
// which has no role and full read access to what it lists. Every other role
// is subject to dag_roles.
func seesAllDags(c *gin.Context) bool {
	return middleware.IsService(c) || middleware.Role(c) == "admin"
}

// parseID reads the :id path param, writing a 400 and returning ok=false if
// it isn't a valid integer.
func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dag id"})
		return 0, false
	}
	return id, true
}

// List returns the stored DAGs as {id, name, active}. ?active=true restricts
// the result to active DAGs. Admins see every DAG; other roles see only DAGs
// where their role has a dag_roles grant.
func (h *DAG) List(c *gin.Context) {
	activeOnly := c.Query("active") == "true"

	var entries []dag.ListEntry
	var err error
	if seesAllDags(c) {
		entries, err = h.store.List(c.Request.Context(), activeOnly)
	} else {
		entries, err = h.store.ListForRole(c.Request.Context(), middleware.Role(c), activeOnly)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"dags": entries})
}

// Get returns the DAG with :id as YAML, reassembled from the graph.
// Non-admin roles without a dag_roles grant on this DAG get 404 rather than
// 403, so a guessed id doesn't leak existence.
func (h *DAG) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.checkVisible(c, id); err != nil {
		writeDagError(c, err)
		return
	}

	d, err := h.store.Get(c.Request.Context(), id)
	if err != nil {
		writeDagError(c, err)
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
// creating or replacing the stored graph, returning the DAG's id. New DAGs
// start inactive and, if created (not replacing an existing DAG), have no
// dag_roles grant yet — visible only to roles with the bare dag.create
// permission (i.e. effectively admin/editor) until an admin assigns
// dag_roles.
//
// Replacing an existing DAG additionally requires the caller's role to
// already be visible on it (or be admin) — dag.create alone doesn't let an
// editor overwrite a DAG scoped to a different role.
func (h *DAG) Upload(c *gin.Context) {
	name := c.Param("name")

	if !seesAllDags(c) {
		existing, err := h.store.List(c.Request.Context(), false)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, e := range existing {
			if e.Name != name {
				continue
			}
			visible, err := h.store.VisibleToRole(c.Request.Context(), e.ID, middleware.Role(c))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if !visible {
				c.JSON(http.StatusNotFound, gin.H{"error": "dag not found: " + name})
				return
			}
			break
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

	id, err := h.store.Save(c.Request.Context(), name, d)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to store DAG: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "saved", "id": id, "name": name})
}

// Delete removes the DAG with :id.
func (h *DAG) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.checkVisible(c, id); err != nil {
		writeDagError(c, err)
		return
	}

	if err := h.store.Delete(c.Request.Context(), id); err != nil {
		writeDagError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted", "id": id})
}

// Activate marks the DAG with :id active.
func (h *DAG) Activate(c *gin.Context) {
	h.setActive(c, true)
}

// Deactivate marks the DAG with :id inactive without deleting it.
func (h *DAG) Deactivate(c *gin.Context) {
	h.setActive(c, false)
}

func (h *DAG) setActive(c *gin.Context, active bool) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.checkVisible(c, id); err != nil {
		writeDagError(c, err)
		return
	}

	if err := h.store.SetActive(c.Request.Context(), id, active); err != nil {
		writeDagError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "id": id, "active": active})
}

type setDagRolesRequest struct {
	RoleIDs []int64 `json:"roleIds"`
}

// SetRoles replaces the DAG's dag_roles with exactly the given role IDs —
// backs RolesAccessPanel's "Save changes" for the DAG<->role grant matrix.
func (h *DAG) SetRoles(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.checkVisible(c, id); err != nil {
		writeDagError(c, err)
		return
	}

	var req setDagRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "roleIds is required"})
		return
	}

	if err := h.userStore.SetDagRoles(c.Request.Context(), id, req.RoleIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "id": id})
}

// checkVisible re-checks that the DAG with id exists and, for non-admins,
// that the caller's role has a dag_roles grant on it. A missing DAG or an
// ungranted DAG both surface as ErrNotFound so existence isn't leaked.
func (h *DAG) checkVisible(c *gin.Context, id int64) error {
	if _, err := h.store.Meta(c.Request.Context(), id); err != nil {
		return err
	}
	if seesAllDags(c) {
		return nil
	}
	visible, err := h.store.VisibleToRole(c.Request.Context(), id, middleware.Role(c))
	if err != nil {
		return err
	}
	if !visible {
		return dag.ErrNotFound
	}
	return nil
}

func writeDagError(c *gin.Context, err error) {
	if errors.Is(err, dag.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "dag not found"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
