package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	dtoRequest "github.com/histopathai/auth-service/internal/api/http/dto/request"
	dtoResponse "github.com/histopathai/auth-service/internal/api/http/dto/response"
	"github.com/histopathai/auth-service/internal/domain/model"
	"github.com/histopathai/auth-service/internal/service"
	"github.com/histopathai/auth-service/internal/shared/errors"
	"github.com/histopathai/auth-service/pkg/config"
)

func (h *SessionHandler) setSessionCookie(c *gin.Context, sessionID string, expiresAt time.Time) {
	cookieCfg := h.config.Cookie
	maxAge := int(time.Until(expiresAt).Seconds())

	c.SetSameSite(h.getSameSiteMode(cookieCfg.SameSite))
	c.SetCookie(
		cookieCfg.Name,     // name
		sessionID,          // value
		maxAge,             // maxAge
		"/",                // path
		cookieCfg.Domain,   // domain
		cookieCfg.Secure,   // secure (HTTPS only)
		cookieCfg.HTTPOnly, // httpOnly
	)

	h.logger.Debug("Session cookie set",
		"environment", h.config.Server.Environment,
		"secure", cookieCfg.Secure,
		"sameSite", cookieCfg.SameSite,
		"domain", cookieCfg.Domain,
	)
}

func (h *SessionHandler) clearSessionCookie(c *gin.Context) {
	cookieCfg := h.config.Cookie

	c.SetSameSite(h.getSameSiteMode(cookieCfg.SameSite))
	c.SetCookie(
		cookieCfg.Name,
		"",
		-1, // Delete immediately
		"/",
		cookieCfg.Domain,
		cookieCfg.Secure,
		cookieCfg.HTTPOnly,
	)
}

func (h *SessionHandler) getSameSiteMode(mode string) http.SameSite {
	switch mode {
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	case "Lax":
		fallthrough
	default:
		return http.SameSiteLaxMode
	}
}

type SessionHandler struct {
	sessionService *service.SessionService
	authService    *service.AuthService
	config         *config.Config
	BaseHandler
}

func NewSessionHandler(
	sessionService *service.SessionService,
	authService *service.AuthService,
	config *config.Config,
	logger *slog.Logger,
) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
		authService:    authService,
		config:         config,
		BaseHandler:    BaseHandler{logger: logger, response: &ResponseHelper{}},
	}
}

// CreateSession
// @Summary Create Session
// @Description Create a new session with authentication token
// @Tags Session
// @Accept json
// @Produce json
// @Param payload body request.CreateSessionRequest true "Authentication token"
// @Success 201 {object} response.CreateSessionResponse "Session created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 401 {object} response.ErrorResponse "Invalid token"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /sessions [post]
func (h *SessionHandler) CreateSession(c *gin.Context) {
	var req dtoRequest.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.NewValidationError("Invalid request payload", nil))
		return
	}

	// Verify token and get user
	user, err := h.authService.VerifyToken(c.Request.Context(), req.Token)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Create session
	sessionID, err := h.sessionService.CreateSession(c.Request.Context(), user.UserID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Get created session details
	session, err := h.sessionService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Set cookie with environment-aware configuration
	h.setSessionCookie(c, sessionID, session.ExpiresAt)

	response := dtoResponse.CreateSessionResponse{
		ExpiresAt: session.ExpiresAt,
		Message:   "Session created successfully",
		Session:   mapToSessionResponse(session),
	}

	h.response.Success(c, http.StatusCreated, response)
}

// ListMySessions
// @Summary List My Sessions
// @Description Get list of authenticated user's active sessions
// @Tags Session
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.SessionListResponse "Sessions retrieved successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /sessions [get]
func (h *SessionHandler) ListMySessions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.handleError(c, errors.NewUnauthorizedError("User not authenticated"))
		return
	}

	sessions, err := h.sessionService.GetUserSessionStats(c.Request.Context(), userID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Convert to response format
	sessionList := sessions["sessions"].([]map[string]interface{})
	responseSessions := make([]dtoResponse.SessionResponse, 0, len(sessionList))

	for _, s := range sessionList {
		responseSessions = append(responseSessions, dtoResponse.SessionResponse{
			SessionID:    s["session_id"].(string),
			CreatedAt:    s["created_at"].(time.Time),
			ExpiresAt:    s["expires_at"].(time.Time),
			LastUsedAt:   s["last_used"].(time.Time),
			RequestCount: s["request_count"].(int64),
		})
	}

	response := dtoResponse.SessionListResponse{
		ActiveSessions: sessions["active_sessions"].(int),
		Sessions:       responseSessions,
	}

	h.response.Success(c, http.StatusOK, response)
}

// GetMySessionStats
// @Summary Get My Session Statistics
// @Description Get detailed statistics of authenticated user's sessions
// @Tags Session
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.SessionStatsResponse "Session statistics retrieved successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /sessions/stats [get]
func (h *SessionHandler) GetMySessionStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.handleError(c, errors.NewUnauthorizedError("User not authenticated"))
		return
	}

	stats, err := h.sessionService.GetUserSessionStats(c.Request.Context(), userID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Convert to detailed response format
	sessionList := stats["sessions"].([]map[string]interface{})
	detailedSessions := make([]dtoResponse.SessionDetailedStats, 0, len(sessionList))
	var totalRequests int64

	for _, s := range sessionList {
		expiresAt := s["expires_at"].(time.Time)
		timeLeft := time.Until(expiresAt)
		requestCount := s["request_count"].(int64)
		totalRequests += requestCount

		detailedSessions = append(detailedSessions, dtoResponse.SessionDetailedStats{
			SessionID:    s["session_id"].(string),
			CreatedAt:    s["created_at"].(time.Time),
			ExpiresAt:    expiresAt,
			LastUsedAt:   s["last_used"].(time.Time),
			RequestCount: requestCount,
			TimeLeft:     timeLeft.Round(time.Second).String(),
		})
	}

	response := dtoResponse.SessionStatsResponse{
		ActiveSessions: stats["active_sessions"].(int),
		TotalRequests:  totalRequests,
		Sessions:       detailedSessions,
		Summary: map[string]interface{}{
			"average_requests_per_session": float64(totalRequests) / float64(len(detailedSessions)),
		},
	}

	h.response.Success(c, http.StatusOK, response)
}

// ExtendSession
// @Summary Extend Session
// @Description Extend the expiration time of a session
// @Tags Session
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param session_id path string true "Session ID"
// @Success 200 {object} response.ExtendSessionResponse "Session extended successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid session ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "Session not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /sessions/{session_id}/extend [post]
func (h *SessionHandler) ExtendSession(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		h.handleError(c, errors.NewValidationError("Missing session ID", nil))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.handleError(c, errors.NewUnauthorizedError("User not authenticated"))
		return
	}

	// Verify session belongs to user
	session, err := h.sessionService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	if session.UserID != userID.(string) {
		h.handleError(c, errors.NewForbiddenError("You can only extend your own sessions"))
		return
	}

	// Extend session
	if err := h.sessionService.ExtendSession(c.Request.Context(), sessionID); err != nil {
		h.handleError(c, err)
		return
	}

	// Get updated session
	updatedSession, _ := h.sessionService.ValidateSession(c.Request.Context(), sessionID)

	response := dtoResponse.ExtendSessionResponse{
		SessionID: sessionID,
		ExpiresAt: updatedSession.ExpiresAt,
		Message:   "Session extended successfully",
	}

	h.response.Success(c, http.StatusOK, response)
}

// RevokeSession
// @Summary Revoke Session
// @Description Revoke/delete a specific session
// @Tags Session
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param session_id path string true "Session ID"
// @Success 200 {object} response.RevokeSessionResponse "Session revoked successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid session ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "Session not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /sessions/{session_id} [delete]
func (h *SessionHandler) RevokeSession(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		h.handleError(c, errors.NewValidationError("Missing session ID", nil))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.handleError(c, errors.NewUnauthorizedError("User not authenticated"))
		return
	}

	session, err := h.sessionService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	if session.UserID != userID.(string) {
		h.handleError(c, errors.NewForbiddenError("You can only revoke your own sessions"))
		return
	}

	if err := h.sessionService.RevokeSession(c.Request.Context(), sessionID); err != nil {
		h.handleError(c, err)
		return
	}

	// Clear cookie if it's the current session
	if currentSessionID, _ := c.Cookie(h.config.Cookie.Name); currentSessionID == sessionID {
		h.clearSessionCookie(c)
	}

	response := dtoResponse.RevokeSessionResponse{
		Message: "Session revoked successfully",
	}

	h.response.Success(c, http.StatusOK, response)
}

// RevokeAllMySessions
// @Summary Revoke All My Sessions
// @Description Revoke all sessions belonging to the authenticated user
// @Tags Session
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.RevokeAllSessionsResponse "All sessions revoked successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /sessions/revoke-all [post]
func (h *SessionHandler) RevokeAllMySessions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.handleError(c, errors.NewUnauthorizedError("User not authenticated"))
		return
	}

	// Get session count before revoking
	count, _ := h.sessionService.GetActiveSessionCount(c.Request.Context(), userID.(string))

	// Revoke all sessions
	if err := h.sessionService.RevokeAllUserSessions(c.Request.Context(), userID.(string)); err != nil {
		h.handleError(c, err)
		return
	}

	response := dtoResponse.RevokeAllSessionsResponse{
		Message:         "All sessions revoked successfully",
		RevokedSessions: count,
	}

	h.response.Success(c, http.StatusOK, response)
}

// Admin Endpoints

// ListUserSessions (Admin)
// @Summary List User Sessions (Admin)
// @Description Get list of sessions for a specific user (Admin only)
// @Tags Admin - Sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path string true "User UserID"
// @Success 200 {object} response.SessionListResponse "Sessions retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid UserID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/users/{user_id}/sessions [get]
func (h *SessionHandler) ListUserSessions(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		h.handleError(c, errors.NewValidationError("Missing UserID", nil))
		return
	}

	sessions, err := h.sessionService.GetUserSessionStats(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Convert to response format
	sessionList := sessions["sessions"].([]map[string]interface{})
	responseSessions := make([]dtoResponse.SessionResponse, 0, len(sessionList))

	for _, s := range sessionList {
		responseSessions = append(responseSessions, dtoResponse.SessionResponse{
			SessionID:    s["session_id"].(string),
			CreatedAt:    s["created_at"].(time.Time),
			ExpiresAt:    s["expires_at"].(time.Time),
			LastUsedAt:   s["last_used"].(time.Time),
			RequestCount: s["request_count"].(int64),
		})
	}

	response := dtoResponse.SessionListResponse{
		ActiveSessions: sessions["active_sessions"].(int),
		Sessions:       responseSessions,
	}

	h.response.Success(c, http.StatusOK, response)
}

// RevokeUserSession (Admin)
// @Summary Revoke User Session (Admin)
// @Description Revoke a specific session of any user (Admin only)
// @Tags Admin - Sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param session_id path string true "Session ID"
// @Success 200 {object} response.RevokeSessionResponse "Session revoked successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid session ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "Session not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/sessions/{session_id} [delete]
func (h *SessionHandler) RevokeUserSession(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		h.handleError(c, errors.NewValidationError("Missing session ID", nil))
		return
	}

	// Admin can revoke any session without validation
	if err := h.sessionService.RevokeSession(c.Request.Context(), sessionID); err != nil {
		h.handleError(c, err)
		return
	}

	response := dtoResponse.RevokeSessionResponse{
		Message: "Session revoked successfully",
	}

	h.response.Success(c, http.StatusOK, response)
}

// RevokeAllUserSessions (Admin)
// @Summary Revoke All User Sessions (Admin)
// @Description Revoke all sessions of a specific user (Admin only)
// @Tags Admin - Sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path string true "User UserID"
// @Success 200 {object} response.RevokeAllSessionsResponse "All user sessions revoked successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid UserID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/users/{user_id}/sessions [delete]
func (h *SessionHandler) RevokeAllUserSessions(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		h.handleError(c, errors.NewValidationError("Missing UserID", nil))
		return
	}

	// Get session count before revoking
	count, _ := h.sessionService.GetActiveSessionCount(c.Request.Context(), userID)

	// Revoke all user sessions
	if err := h.sessionService.RevokeAllUserSessions(c.Request.Context(), userID); err != nil {
		h.handleError(c, err)
		return
	}

	response := dtoResponse.RevokeAllSessionsResponse{
		Message:         "All user sessions revoked successfully",
		RevokedSessions: count,
	}

	h.response.Success(c, http.StatusOK, response)
}

// Helper function to map session model to response
func mapToSessionResponse(session *model.Session) dtoResponse.SessionResponse {
	return dtoResponse.SessionResponse{
		SessionID:    session.SessionID,
		UserID:       session.UserID,
		CreatedAt:    session.CreatedAt,
		ExpiresAt:    session.ExpiresAt,
		LastUsedAt:   session.LastUsedAt,
		RequestCount: session.RequestCount,
		Metadata:     session.Metadata,
	}
}
