package repository

import (
	"context"

	"github.com/histopathai/auth-service/internal/domain/model"
	"github.com/histopathai/auth-service/internal/shared/query"
)

type UserRepository interface {
	// Create creates a new user in the repository.
	Create(ctx context.Context, user *model.User) error

	// GetByUID retrieves a user by their unique identifier (UID).
	GetByUID(ctx context.Context, uid string) (*model.User, error)

	// Update updates an existing user in the repository.
	Update(ctx context.Context, uid string, updates *model.UpdateUser) error

	// Delete deletes a user from the repository.
	Delete(ctx context.Context, uid string) error

	// List retrieves all users from the repository.
	List(ctx context.Context, pagination *query.Pagination) (*query.Result[*model.User], error)
}

type AuthRepository interface {

	// RegisterUser registers a new user with the provided payload in Firebase Auth.
	Register(ctx context.Context, payload *model.RegisterUser) (*model.User, error)

	// VerifyIDToken verifies the provided ID token and returns user information.
	VerifyIDToken(ctx context.Context, idToken string) (*model.User, error)

	// ChangeUserPassword changes the user's password in Firebase Auth.
	ChangePassword(ctx context.Context, uid, newPassword string) error

	// DeleteAuthUser deletes a user from Firebase Auth.
	Delete(ctx context.Context, uid string) error

	// GetUserAuthInfo retrieves user authentication information by UID.
	GetAuthInfo(ctx context.Context, uid string) (*model.User, error)
}
