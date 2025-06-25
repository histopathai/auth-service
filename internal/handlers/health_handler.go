package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health returns the health status of the service
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "auth-service",
	})
}

// Ready returns the readiness status of the service
func (h *HealthHandler) Ready(c *gin.Context) {
	// Add any readiness checks here (database connectivity, etc.)
	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"service": "auth-service",
	})
}
