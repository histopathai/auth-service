// internal/routes/routes.go - Düzeltilmiş routes
package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/internal/handlers"
	"github.com/histopathai/auth-service/internal/middleware"
	"github.com/histopathai/auth-service/internal/models"
	"github.com/histopathai/auth-service/internal/proxy"
	"github.com/histopathai/auth-service/internal/service"

	_ "github.com/histopathai/auth-service/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(authService service.AuthService, sessionService *service.ImageSessionService, rateLimiter *middleware.RateLimiter, imgCatalogURL string) *gin.Engine {
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	adminHandler := handlers.NewAdminHandler(authService)
	healthHandler := handlers.NewHealthHandler()
	sessionHandler := handlers.NewSessionHandler(sessionService) // 🆕 Session handler

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Create a new Gin router
	router := gin.New()

	// Global middleware
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.CORSMiddleware())

	// Health check routes (no rate limiting)
	health := router.Group("/health")
	{
		health.GET("/", healthHandler.Health)
		health.GET("/ready", healthHandler.Ready)
	}

	// API routes with rate limiting
	api := router.Group("/api/v1")
	api.Use(rateLimiter.RateLimit())

	// Public Authentication routes
	auth := api.Group("/auth")
	{
		// Registration
		auth.POST("/register", authHandler.Register)

		// Token verification
		auth.POST("/verify", authHandler.VerifyToken)

		// 🆕 Image Session endpoints (protected)
		sessionRoutes := auth.Group("")
		sessionRoutes.Use(authMiddleware.RequireAuth())
		sessionRoutes.Use(authMiddleware.RequireStatus(models.StatusActive))
		{
			sessionRoutes.POST("/image-session", sessionHandler.CreateImageSession)
			sessionRoutes.GET("/image-session/stats", sessionHandler.GetSessionStats)
			sessionRoutes.DELETE("/image-session/:session_id", sessionHandler.RevokeSession)
			sessionRoutes.POST("/image-session/revoke-all", sessionHandler.RevokeAllSessions)
		}
	}

	// Protected routes (require authentication)
	user := api.Group("/user")
	user.Use(authMiddleware.RequireAuth())
	user.Use(authMiddleware.RequireStatus(models.StatusActive))
	{
		user.GET("/profile", authHandler.GetProfile)
		user.PUT("/password", authHandler.ChangePasswordSelf)
		user.DELETE("/account", authHandler.DeleteAccount)
	}

	// Admin routes (require admin role)
	admin := api.Group("/admin")
	admin.Use(authMiddleware.RequireAuth())
	admin.Use(authMiddleware.RequireRole(models.RoleAdmin))
	admin.Use(authMiddleware.RequireStatus(models.StatusActive))
	{
		users := admin.Group("/users")
		{
			users.GET("/", adminHandler.GetAllUsers)
			users.GET("/:uid", adminHandler.GetUser)
			users.POST("/:uid/approve", adminHandler.ApproveUser)
			users.POST("/:uid/suspend", adminHandler.SuspendUser)
			users.PUT("/:uid/password", adminHandler.ChangePasswordForUser)
			users.POST("/:uid/promote", adminHandler.MakeAdmin)
		}
	}

	// 🚀 Optimized Image Catalog Routes (NO RATE LIMITING)
	// Bu routes rate limiting DIŞINDA - çünkü OpenSeadragon 50+ request yapar
	imageCatalogGroup := router.Group("/api/v1/image-catalog")
	{
		// AuthService'i de geçin
		imageCatalogGroup.Any("/*proxyPath", proxy.NewImageCatalogProxy(imgCatalogURL, authService, sessionService))
		imageCatalogGroup.Any("", proxy.NewImageCatalogProxy(imgCatalogURL, authService, sessionService))
	}

	// Swagger UI route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
