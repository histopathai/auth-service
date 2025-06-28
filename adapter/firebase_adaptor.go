package adapter

import (
	"context"
	"fmt"

	"firebase.google.com/go/auth"
	"github.com/histopathai/auth-service/internal/models"
	"github.com/histopathai/auth-service/internal/repository"
)

var _ repository.AuthRepository = &FirebaseAuthAdapter{}

// FirebaseAdapter implements AuthRepository using Firebase Authentication.
type FirebaseAuthAdapter struct {
	client *auth.Client
}

// NewFirebaseAdapter creates a new FirebaseAdapter instance.
func NewFirebaseAuthAdapter(authClient *auth.Client) (*FirebaseAuthAdapter, error) {

	return &FirebaseAuthAdapter{
		client: authClient,
	}, nil
}

// RegisterUser registers a new user with the provided payload.
func (fa *FirebaseAuthAdapter) RegisterUser(ctx context.Context, payload *models.UserRegistrationPayload) (*models.UserAuthInfo, error) {
	params := (&auth.UserToCreate{}).
		Email(payload.Email).
		Password(payload.Password).
		DisplayName(payload.DisplayName).
		EmailVerified(false).
		Disabled(false)

	u, err := fa.client.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	authUser := &models.UserAuthInfo{
		UID:           u.UID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		DisplayName:   u.DisplayName,
	}

	return authUser, nil
}

func (fa *FirebaseAuthAdapter) VerifyIDToken(ctx context.Context, idToken string) (*models.UserAuthInfo, error) {

	token, err := fa.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("ID Token Couldn't Verified: %w", err)
	}

	authUser := &models.UserAuthInfo{
		UID:           token.UID,
		Email:         token.Claims["email"].(string),
		EmailVerified: token.Claims["email_verified"].(bool),
	}

	if displayName, ok := token.Claims["name"].(string); ok {
		authUser.DisplayName = displayName
	}

	return authUser, nil
}

func (fa *FirebaseAuthAdapter) ChangeUserPassword(ctx context.Context, uid, newPassword string) error {

	_, err := fa.client.UpdateUser(ctx, uid, (&auth.UserToUpdate{}).Password(newPassword))
	if err != nil {
		return fmt.Errorf("failed to change user password: %w", err)
	}
	return nil
}

func (fa *FirebaseAuthAdapter) DeleteAuthUser(ctx context.Context, uid string) error {
	if err := fa.client.DeleteUser(ctx, uid); err != nil {
		return fmt.Errorf("failed to delete user with UID %s: %w", uid, err)
	}
	return nil
}

// GetUserAuthInfo retrieves user authentication information by UID.
func (fa *FirebaseAuthAdapter) GetUserAuthInfo(ctx context.Context, uid string) (*models.UserAuthInfo, error) {
	u, err := fa.client.GetUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	authUser := &models.UserAuthInfo{
		UID:           u.UID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		DisplayName:   u.DisplayName,
	}

	return authUser, nil
}
