package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterHealth wires GET /api/health, returning the same {"status":"UP"} shape as the Kotlin backend.
func RegisterHealth(r gin.IRouter) {
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})
}
