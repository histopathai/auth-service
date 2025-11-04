package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/internal/api/http/handler"
	"github.com/histopathai/auth-service/internal/api/http/middleware"
	"github.com/histopathai/auth-service/internal/api/http/proxy"
	"github.com/histopathai/auth-service/internal/domain/model"
	"github.com/histopathai/auth-service/internal/service"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	engine         *gin.Engine
	authHandler    *handler.AuthHandler
	adminHandler   *handler.AdminHandler
	healthHandler  *handler.HealthHandler
	sessionHandler *handler.SessionHandler
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
	sessionHandler := handler.NewSessionHandler(config.SessionService, config.AuthService, config.Logger)

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
		sessionHandler: sessionHandler,
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

	r.engine.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(204) // No Content
	})

	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
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

		sessions := v1.Group("/sessions")
		{
			sessions.POST("", r.sessionHandler.CreateSession)

			authenticatedSessions := sessions.Group("")
			authenticatedSessions.Use(r.authMiddleware.RequireAuth())
			authenticatedSessions.Use(r.authMiddleware.RequireStatus(model.StatusActive))
			{
				authenticatedSessions.GET("", r.sessionHandler.ListMySessions)
				authenticatedSessions.GET("/stats", r.sessionHandler.GetMySessionStats)
				authenticatedSessions.POST("/revoke-all", r.sessionHandler.RevokeAllMySessions)
				authenticatedSessions.DELETE("/:session_id", r.sessionHandler.RevokeSession)
				authenticatedSessions.POST("/:session_id/extend", r.sessionHandler.ExtendSession)
			}

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

				users.GET("/:uid/sessions", r.sessionHandler.ListUserSessions)
				users.DELETE("/:uid/sessions", r.sessionHandler.RevokeAllUserSessions)
			}

			adminSessions := admin.Group("/sessions")
			{
				adminSessions.DELETE("/:session_id", r.sessionHandler.RevokeUserSession)
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
			"POST /api/v1/sessions",
			"GET /api/v1/sessions",
			"GET /api/v1/sessions/stats",
			"POST /api/v1/sessions/revoke-all",
			"DELETE /api/v1/sessions/:session_id",
			"POST /api/v1/sessions/:session_id/extend",
			"GET /api/v1/admin/users",
			"GET /api/v1/admin/users/:uid",
			"POST /api/v1/admin/users/:uid/approve",
			"POST /api/v1/admin/users/:uid/suspend",
			"POST /api/v1/admin/users/:uid/make-admin",
			"POST /api/v1/admin/users/:uid/change-password",
			"GET /api/v1/admin/users/:uid/sessions",
			"DELETE /api/v1/admin/users/:uid/sessions",
			"DELETE /api/v1/admin/sessions/:session_id",
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
