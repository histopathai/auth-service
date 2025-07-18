package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/internal/handlers"
	"github.com/histopathai/auth-service/internal/middleware"
	"github.com/histopathai/auth-service/internal/models"
	"github.com/histopathai/auth-service/internal/proxy"
	"github.com/histopathai/auth-service/internal/service"

	// Swagger dokümantasyonu için gerekli importlar
	_ "github.com/histopathai/auth-service/docs" // Oluşturulan docs/swagger.go dosyasını import edin
	swaggerFiles "github.com/swaggo/files"       // gin-swagger için gerekli dosyalar
	ginSwagger "github.com/swaggo/gin-swagger"   // gin ile Swagger entegrasyonu
)

// SetupRoutes configures all routes for the application
func SetupRoutes(authService service.AuthService, rateLimiter *middleware.RateLimiter, imgCatalogURL string) *gin.Engine {

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	adminHandler := handlers.NewAdminHandler(authService)
	healthHandler := handlers.NewHealthHandler()

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
	}

	// Protected routes (require authentication)
	user := api.Group("/user")
	user.Use(authMiddleware.RequireAuth())
	user.Use(authMiddleware.RequireStatus(models.StatusActive)) // Ensure user is active)
	{
		// Get own profile
		user.GET("/profile", authHandler.GetProfile)

		// Change own password
		user.PUT("/password", authHandler.ChangePasswordSelf)

		// Delete own account
		user.DELETE("/account", authHandler.DeleteAccount)
	}

	// Admin routes (require admin role)
	admin := api.Group("/admin")
	admin.Use(authMiddleware.RequireAuth())
	admin.Use(authMiddleware.RequireRole(models.RoleAdmin))      // Ensure user is admin
	admin.Use(authMiddleware.RequireStatus(models.StatusActive)) // Ensure user is active
	{

		users := admin.Group("/users")
		{

			// User management
			users.GET("/", adminHandler.GetAllUsers)

			// Get specific user by UID
			users.GET("/:uid", adminHandler.GetUser)

			// Approve user account
			users.POST("/:uid/approve", adminHandler.ApproveUser)

			// Suspend user account
			users.POST("/:uid/suspend", adminHandler.SuspendUser)

			// Change user password (admin)
			users.PUT("/:uid/password", adminHandler.ChangePasswordForUser)

			// Promote user to admin
			users.POST("/:uid/promote", adminHandler.MakeAdmin)

		}
	}

	// Swagger UI route
	// Bu satır, uygulamanızın /swagger yoluna gelen istekleri Swagger UI'a yönlendirir.

	//Proxies for image catalog service
	apiProxy := api.Group("/image-catalog")
	apiProxy.Use(authMiddleware.RequireAuth())
	apiProxy.Use(authMiddleware.RequireStatus(models.StatusActive)) // Ensure user is active
	{
		imageCatalogURL := imgCatalogURL
		apiProxy.Any("/*any", proxy.NewImageCatalogProxy(imageCatalogURL))
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
