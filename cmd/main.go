package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/histopathai/auth-service/docs"
	"github.com/histopathai/auth-service/pkg/config"
	"github.com/histopathai/auth-service/pkg/container"
	"github.com/histopathai/auth-service/pkg/logger"
)

// @title Histopath AI API
// @version 1.0
// @description API for auth session management and user authentication.
// @termsOfService http://histopathai.com/terms/

// @contact.name API Support
// @contact.url http://www.histopathai.com/support
// @contact.email histopathai@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {

	useHTTPS := flag.Bool("https", false, "Enable HTTPS (TLS) for development")
	flag.Parse()

	appConfig := config.LoadConfig()

	appLogger := logger.New(&appConfig.Logging)

	appLogger.Info("Starting application",
		"environment", appConfig.Server.Environment,
		"cookie_secure", appConfig.Cookie.Secure,
		"cookie_samesite", appConfig.Cookie.SameSite,
	)

	ctx := context.Background()
	appContainer, err := container.New(ctx, appConfig, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize application container", "error", err)
		os.Exit(1)
	}

	defer func() {
		if err := appContainer.Close(); err != nil {
			appLogger.Error("Failed to close application container", "error", err)
		}
	}()

	engine := appContainer.Router.Setup(appConfig)

	server := &http.Server{
		Addr:         ":" + appConfig.Server.Port,
		Handler:      engine,
		ReadTimeout:  time.Duration(appConfig.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(appConfig.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(appConfig.Server.IdleTimeout) * time.Second,
	}

	go func() {
		appLogger.Info("Starting HTTP server", "port", appConfig.Server.Port)
		if *useHTTPS && appConfig.Server.Environment == "development" {
			appLogger.Info("HTTPS enabled for development",
				"cert_path", appConfig.TLS.CertPath,
				"key_path", appConfig.TLS.KeyPath,
			)
			if err := server.ListenAndServeTLS(appConfig.TLS.CertPath, appConfig.TLS.KeyPath); err != nil && !errors.Is(err, http.ErrServerClosed) {
				appLogger.Error("HTTPS server error", "error", err)
				os.Exit(1)
			}

		} else {
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				appLogger.Error("HTTP server error", "error", err)
				os.Exit(1)
			}
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
