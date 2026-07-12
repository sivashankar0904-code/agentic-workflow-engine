package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"orchestrator/internal/dag"
	"orchestrator/internal/server/handlers"
	"orchestrator/internal/server/middleware"
)

// New builds the Gin engine with CORS, a health check, and the DAG API.
func New(store *dag.Store) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "orchestrator"})
	})

	d := handlers.NewDAG(store)
	r.GET("/config", d.Get)     // ?name= -> YAML
	r.POST("/config", d.Upload) // ?name= <- YAML body
	r.DELETE("/config", d.Delete)
	r.GET("/dags", d.List)

	return r
}
