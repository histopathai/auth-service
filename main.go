package main

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"

	"github.com/histopathai/auth-service/internal/adapters"
	"github.com/histopathai/auth-service/internal/api"
	"github.com/histopathai/auth-service/internal/service"
)

func main() {

	ctx := context.Background()

	// Initialize Firebase app
	var firebaseApp *firebase.App
	var err error

	serviceAccountKeyPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if serviceAccountKeyPath != "" {

		opt := option.WithCredentialsFile(serviceAccountKeyPath)
		firebaseApp, err = firebase.NewApp(ctx, nil, opt)
	} else {

		firebaseApp, err = firebase.NewApp(ctx, nil)
	}

	if err != nil {
		log.Fatalf("error initializing Firebase app: %v", err)
	}

	fbAuthClient, err := firebaseApp.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Firebase Auth client: %v", err)
	}

	firestoreClient, err := firestore.NewClient(ctx, os.Getenv("GCP_PROJECT_ID"))
	if err != nil {
		log.Fatalf("error creating Firestore client: %v", err)
	}
	defer firestoreClient.Close()

	authClientAdapter := adapters.NewFirebaseAuthClient(fbAuthClient, adapters.FirebaseAuthConfig{})
	userRepoAdapter := adapters.NewFirestoreUserRepository(firestoreClient)

	authService := service.NewAuthService(authClientAdapter, userRepoAdapter)
	authHandler := api.NewAuthHandler(authService)

	router := gin.Default()

	api.SetupAuthRoutes(router, authHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
