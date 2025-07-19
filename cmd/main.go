package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	firebase "firebase.google.com/go"
	"github.com/histopathai/auth-service/adapter"
	"github.com/histopathai/auth-service/config"
	"github.com/histopathai/auth-service/internal/service"
	"github.com/histopathai/auth-service/server"
)

func main() {
	// Load configuration from env
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Firebase App init (uses GOOGLE_APPLICATION_CREDENTIALS if set, otherwise default credentials)
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("❌ Failed to initialize Firebase app: %v", err)
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("❌ Failed to initialize Firebase Auth client: %v", err)
	}

	firestoreClient, err := app.Firestore(context.Background())
	if err != nil {
		log.Fatalf("❌ Failed to initialize Firestore client: %v", err)
	}

	authRepo, err := adapter.NewFirebaseAuthAdapter(authClient)
	if err != nil {
		log.Fatalf("❌ Failed to initialize Firebase Auth adapter: %v", err)
	}

	userRepo, err := adapter.NewFirestoreAdapter(firestoreClient, "users")
	if err != nil {
		log.Fatalf("❌ Failed to initialize Firestore adapter: %v", err)
	}

	authService := service.NewAuthService(authRepo, userRepo)

	// 🆕 Initialize session service
	sessionService := service.NewImageSessionService(authService)
	slog.Info("✅ Image Session Service initialized")

	// 🔄 Server'ı session service ile initialize et
	authServer := server.NewServer(cfg, authService, sessionService)
	if err := authServer.Start(); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
