package service

import (
	"context"

	"github.com/histopathai/auth-service/internal/models"
)

type AuthClientConfig struct {
	// Empty struct for future configurations
}

type AuthClientCreateUserRequest struct {
	Email       string
	Password    string
	DisplayName string
}

type AuthClientUserRecord struct {
	UID         string
	Email       string
	DisplayName string
}

type AuthClient interface {
	VerifyIDToken(ctx context.Context, idToken string) (*models.TokenClaims, error)
	CreateUser(ctx context.Context, req *AuthClientCreateUserRequest) (*AuthClientUserRecord, error)
	DeleteUser(ctx context.Context, uid string) error
}

type UserRepository interface {
	GetUserByUID(ctx context.Context, uid string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, uid string, updates map[string]interface{}) error
	DeleteUser(ctx context.Context, uid string) error
	ListUsers(ctx context.Context) ([]*models.User, error)
}
