package main

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"

	"github.com/histopathai/auth-service/internal/adapter"
	"github.com/histopathai/auth-service/internal/api"
	"github.com/histopathai/auth-service/internal/repository"
	"github.com/histopathai/auth-service/internal/service"
)

func main() {

	ctx := context.Background()

	serviceAccountKeyPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if serviceAccountKeyPath == "" {
		log.Fatal("GOOGLE_APPLICATION_CREDENTIALS environment variable not set.")
	}
	opt := option.WithCredentialsFile(serviceAccountKeyPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing Firebase app: %v", err)
	}

	fbAuthClient, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Firebase Auth client: %v", err)
	}

	// Initialize Firestore client
	firestoreClient, err := firestore.NewClient(ctx, os.Getenv("GCP_PROJECT_ID"))
	if err != nil {
		log.Fatalf("error creating Firestore client: %v", err)
	}
	defer firestoreClient.Close()

	authClientAdapter := adapter.NewFirebaseAuthClient(fbAuthClient, adapter.FirebaseAuthConfig{})

	userRepo := repository.NewFirestoreUserRepository(firestoreClient)

	authService := service.NewAuthService(authClientAdapter, userRepo)
	authHandler := api.NewAuthHandler(authService)

	router := gin.Default()

	api.SetupAuthRoutes(router, authHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not set
	}
	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
