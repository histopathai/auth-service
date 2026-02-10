package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/internal/domain/model"
	"github.com/histopathai/auth-service/internal/service"
	"github.com/histopathai/auth-service/pkg/config"
)

type AuthMiddleware struct {
	authService    service.AuthService
	sessionService *service.SessionService
	config         *config.Config
	logger         *slog.Logger
}

func NewAuthMiddleware(
	authService service.AuthService,
	sessionService *service.SessionService,
	config *config.Config,
	logger *slog.Logger,
) *AuthMiddleware {
	return &AuthMiddleware{
		authService:    authService,
		sessionService: sessionService,
		config:         config,
		logger:         logger,
	}
}

// setUserContext sets user information in the context
func (m *AuthMiddleware) setUserContext(c *gin.Context, user *model.User, authMethod string) {
	c.Set("user", user)
	c.Set("user_id", user.UserID)
	c.Set("auth_method", authMethod)
}

// respondUnauthorized sends a standardized unauthorized response
func respondUnauthorized(c *gin.Context, errorCode, message string, details map[string]interface{}) {
	response := gin.H{
		"error":   errorCode,
		"message": message,
	}
	if details != nil {
		response["details"] = details
	}
	c.JSON(http.StatusUnauthorized, response)
	c.Abort()
}

// respondForbidden sends a standardized forbidden response
func respondForbidden(c *gin.Context, errorCode, message string) {
	c.JSON(http.StatusForbidden, gin.H{
		"error":   errorCode,
		"message": message,
	})
	c.Abort()
}

// respondInternalError sends a standardized internal error response
func respondInternalError(c *gin.Context, errorCode, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"error":   errorCode,
		"message": message,
	})
	c.Abort()
}

// getUserFromContext retrieves and validates user from context
func getUserFromContext(c *gin.Context) (*model.User, bool) {
	userInterface, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	user, ok := userInterface.(*model.User)
	if !ok {
		return nil, false
	}

	return user, true
}

// authenticateWithBearer attempts to authenticate using Bearer token
func (m *AuthMiddleware) authenticateWithBearer(c *gin.Context, authHeader string) (*model.User, error) {
	tokenParts := strings.SplitN(authHeader, " ", 2)
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return nil, gin.Error{Meta: "invalid_token_format"}
	}

	return m.authService.VerifyToken(c.Request.Context(), tokenParts[1])
}

// authenticateWithSession attempts to authenticate using session cookie
func (m *AuthMiddleware) authenticateWithSession(c *gin.Context) (*model.User, string, error) {
	sessionID, err := c.Cookie(m.config.Cookie.Name)
	if err != nil || sessionID == "" {
		return nil, "", err
	}

	session, err := m.sessionService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		return nil, "", err
	}

	user, err := m.authService.GetUserByUserID(c.Request.Context(), session.UserID)
	if err != nil {
		return nil, "", err
	}

	return user, sessionID, nil
}

// RequireAuth middleware that requires a valid JWT token
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			respondUnauthorized(c, "missing_authorization_header", "Authorization header is required", nil)
			return
		}

		user, err := m.authenticateWithBearer(c, authHeader)
		if err != nil {
			respondUnauthorized(c, "invalid_token", "Token verification failed", nil)
			return
		}

		m.setUserContext(c, user, "bearer")
		c.Next()
	}
}

// RequireSession middleware that requires a valid session cookie
func (m *AuthMiddleware) RequireSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, sessionID, err := m.authenticateWithSession(c)
		if err != nil {
			m.logger.Warn("Session authentication failed",
				"error", err,
				"path", c.Request.URL.Path,
				"ip", c.ClientIP(),
			)
			respondUnauthorized(c, "invalid_session", "Session is invalid or expired", nil)
			return
		}

		m.setUserContext(c, user, "session")
		c.Set("session_id", sessionID)
		c.Next()
	}
}

// RequireAuthOrSession middleware that accepts either Bearer token or session cookie
func (m *AuthMiddleware) RequireAuthOrSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		details := make(map[string]interface{})

		// Try session authentication first
		if user, sessionID, err := m.authenticateWithSession(c); err == nil {
			m.setUserContext(c, user, "session")
			c.Set("session_id", sessionID)
			c.Next()
			return
		} else if err != nil {
			details["session_error"] = err.Error()
		}

		// Try bearer token authentication
		if authHeader := c.GetHeader("Authorization"); authHeader != "" {
			if user, err := m.authenticateWithBearer(c, authHeader); err == nil {
				m.setUserContext(c, user, "bearer")
				c.Next()
				return
			} else if err != nil {
				details["bearer_error"] = err.Error()
			}
		} else {
			details["bearer_error"] = "Authorization header not provided"
		}

		respondUnauthorized(c, "unauthorized", "Valid session cookie or Bearer token required", details)
	}
}

// OptionalAuth middleware that extracts user if token is present but does not require it
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		user, err := m.authenticateWithBearer(c, authHeader)
		if err == nil && user != nil {
			m.setUserContext(c, user, "bearer")
		}

		c.Next()
	}
}

// RequireRole middleware that requires a specific user role
func (m *AuthMiddleware) RequireRole(roles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := getUserFromContext(c)
		if !ok {
			respondInternalError(c, "invalid_user_context", "User context is invalid")
			return
		}

		for _, role := range roles {
			if user.Role == role {
				c.Next()
				return
			}
		}

		respondForbidden(c, "insufficient_permissions", "You do not have permission to access this resource")
	}
}

// RequireStatus middleware that requires a specific user status
func (m *AuthMiddleware) RequireStatus(statuses ...model.UserStatus) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := getUserFromContext(c)
		if !ok {
			respondInternalError(c, "invalid_user_context", "User context is invalid")
			return
		}

		for _, status := range statuses {
			if user.Status == status {
				c.Next()
				return
			}
		}

		respondForbidden(c, "account_status_invalid", "Your account status doesn't allow this operation")
	}
}
