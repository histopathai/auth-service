package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/internal/domain/model"
)

// SelfOrAdminOnly middleware access only if user is accesing their own resource or is an admin
func SelfOrAdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":    "user_not_found",
				"meessage": "User not found in context",
			})
		}

		u, ok := user.(*model.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "invalid_user_context",
				"message": "User context is invalid",
			})
			c.Abort()
			return
		}

		// get the UID from the URL parameter
		targetUID := c.Param("uid")

		//Allow if user is admin or accesing their own resource
		if u.Role == model.RoleAdmin || u.UID == targetUID {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "access_denied",
			"message": "You can only access your own resources",
		})
		c.Abort()
	}

}
