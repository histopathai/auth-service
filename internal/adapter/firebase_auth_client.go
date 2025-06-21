package adapter

import (
	"context"
	"fmt"

	"firebase.google.com/go/v4/auth"
	"github.com/histopathai/auth-service/internal/models"
	"github.com/histopathai/auth-service/internal/service"
)

// FirebaseAuthConfig is the configuration for the Firebase Auth Client.
type FirebaseAuthConfig struct {
	ProjectID string
}

// FirebaseAuthClient implements the service.AuthClient interface for Firebase Authentication.
type FirebaseAuthClient struct {
	client *auth.Client
}

// NewFirebaseAuthClient creates a new FirebaseAuthClient adapter.
func NewFirebaseAuthClient(client *auth.Client, config FirebaseAuthConfig) service.AuthClient {
	return &FirebaseAuthClient{client: client}
}

// VerifyIDToken verifies a Firebase ID token and converts it to a generic models.TokenClaims
func (f *FirebaseAuthClient) VerifyIDToken(ctx context.Context, idToken string) (*models.TokenClaims, error) {
	token, err := f.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("firebase verification failed: %w", err)
	}

	claims := &models.TokenClaims{
		UID:             token.UID,
		Email:           token.Claims["email"].(string),
		IsEmailVerified: token.Firebase.SignInProvider == "password" && token.Claims["email_verified"].(bool)}

	if role, ok := token.Claims["role"].(string); ok {
		claims.Role = role
	} else {
		claims.Role = "unknown"
	}
	return claims, nil

}

// CreateUser creates a new user in Firebase Auth
func (f *FirebaseAuthClient) CreateUser(ctx context.Context, req *service.AuthClientCreateUserRequest) (*service.AuthClientUserRecord, error) {
	params := (&auth.UserToCreate{}).
		Email(req.Email).
		Password(req.Password).
		DisplayName(req.DisplayName).
		EmailVerified(false)

	userRecord, err := f.client.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("firebase user creation dailed: %w", err)
	}

	genericUserRecord := &service.AuthClientUserRecord{
		UID:         userRecord.UID,
		Email:       userRecord.Email,
		DisplayName: userRecord.DisplayName,
	}
	return genericUserRecord, nil
}

// DeeleteUSer deletes a user from Firebase Auth
func (f *FirebaseAuthClient) DeleteUser(ctx context.Context, uid string) error {
	if err := f.client.DeleteUser(ctx, uid); err != nil {
		return fmt.Errorf("firebase user deletion failed: %w", err)
	}
	return nil
}
