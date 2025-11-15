package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/pkg/config"
)

func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	corsConfig := cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           12 * 3600, // 12 hours
	}

	return cors.New(corsConfig)
}
