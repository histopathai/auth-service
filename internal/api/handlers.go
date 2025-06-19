package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/histopathai/auth-service/internal/service"
	"github.com/histopathai/auth-service/pkg/models"
)

// Handler struct holds dependencies
type AuthHandler struct {
	AuthService service.AuthService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		AuthService: authService,
	}
}

// AuthMiddleware checks for a valid Firebase ID Token and attaches user info to context
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		if idToken == authHeader { // No "Bearer " prefix found
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}

		token, err := h.AuthService.VerifyIDToken(c.Request.Context(), idToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Add UID and the other information to the context
		c.Set("userId", token.UID)
		if emails, ok := token.Firebase.Identities["email"].([]interface{}); ok && len(emails) > 0 {
			if email, ok := emails[0].(string); ok {
				c.Set("userEmail", email)
			}
		}

		// check role
		if role, ok := token.Claims["role"].(string); ok {
			c.Set("userRole", role)
		} else {
			c.Set("userRole", "unknown")
		}

		c.Next()

	}
}

// AdminAuthMiddleware (admin-only)
func (h *AuthHandler) AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists || fmt.Sprintf("%v", role) != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return
		}

		c.Next()
	}
}

// HandleUserCreation (admin only)
func (h *AuthHandler) HandleUserCreation(c *gin.Context) {
	var req models.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate role before sending to service (e.g., must be patolog, admin or viewer)

	if !models.ValidRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role specified"})
		return
	}

	user, err := h.AuthService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user": user})
}

// HandleUpdateUserRole (admin only)
func (h *AuthHandler) HandleUpdateUserRole(c *gin.Context) {
	uid := c.Param("uid")

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.AuthService.UpdateUserRole(c.Request.Context(), uid, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User role updated successfully"})
}

// DeactivateUser (admin only)
func (h *AuthHandler) DeactivateUser(c *gin.Context) {
	uid := c.Param("uid")

	err := h.AuthService.DeactivateUser(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deactivated successfully"})
}

// ActivateUser (admin only)
func (h *AuthHandler) ActivateUser(c *gin.Context) {
	uid := c.Param("uid")

	err := h.AuthService.ActivateUser(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User activated successfully"})
}
