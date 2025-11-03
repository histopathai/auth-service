package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingMddileware Logs HTTP requests
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		c.Next() // Process request

		// Log the request details
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// Get user info if available
		var userID string
		if uid, exists := c.Get("uid"); exists {
			userID = uid.(string)
		}

		slog.Info("HTTP Request",
			"method", method,
			"path", path,
			"client_ip", clientIP,
			"user_agent", userAgent,
			"status_code", statusCode,
			"latency", latency,
			"user_id", userID,
		)
	}

}
