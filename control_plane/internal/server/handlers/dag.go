package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"controlplane/internal/dag"
)

// DAG handles the /dags endpoints, backed by a dag store.
type DAG struct {
	store *dag.Store
}

// NewDAG returns a DAG handler bound to store.
func NewDAG(store *dag.Store) *DAG {
	return &DAG{store: store}
}

// List returns the names of stored DAGs. ?active=true restricts the result to
// DAGs currently in the active lifecycle state.
func (h *DAG) List(c *gin.Context) {
	activeOnly := c.Query("active") == "true"

	names, err := h.store.List(c.Request.Context(), activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"dags": names})
}

// Get returns the DAG named by :name as YAML, reassembled from the graph.
func (h *DAG) Get(c *gin.Context) {
	name := c.Param("name")

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
// creating or replacing the stored graph. New DAGs start inactive.
func (h *DAG) Upload(c *gin.Context) {
	name := c.Param("name")

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
