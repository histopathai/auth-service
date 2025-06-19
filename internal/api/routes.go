package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes configures the authentication and user management API routes.
func SetupAuthRoutes(router *gin.Engine, handler *AuthHandler) {
	authGroup := router.Group("/auth")

	// 🔐 Public/admin tools
	authGroup.POST("/users", handler.AdminAuthMiddleware(), handler.HandleUserCreation)
	authGroup.PUT("/users/:uid/role", handler.AdminAuthMiddleware(), handler.HandleUpdateUserRole)
	authGroup.PUT("/users/:uid/activate", handler.AdminAuthMiddleware(), handler.ActivateUser)
	authGroup.PUT("/users/:uid/deactivate", handler.AdminAuthMiddleware(), handler.DeactivateUser)

	// 🔒 Protected user self-info
	authGroup.GET("/me", handler.AuthMiddleware(), func(c *gin.Context) {
		uidVal, uidExists := c.Get("userId")
		emailVal, emailExists := c.Get("userEmail")
		roleVal, roleExists := c.Get("userRole")

		if !uidExists || !emailExists || !roleExists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized or incomplete context"})
			return
		}

		uid, ok1 := uidVal.(string)
		email, ok2 := emailVal.(string)
		role, ok3 := roleVal.(string)

		if !ok1 || !ok2 || !ok3 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Context value type mismatch"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"uid":     uid,
			"email":   email,
			"role":    role,
			"message": "Authenticated successfully",
		})
	})
}
