package repository

import (
	"context"

	"github.com/histopathai/auth-service/internal/models"
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	GetUserByUID(ctx context.Context, uid string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, uid string, updates map[string]interface{}) error
	DeleteUser(ctx context.Context, uid string) error
	ListUsers(ctx context.Context) ([]*models.User, error)
}
