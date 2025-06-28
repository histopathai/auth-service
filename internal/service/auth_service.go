package service

import (
	"context"

	"github.com/histopathai/auth-service/internal/models"
)

// Auth Service defines the interface for authhentication and user management operations.
type AuthService interface {

	//RegisterUser handles the full registration process.
	RegisterUser(ctx context.Context, payload *models.UserRegistrationPayload) (*models.User, error)

	//VerifyToken verifies the provided ID token and returns the associated user.
	VerifyToken(ctx context.Context, idToken string) (*models.User, error)

	//ChangePassword changes the user's password.
	ChangePassword(ctx context.Context, uid string, newPassword string) error

	//DeleteUser deletes a user account by their unique identifier.
	DeleteUser(ctx context.Context, uid string) error

	// --- Admin-specific operations ---

	//ApproveUser approves a pending user, assigns a role, and sets the approval date.
	ApproveUser(ctx context.Context, uid string, role models.UserRole) error

	//SuspendUser suspends a user account.
	SuspendUser(ctx context.Context, uid string) error

	//ActivateUser activates a user account.
	ActivateUser(ctx context.Context, uid string) error

	//GetUser retrieves a user by their unique identifier.
	GetUser(ctx context.Context, uid string) (*models.User, error)

	//GetAllUsers retrieves all users with optional pagination.
	GetAllUsers(ctx context.Context) ([]*models.User, error)
}
