package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	BaseHandler
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		BaseHandler: BaseHandler{logger: logger, response: &ResponseHelper{}},
	}
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

	message := gin.H{
		"status":  "healthy",
		"service": "auth-service",
	}
	h.response.Success(c, http.StatusOK, message)
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

	message := gin.H{
		"status":  "ready",
		"service": "auth-service",
	}
	h.response.Success(c, http.StatusOK, message)
}
