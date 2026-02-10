package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/internal/api/http/handler"
	"github.com/histopathai/auth-service/internal/api/http/middleware"
	"github.com/histopathai/auth-service/internal/api/http/proxy"
	"github.com/histopathai/auth-service/internal/domain/model"
	"github.com/histopathai/auth-service/internal/service"
	"github.com/histopathai/auth-service/pkg/config"

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
	Config         *config.Config
}

type RouterConfig struct {
	AuthService    *service.AuthService
	SessionService *service.SessionService
	Logger         *slog.Logger
	MainServiceURL string
	Config         *config.Config
}

func NewRouter(config *RouterConfig, appConfig *config.Config) (*Router, error) {
	authHandler := handler.NewAuthHandler(*config.AuthService, config.Logger)
	adminHandler := handler.NewAdminHandler(*config.AuthService, config.Logger)
	healthHandler := handler.NewHealthHandler(config.Logger)
	sessionHandler := handler.NewSessionHandler(config.SessionService, config.AuthService, appConfig, config.Logger)

	authMiddleware := middleware.NewAuthMiddleware(
		*config.AuthService,
		config.SessionService,
		appConfig,
		config.Logger,
	)

	// Pass config to proxy
	mainProxy, err := proxy.NewMainServiceProxy(
		config.MainServiceURL,
		config.AuthService,
		config.SessionService,
		config.Config,
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

func (r *Router) Setup(appConfig *config.Config) *gin.Engine {

	if len(appConfig.Security.TrustedProxies) > 0 {
		r.engine.SetTrustedProxies(appConfig.Security.TrustedProxies)
	}

	// Global middleware
	r.engine.Use(middleware.RecoveryMiddleware())
	r.engine.Use(middleware.LoggingMiddleware())
	r.engine.Use(middleware.CORSMiddleware(appConfig))

	// Rate limiter
	rateLimiter := middleware.NewRateLimiter(100, 200)
	r.engine.Use(rateLimiter.RateLimit())

	r.engine.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(204)
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
			// Public endpoints (no authentication required)
			auth.POST("/register", r.authHandler.Register)
			auth.POST("/verify", r.authHandler.VerifyToken)

			// Protected endpoints (require session)
			authenticated := auth.Group("")
			authenticated.Use(r.authMiddleware.RequireSession())
			authenticated.Use(r.authMiddleware.RequireStatus(model.StatusActive))
			{
				authenticated.PUT("/password", r.authHandler.ChangePasswordSelf)
			}
		}

		// User routes (protected - require session or bearer token)
		user := v1.Group("/user")
		user.Use(r.authMiddleware.RequireAuthOrSession())
		user.Use(r.authMiddleware.RequireStatus(model.StatusActive))
		{
			user.GET("/profile", r.authHandler.GetProfile)
			user.DELETE("/account", r.authHandler.DeleteAccount)
		}

		// Session routes
		sessions := v1.Group("/sessions")
		{
			sessions.PUT("", r.sessionHandler.CreateSession)

			sessions.DELETE("/current", r.sessionHandler.Logout)

			authenticated := sessions.Group("")
			authenticated.Use(r.authMiddleware.RequireSession())
			authenticated.Use(r.authMiddleware.RequireStatus(model.StatusActive))
			{
				authenticated.GET("", r.sessionHandler.ListMySessions)
				authenticated.GET("/stats", r.sessionHandler.GetMySessionStats)
				authenticated.PUT("/revoke-all", r.sessionHandler.RevokeAllMySessions)
				authenticated.PUT("/:session_id/extend", r.sessionHandler.ExtendSession)
				authenticated.GET("/current", r.sessionHandler.GetCurrentSession)
				authenticated.DELETE("/:session_id", r.sessionHandler.RevokeSession)

			}
		}

		// Admin routes (admin only)
		admin := v1.Group("/admin")
		admin.Use(r.authMiddleware.RequireAuthOrSession())
		admin.Use(r.authMiddleware.RequireRole(model.RoleAdmin))
		admin.Use(r.authMiddleware.RequireStatus(model.StatusActive))
		{
			users := admin.Group("/users")
			{
				users.GET("", r.adminHandler.ListUsers)
				users.GET("/:user_id", r.adminHandler.GetUser)
				users.POST("/:user_id/approve", r.adminHandler.ApproveUser)
				users.POST("/:user_id/suspend", r.adminHandler.SuspendUser)
				users.POST("/:user_id/make-admin", r.adminHandler.MakeAdmin)
				users.PUT("/:user_id/delete", r.adminHandler.DeleteUser)
				users.GET("/:user_id/sessions", r.sessionHandler.ListUserSessions)
				users.DELETE("/:user_id/sessions", r.sessionHandler.RevokeAllUserSessions)

			}

			adminSessions := admin.Group("/sessions")
			{
				adminSessions.DELETE("/:session_id", r.sessionHandler.RevokeUserSession)
			}
		}

		// Main service proxy routes
		proxy := v1.Group("/proxy")
		{
			proxy.Any("/*proxyPath", r.mainProxy.Handler())
		}
	}

	r.logger.Info("Router setup completed",
		"routes", []string{
			"POST /api/v1/auth/register (public)",
			"POST /api/v1/auth/verify (public)",
			"PUT /api/v1/auth/password (session required)",
			"GET /api/v1/user/profile (auth or session)",
			"DELETE /api/v1/user/account (auth or session)",
			"PUT /api/v1/sessions (token in body)",
			"GET /api/v1/sessions/current (session required)",
			"GET /api/v1/sessions (session required)",
			"GET /api/v1/sessions/stats (session required)",
			"PUT /api/v1/sessions/revoke-all (session required)",
			"DELETE /api/v1/sessions/:session_id (session required)",
			"PUT /api/v1/sessions/:session_id/extend (session required)",
			"GET /api/v1/admin/users (admin + session or bearer)",
			"GET /api/v1/admin/users/:user_id (admin + session or bearer)",
			"POST /api/v1/admin/users/:user_id/approve (admin + session or bearer)",
			"POST /api/v1/admin/users/:user_id/suspend (admin + session or bearer)",
			"POST /api/v1/admin/users/:user_id/make-admin (admin + session or bearer)",
			"PUT /api/v1/admin/users/:user_id/delete (admin + session or bearer)",
			"GET /api/v1/admin/users/:user_id/sessions (admin + session or bearer)",
			"DELETE /api/v1/admin/users/:user_id/sessions (admin + session or bearer)",
			"DELETE /api/v1/admin/sessions/:session_id (admin + session or bearer)",
			"ANY /api/v1/proxy/*proxyPath (auth or session)",
			"GET /api/v1/health (public)",
			"GET /api/v1/health/ready (public)",
		},
	)

	return r.engine
}

func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
