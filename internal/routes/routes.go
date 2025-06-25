package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/internal/handlers"
	"github.com/histopathai/auth-service/internal/middleware"
	"github.com/histopathai/auth-service/internal/models"
	"github.com/histopathai/auth-service/internal/service"
)

//SetoRoutes configures all routes for the application

func SetupRoutes(authService service.AuthService, rateLimiter *middleware.RateLimiter) *gin.Engine {

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	adminHandler := handlers.NewAdminHandler(authService)
	healthHandler := handlers.NewHealthHandler()

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Create a new Gin router
	router := gin.New()

	//Global middleware
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

	//Public Authentication routes
	auth := api.Group("/auth")
	{
		//Registration
		auth.POST("/register", authHandler.Register)

		//Token verification
		auth.POST("/verify", authHandler.VerifyToken)

		//Password Reset initiation (by UID - requires admin or self)
		auth.POST("/password-reset/:uid",
			authMiddleware.RequireAuth(),
			middleware.SelfOrAdminOnly(),
			authHandler.InitiatePasswordReset)

	}

	//Protected routes (require authentication)
	user := api.Group("/user")
	user.Use(authMiddleware.RequireAuth())
	user.Use(authMiddleware.RequireStatus(models.StatusActive)) // Ensure user is active)
	{
		// Get own profile
		user.GET("/profile", authHandler.GetProfile)

		//Change own password
		user.PUT("/password", authHandler.ChangePassword)

		//Delete own account
		user.DELETE("/account", authHandler.DeleteAccount)
	}

	//Admin routes (require admin role)
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

			// Deactivate user account
			users.POST("/:uid/deactivate", adminHandler.DeactivateUser)

			// Initiate email verification
			users.POST("/:uid/verify-email", adminHandler.InitiateEmailVerification)
		}
	}
	return router
}
