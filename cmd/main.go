package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/histopathai/auth-service/adapter"
	"github.com/histopathai/auth-service/config"

	"github.com/histopathai/auth-service/internal/service"
	"github.com/histopathai/auth-service/internal/utils"
	"github.com/histopathai/auth-service/server"
)

func main() {
	// Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config : %v", err)
	}

	// Initialize Firabase  Auth repository
	authRepo, err := adapter.NewFirebaseAuthAdapter(cfg.Firebase)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase Auth adapter: %v", err)
	}

	//Initialize Firestore repository
	userRepo, err := adapter.NewFirestoreAdapter(cfg.Firestore)
	if err != nil {
		log.Fatalf("Failed to initialize Firestore adapter: %v", err)
	}

	// Initialize AuthService
	emailService := utils.NewMailService(cfg.SMTP)

	// Initialize AuthService
	authService := service.NewAuthService(authRepo, userRepo, emailService)

	// Create and start the server
	authServer := server.NewServer(cfg, authService)
	if err := authServer.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
