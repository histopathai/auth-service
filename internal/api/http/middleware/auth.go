package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/internal/domain/model"

	"github.com/histopathai/auth-service/internal/service"
)

// AuthMiddleeware provides authentication middleware
type AuthMiddleware struct {
	authService    service.AuthService
	sessionService *service.SessionService
}

// NewAuthMiddleware creates a new AuthMiddleware instance
func NewAuthMiddleware(authService service.AuthService, sessionService *service.SessionService) *AuthMiddleware {
	return &AuthMiddleware{
		authService:    authService,
		sessionService: sessionService,
	}
}

// RequireAuth middleware that requires a valid JWT token
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "missing_authorization_header",
				"message": "Authorization header is required."})
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token_format",
				"message": "Authorization header must be Bearer <token>."})
			c.Abort()
			return
		}
		token := tokenParts[1]

		// Verify token and get user

		user, err := m.authService.VerifyToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token",
				"message": "Token verification failed",
			})
			c.Abort()
			return
		}

		// Store user information in context
		c.Set("user", user)
		c.Set("userID", user.UserID)
		c.Next()
	}
}

// ReuireRole middleware that requires a specific user role
func (m *AuthMiddleware) RequireRole(roles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "user_not_found",
				"message": "User not found in context.",
			})
			c.Abort()
			return
		}

		u, ok := user.(*model.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "invalid_user_context",
				"message": "Invalid user context",
			})
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range roles {
			if u.Role == role {
				hasRole = true
				break
			}
		}
		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"erorr":   "insufficient_permissions",
				"message": "You do not have permission to access this resource.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ReuireStatus middleware that requires a specific user status
func (m *AuthMiddleware) RequireStatus(statuses ...model.UserStatus) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "user_not_found",
				"message": "User not found in context.",
			})
			c.Abort()
			return
		}

		u, ok := user.(*model.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "invalid_user_context",
				"message": "Invalid user context",
			})
			c.Abort()
			return
		}

		hasStatus := false
		for _, status := range statuses {
			if u.Status == status {
				hasStatus = true
				break
			}
		}
		if !hasStatus {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "account_status_invalid",
				"message": "Your account status doesn't allow this operation.",
			})
			c.Abort()
			return
		}

		c.Next()
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

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Next()
			return
		}

		token := tokenParts[1]

		user, err := m.authService.VerifyToken(c.Request.Context(), token)
		if err != nil {
			c.Set("user", user)
			c.Set("user_id", user.UserID)
		}

		c.Next()
	}
}

func (m *AuthMiddleware) RequireSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Cookie'den session_id oku
		sessionID, err := c.Cookie("session_id")
		if err != nil || sessionID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "missing_session",
				"message": "Session cookie is required"})
			c.Abort()
			return
		}

		// Session validate et
		session, err := m.sessionService.ValidateSession(c.Request.Context(), sessionID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_session",
				"message": "Session is invalid or expired"})
			c.Abort()
			return
		}

		// User bilgisini context'e ekle
		user, err := m.authService.GetUserByUserID(c.Request.Context(), session.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "user_not_found",
				"message": "User not found"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("user_id", user.UserID)
		c.Set("session_id", sessionID)
		c.Next()
	}
}

func (m *AuthMiddleware) RequireAuthOrSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Önce session cookie'ye bak
		sessionID, err := c.Cookie("session_id")
		if err == nil && sessionID != "" {
			// Session varsa validate et
			session, err := m.sessionService.ValidateSession(c.Request.Context(), sessionID)
			if err == nil {
				user, err := m.authService.GetUserByUserID(c.Request.Context(), session.UserID)
				if err == nil {
					c.Set("user", user)
					c.Set("user_id", user.UserID)
					c.Set("session_id", sessionID)
					c.Set("auth_method", "session") // Hangi method kullanıldığını işaretle
					c.Next()
					return
				}
			}
		}

		// Session yoksa veya geçersizse, Bearer token'a bak
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				token := tokenParts[1]
				user, err := m.authService.VerifyToken(c.Request.Context(), token)
				if err == nil {
					c.Set("user", user)
					c.Set("user_id", user.UserID)
					c.Set("auth_method", "bearer") // Hangi method kullanıldığını işaretle
					c.Next()
					return
				}
			}
		}

		// İkisi de yoksa veya geçersizse
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Valid session cookie or Bearer token required",
		})
		c.Abort()
	}
}
