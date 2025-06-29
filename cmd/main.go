// @title HistopathAI Auth Service API
// @version 1.0
// @description HistopathAI Authentication Service API documentation.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	firebase "firebase.google.com/go"
	"github.com/histopathai/auth-service/adapter"
	"github.com/histopathai/auth-service/config"
	"github.com/joho/godotenv"

	"github.com/histopathai/auth-service/internal/service"
	"github.com/histopathai/auth-service/internal/utils"
	"github.com/histopathai/auth-service/server"
)

func main() {

	// initialize environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env file not loaded, falling back to OS env vars")
	}

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

	// Initialize Firebase App with ProjectID
	app, err := firebase.NewApp(context.Background(), &firebase.Config{
		ProjectID: cfg.Firebase.ProjectID,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Firebase app: %v", err)
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize Firebase Auth client: %v", err)
	}

	// Initialize Firestore client
	firestoreClient, err := app.Firestore(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize Firestore client: %v", err)
	}

	authRepo, err := adapter.NewFirebaseAuthAdapter(authClient)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase Auth adapter: %v", err)
	}

	// Initialize Firestore repository
	userRepo, err := adapter.NewFirestoreAdapter(firestoreClient, cfg.Firestore.UsersCollection)
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
