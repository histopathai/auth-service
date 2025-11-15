package container

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"

	"github.com/histopathai/auth-service/internal/api/http/router"
	"github.com/histopathai/auth-service/internal/domain/repository"
	firebaseAuth "github.com/histopathai/auth-service/internal/infrastructure/auth/firebase"
	firestoreRepo "github.com/histopathai/auth-service/internal/infrastructure/storage/firestore"
	memoryRepo "github.com/histopathai/auth-service/internal/infrastructure/storage/memory"
	"github.com/histopathai/auth-service/internal/service"
	"github.com/histopathai/auth-service/pkg/config"
	"github.com/histopathai/auth-service/pkg/logger"
)

type Container struct {
	Config *config.Config
	Logger *logger.Logger

	//Infrastructure
	FirebaseApp     *firebase.App
	AuthClient      *auth.Client
	FirestoreClient *firestore.Client

	//Repositories
	AuthRepository    repository.AuthRepository
	UserRepository    repository.UserRepository
	SessionRepository repository.SessionRepository

	//Services
	AuthService    *service.AuthService
	SessionService *service.SessionService

	//Router
	Router *router.Router
}

func New(ctx context.Context, cfg *config.Config, logger *logger.Logger) (*Container, error) {
	c := &Container{
		Config: cfg,
		Logger: logger,
	}

	if err := c.initInfrastructure(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	if err := c.initRepositories(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}
	if err := c.initServices(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	if err := c.initHTTPLayer(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize HTTP layer: %w", err)
	}
	c.Logger.Info("Container initialized successfully")
	return c, nil
}

func (c *Container) initInfrastructure(ctx context.Context) error {

	fbApp, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize Firebase app: %w", err)
	}
	c.FirebaseApp = fbApp

	authClient, err := fbApp.Auth(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize Firebase Auth client: %w", err)
	}
	c.AuthClient = authClient

	firestoreClient, err := fbApp.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize Firestore client: %w", err)
	}
	c.FirestoreClient = firestoreClient

	c.Logger.Info("Infrastructure initialized")
	return nil
}

func (c *Container) initRepositories(ctx context.Context) error {

	c.AuthRepository = firebaseAuth.NewFirebaseAuthRepository(c.AuthClient)
	c.UserRepository = firestoreRepo.NewFirestoreUserRepository(c.FirestoreClient, "users")

	c.SessionRepository = memoryRepo.NewInMemorySessionRepository(memoryRepo.DefaultMaxSessionsPerUser)
	c.Logger.Info("Repositories initialized")
	return nil
}

func (c *Container) initServices(ctx context.Context) error {

	c.AuthService = service.NewAuthService(c.AuthRepository, c.UserRepository)

	c.SessionService = service.NewSessionService(c.SessionRepository, *c.AuthService, c.Logger.Logger)
	c.Logger.Info("Services initialized")
	return nil
}

func (c *Container) initHTTPLayer(ctx context.Context) error {
	routerConfig := &router.RouterConfig{
		AuthService:    c.AuthService,
		SessionService: c.SessionService,
		Logger:         c.Logger.Logger,
		MainServiceURL: c.Config.MainServiceURL,
		Config:         c.Config,
	}

	appRouter, err := router.NewRouter(routerConfig, c.Config)
	if err != nil {
		return fmt.Errorf("failed to initialize router: %w", err)
	}
	c.Router = appRouter

	c.Logger.Info("HTTP Layer initialized")
	return nil
}

func (c *Container) Close() error {
	c.Logger.Info("Closing Container resources")

	if err := c.FirestoreClient.Close(); err != nil {
		return fmt.Errorf("failed to close Firestore client: %w", err)
	}

	c.Logger.Info("Container resources closed successfully")
	return nil
}
