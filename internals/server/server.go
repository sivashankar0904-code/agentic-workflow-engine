package server

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"orchestrator/internals/dagconfig"
)

// New builds the Gin engine with CORS and the orchestrator's HTTP routes.
func New(store *dagconfig.Store) *gin.Engine {
	r := gin.Default()
	r.Use(cors())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "orchestrator"})
	})

	r.GET("/config", func(c *gin.Context) {
		c.JSON(http.StatusOK, store.Get())
	})

	r.POST("/config", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
			return
		}
		// The DAG dir holds only DAG files; store under the user-provided name.
		// Defaults to the currently active key when unspecified.
		name := c.Query("name")
		if name == "" {
			name = store.ActiveKey()
		}
		if _, err := store.Replace(c.Request.Context(), name, body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to store DAG: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "reloaded", "key": name})
	})

	return r
}

// cors mirrors the permissive CORS policy the UI relies on.
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
