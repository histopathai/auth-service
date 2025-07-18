package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/config"
	"github.com/histopathai/auth-service/internal/middleware"
	"github.com/histopathai/auth-service/internal/routes"
	"github.com/histopathai/auth-service/internal/service"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	config     *config.Config
}

// NewServer creates a new Server instance
func NewServer(cfg *config.Config, authService service.AuthService) *Server {
	gin.SetMode(cfg.Server.GINMode)

	// Create rate limiter ( 10 requests per second, burst of 20)
	rateLimiter := middleware.NewRateLimiter(10, 20)

	//Setup routes
	router := routes.SetupRoutes(authService, rateLimiter, cfg.ImageCatalogURL)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		config:     cfg,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	slog.Info("Starting server", "port", s.config.Server.Port)

	// Start server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for the interrupt signal to gracefully shut down server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	// Create a context with a timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		return err
	}

	slog.Info("Server Exited")
	return nil
}
