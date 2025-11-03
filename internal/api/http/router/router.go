package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/internal/api/http/handler"
	"github.com/histopathai/auth-service/internal/api/http/middleware"
	"github.com/histopathai/auth-service/internal/api/http/proxy"
	"github.com/histopathai/auth-service/internal/domain/model"
	"github.com/histopathai/auth-service/internal/service"
)

type Router struct {
	engine         *gin.Engine
	authHandler    *handler.AuthHandler
	adminHandler   *handler.AdminHandler
	healthHandler  *handler.HealthHandler
	authMiddleware *middleware.AuthMiddleware
	logger         *slog.Logger
	mainProxy      *proxy.MainServiceProxy
}

type RouterConfig struct {
	AuthService    *service.AuthService
	SessionService *service.SessionService
	Logger         *slog.Logger
	MainServiceURL string
}

func NewRouter(config *RouterConfig) (*Router, error) {
	// Initialize handlers
	authHandler := handler.NewAuthHandler(*config.AuthService, config.Logger)
	adminHandler := handler.NewAdminHandler(*config.AuthService, config.Logger)
	healthHandler := handler.NewHealthHandler(config.Logger)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(*config.AuthService)

	// Initialize proxy
	mainProxy, err := proxy.NewMainServiceProxy(
		config.MainServiceURL,
		config.AuthService,
		config.SessionService,
		config.Logger,
	)
	if err != nil {
		return nil, err
	}

	return &Router{
		engine:         gin.New(),
		authHandler:    authHandler,
		adminHandler:   adminHandler,
		healthHandler:  healthHandler,
		authMiddleware: authMiddleware,
		mainProxy:      mainProxy,
		logger:         config.Logger,
	}, nil
}

func (r *Router) Setup() *gin.Engine {
	// Global middleware
	r.engine.Use(middleware.RecoveryMiddleware())
	r.engine.Use(middleware.LoggingMiddleware())
	r.engine.Use(middleware.CORSMiddleware())

	// Rate limiter
	rateLimiter := middleware.NewRateLimiter(100, 200)
	r.engine.Use(rateLimiter.RateLimit())

	// API v1 routes
	v1 := r.engine.Group("/api/v1")
	{
		// Health check routes (no auth required)
		health := v1.Group("/health")
		{
			health.GET("", r.healthHandler.Health)
			health.GET("/ready", r.healthHandler.Ready)
		}

		// Auth routes
		auth := v1.Group("/auth")
		{
			// Public endpoints
			auth.POST("/register", r.authHandler.Register)
			auth.POST("/verify", r.authHandler.VerifyToken)

			// Protected endpoints
			authenticated := auth.Group("")
			authenticated.Use(r.authMiddleware.RequireAuth())
			authenticated.Use(r.authMiddleware.RequireStatus(model.StatusActive))
			{
				authenticated.PUT("/password", r.authHandler.ChangePasswordSelf)
			}
		}

		// User routes (protected)
		user := v1.Group("/user")
		user.Use(r.authMiddleware.RequireAuth())
		user.Use(r.authMiddleware.RequireStatus(model.StatusActive))
		{
			user.GET("/profile", r.authHandler.GetProfile)
			user.DELETE("/account", r.authHandler.DeleteAccount)
		}

		// Admin routes (admin only)
		admin := v1.Group("/admin")
		admin.Use(r.authMiddleware.RequireAuth())
		admin.Use(r.authMiddleware.RequireRole(model.RoleAdmin))
		admin.Use(r.authMiddleware.RequireStatus(model.StatusActive))
		{
			users := admin.Group("/users")
			{
				users.GET("", r.adminHandler.ListUsers)
				users.GET("/:uid", r.adminHandler.GetUser)
				users.POST("/:uid/approve", r.adminHandler.ApproveUser)
				users.POST("/:uid/suspend", r.adminHandler.SuspendUser)
				users.POST("/:uid/make-admin", r.adminHandler.MakeAdmin)
				users.POST("/:uid/change-password", r.adminHandler.ChangePasswordForUser)
			}
		}

		// Main service proxy routes
		// All requests to /api/v1/proxy/* will be forwarded to main-service
		// Authentication is handled by the proxy middleware
		proxy := v1.Group("/proxy")
		{
			proxy.Any("/*proxyPath", r.mainProxy.Handler())
		}
	}

	r.logger.Info("Router setup completed",
		"routes", []string{
			"POST /api/v1/auth/register",
			"POST /api/v1/auth/verify",
			"PUT /api/v1/auth/password",
			"GET /api/v1/user/profile",
			"DELETE /api/v1/user/account",
			"GET /api/v1/admin/users",
			"GET /api/v1/admin/users/:uid",
			"POST /api/v1/admin/users/:uid/approve",
			"POST /api/v1/admin/users/:uid/suspend",
			"POST /api/v1/admin/users/:uid/make-admin",
			"POST /api/v1/admin/users/:uid/change-password",
			"ANY /api/v1/proxy/*proxyPath",
			"GET /api/v1/health",
			"GET /api/v1/health/ready",
		},
	)

	return r.engine
}

func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
