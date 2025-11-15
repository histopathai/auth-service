package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/pkg/config"
)

func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Sadece configured origin'i kabul et (dev'de extension bypass eder zaten)
		if origin != cfg.AllowedOrigin {
			c.Next()
			return
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", cfg.AllowedOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Cookie")
		c.Writer.Header().Set("Access-Control-Max-Age", "3600")

		// Preflight request handling
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
