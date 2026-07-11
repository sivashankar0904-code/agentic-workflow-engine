package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"orchestrator/internal/dag"
)

// DAG handles the /config and /dags endpoints, backed by a dag store.
type DAG struct {
	store *dag.Store
}

// NewDAG returns a DAG handler bound to store.
func NewDAG(store *dag.Store) *DAG {
	return &DAG{store: store}
}

// Get returns the DAG named by ?name= as YAML, reassembled from the graph.
func (h *DAG) Get(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing ?name"})
		return
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

// Upload reads a YAML DAG from the request body and saves it under ?name=,
// creating or replacing the stored graph.
func (h *DAG) Upload(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing ?name"})
		return
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

// Delete removes the DAG named by ?name=.
func (h *DAG) Delete(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing ?name"})
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

// List returns the names of all stored DAGs.
func (h *DAG) List(c *gin.Context) {
	names, err := h.store.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"dags": names})
}
