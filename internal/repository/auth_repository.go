package repository

import (
	"context"

	"github.com/histopathai/auth-service/internal/models"
)

// AuthRepository defines an interface for interacting with Firebase Authentication services.
type AuthRepository interface {

	// RegisterUser registers a new user with the provided payload in Firebase Auth.
	RegisterUser(ctx context.Context, payload *models.UserRegistrationPayload) (*models.UserAuthInfo, error)

	// VerifyIDToken verifies the provided ID token and returns user information.
	VerifyIDToken(ctx context.Context, idToken string) (*models.UserAuthInfo, error)

	// CreatePasswordResetLink creates a password reset link for the given email.
	// This link needs to be sent by a separate email service.
	CreatePasswordResetLink(ctx context.Context, email string) (string, error)

	// SendEmailVerification sends a verification email to the user.
	// Note: Firebase Admin SDK creates the link, actual sending needs a separate email service.
	CreateEmailVerificationLink(ctx context.Context, email string) (string, error)

	// ChangeUserPassword changes the user's password in Firebase Auth.
	ChangeUserPassword(ctx context.Context, uid, newPassword string) error

	// DeleteAuthUser deletes a user from Firebase Auth.
	DeleteAuthUser(ctx context.Context, uid string) error

	//Get UserAuthInfo retrieves user authentication information by UID.
	GetUserAuthInfo(ctx context.Context, uid string) (*models.UserAuthInfo, error)
}
