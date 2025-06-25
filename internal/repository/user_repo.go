package repository

import (
	"context"

	"github.com/histopathai/auth-service/internal/models"
)

// UserRepository
type UserRepository interface {
	// CreateUser creates a new user in the repository.
	CreateUser(ctx context.Context, user *models.User) error

	// GetUserByUID retrieves a user by their unique identifier (UID).
	GetUserByUID(ctx context.Context, uid string) (*models.User, error)

	//UpdateUser updates an existing user in the repository.
	UpdateUser(ctx context.Context, uid string, payload *models.UpdateUserRequest) (*models.User, error)

	// DeleteUser deletes a user from the repository.
	DeleteUser(ctx context.Context, uid string) error

	//SetUserRoleAndStatus sets the role and status of a user.
	SetUserRoleAndStatus(ctx context.Context, uid string, role models.UserRole, status models.UserStatus, adminApproved bool) error

	GetAllUsers(ctx context.Context) ([]*models.User, error)
}
