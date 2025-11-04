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
	authService service.AuthService
}

// NewAuthMiddleware creates a new AuthMiddleware instance
func NewAuthMiddleware(authService service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
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
		c.Set("uid", user.UID)
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
			c.Set("uid", user.UID)
		}

		c.Next()
	}
}
