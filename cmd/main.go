package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/histopathai/auth-service/pkg/config"
	"github.com/histopathai/auth-service/pkg/container"
	"github.com/histopathai/auth-service/pkg/logger"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	appLogger := logger.New(&cfg.Logging)
	slog.SetDefault(appLogger.Logger)

	appLogger.Info("Logger initialized", "level", cfg.Logging.Level, "format", cfg.Logging.Format)

	ctx := context.Background()
	appContainer, err := container.New(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize application container", "error", err)
		os.Exit(1)
	}

	defer func() {
		if err := appContainer.Close(); err != nil {
			appLogger.Error("Failed to close application container", "error", err)
		}
	}()

	engine := appContainer.Router.Setup()

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      engine,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	go func() {
		appLogger.Info("Starting HTTP server", "port", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	appLogger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("Error during server shutdown", "error", err)
		os.Exit(1)
	}
	appLogger.Info("Server gracefully stopped")
}
