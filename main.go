package main

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"

	"github.com/histopathai/auth-service/internal/api"
	"github.com/histopathai/auth-service/internal/repository"
	"github.com/histopathai/auth-service/internal/service"
)

func main() {
	ctx := context.Background()

	// Read the service account key path from environment variable
	serviceAccountKeyPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if serviceAccountKeyPath == "" {
		log.Fatal("GOOGLE_APPLICATION_CREDENTIALS environment variable not set.")
	}
	opt := option.WithCredentialsFile(serviceAccountKeyPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing Firebase app: %v", err)
	}

	// Get Firebase Auth Client
	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Firebase Auth client: %v", err)
	}

	// Get Firestore Client
	firestoreClient, err := firestore.NewClient(ctx, os.Getenv("GCP_PROJECT_ID"))
	if err != nil {
		log.Fatalf("error getting Firestore client: %v", err)
	}
	defer firestoreClient.Close()

	// Inject dependencies and create services
	userRepo := repository.NewFirestoreUserRepository(firestoreClient)
	authService := service.NewAuthService(authClient, userRepo)
	authHandler := api.NewAuthHandler(authService)

	// Create Gin router
	router := gin.Default()

	// Setup API routes
	api.SetupAuthRoutes(router, authHandler)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Cloud Run listens on 8080 by default
	}
	log.Printf("Auth Service starting on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Auth Service failed to start: %v", err)
	}
}
