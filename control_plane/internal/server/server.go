package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"controlplane/internal/dag"
	"controlplane/internal/server/handlers"
	"controlplane/internal/server/middleware"
)

// New builds the Gin engine with CORS, a health check, and the DAG registry API.
func New(store *dag.Store) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "control-plane"})
	})

	d := handlers.NewDAG(store)
	r.GET("/dags", d.List)               // ?active=true -> active-filtered names
	r.GET("/dags/:name", d.Get)          // -> YAML
	r.POST("/dags/:name", d.Upload)      // <- YAML body
	r.DELETE("/dags/:name", d.Delete)
	r.POST("/dags/:name/activate", d.Activate)
	r.POST("/dags/:name/deactivate", d.Deactivate)

	return r
}
