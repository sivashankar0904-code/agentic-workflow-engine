package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"orchestrator/internal/engine"
	"orchestrator/internal/server/middleware"
)

// New builds the Gin engine with CORS, a health check, and read-only
// introspection of the flows currently pulled from the Control Plane.
func New(registry *engine.Registry) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "orchestrator"})
	})

	r.GET("/flows", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"flows": registry.Names()})
	})

	return r
}
