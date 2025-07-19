// internal/handlers/session_handler.go
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/internal/models"
	"github.com/histopathai/auth-service/internal/service"
)

type SessionHandler struct {
	sessionService *service.ImageSessionService
}

func NewSessionHandler(sessionService *service.ImageSessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

// CreateImageSession
// @Summary Create Image Session
// @Description Creates a short-lived session for image viewing with optimized performance
// @Tags Image Session
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} object{session_id=string,expires_in=int,expires_at=string} "Session created successfully"
// @Failure 401 {object} object{error=string,message=string} "Authentication required"
// @Failure 500 {object} object{error=string,message=string} "Session creation failed"
// @Router /auth/image-session [post]
func (h *SessionHandler) CreateImageSession(c *gin.Context) {
	// Get authenticated user
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "authentication_required",
			"message": "User authentication required",
		})
		return
	}

	u, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "invalid_user_context",
			"message": "Invalid user context",
		})
		return
	}

	// Create session
	sessionID, err := h.sessionService.CreateSession(u.UID, string(u.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "session_creation_failed",
			"message": "Failed to create image session",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"expires_in": 1800, // 30 minutes in seconds
		"message":    "Image session created successfully",
	})
}

// GetSessionStats
// @Summary Get Session Statistics
// @Description Get current user's session statistics and active sessions
// @Tags Image Session
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} object{stats=object} "Session statistics"
// @Failure 401 {object} object{error=string,message=string} "Authentication required"
// @Router /auth/image-session/stats [get]
func (h *SessionHandler) GetSessionStats(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication_required",
		})
		return
	}

	u := user.(*models.User)
	stats := h.sessionService.GetUserSessionStats(u.UID)

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// RevokeSession
// @Summary Revoke Image Session
// @Description Revoke a specific image session
// @Tags Image Session
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param session_id path string true "Session ID to revoke"
// @Success 200 {object} object{message=string} "Session revoked successfully"
// @Failure 400 {object} object{error=string,message=string} "Invalid session ID"
// @Failure 401 {object} object{error=string,message=string} "Authentication required"
// @Router /auth/image-session/{session_id} [delete]
func (h *SessionHandler) RevokeSession(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_session_id",
			"message": "Session ID is required",
		})
		return
	}

	err := h.sessionService.RevokeSession(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "revocation_failed",
			"message": "Failed to revoke session",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Session revoked successfully",
	})
}

// RevokeAllSessions
// @Summary Revoke All User Sessions
// @Description Revoke all image sessions for the current user
// @Tags Image Session
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} object{message=string} "All sessions revoked successfully"
// @Failure 401 {object} object{error=string,message=string} "Authentication required"
// @Router /auth/image-session/revoke-all [post]
func (h *SessionHandler) RevokeAllSessions(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication_required",
		})
		return
	}

	u := user.(*models.User)
	err := h.sessionService.RevokeAllUserSessions(u.UID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "revocation_failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All sessions revoked successfully",
	})
}
