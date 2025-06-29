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

// Health
// @Summary Service Health Check
// @Description Returns the overall health status of the service
// @Tags Health
// @Produce json
// @Success 200 {object} object{status=string,service=string} "Service is healthy"
// @Router /health [get]
// Health returns the health status of the service
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "auth-service",
	})
}

// Ready
// @Summary Service Readiness Check
// @Description Returns whether the service is ready to accept requests (e.g., database connectivity)
// @Tags Health
// @Produce json
// @Success 200 {object} object{status=string,service=string} "Service is ready"
// @Router /health/ready [get]
// Ready returns the readiness status of the service
func (h *HealthHandler) Ready(c *gin.Context) {
	// Add any readiness checks here (database connectivity, etc.)
	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"service": "auth-service",
	})
}
