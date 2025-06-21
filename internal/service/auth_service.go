package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/histopathai/auth-service/internal/models"
	"github.com/histopathai/auth-service/internal/repository"
)

// AuthService defines the interface for authentication and user management operations.
type AuthService interface {
	VerifyIDToken(ctx context.Context, idToken string) (*models.TokenClaims, error)
	CreateUser(ctx context.Context, req *models.UserCreateRequest) (*models.User, error)
	UpdateUserRole(ctx context.Context, uid string, role string) error
	DeleteUser(ctx context.Context, uid string) error
	ActivateUser(ctx context.Context, uid string) error
	DeactivateUser(ctx context.Context, uid string) error
	ListUsers(ctx context.Context) ([]*models.User, error)
}

type authService struct {
	authClient AuthClient                // Interface authorizing operations
	userRepo   repository.UserRepository // Interface for user data operations
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(authClient AuthClient, userRepo repository.UserRepository) AuthService {
	return &authService{authClient: authClient, userRepo: userRepo}
}

// VerifyIDToken verifies an ID Token using the abstract AuthClient interface.
func (s *authService) VerifyIDToken(ctx context.Context, idToken string) (*models.TokenClaims, error) {
	// 1 Remove "Bearer " prefix if present
	idToken = strings.TrimPrefix(idToken, "Bearer ")
	if idToken == "" {
		return nil, fmt.Errorf("ID token is required")
	}
	// 2 Call the AuthClient's VerifyIDToken method
	tokenClaims, err := s.authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	return tokenClaims, nil
}

// CreateUser creates a new user in the auth provider and stores profile in the user repository.
func (s *authService) CreateUser(ctx context.Context, req *models.UserCreateRequest) (*models.User, error) {

	// 1. Create user in Auth provider using the abstract Authclient
	authClientReq := &AuthClientCreateUserRequest{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	}

	userRecord, err := s.authClient.CreateUser(ctx, authClientReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create user in auth provider: %w", err)
	}

	// 2. Save user role and profile to the user repository
	newUser := &models.User{
		UID:         userRecord.UID,
		Email:       userRecord.Email,
		DisplayName: userRecord.DisplayName,
		Role:        req.Role,
		Institution: req.Institution,
		IsActive:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.userRepo.CreateUser(ctx, newUser); err != nil {
		s.authClient.DeleteUser(ctx, userRecord.UID) // Clean up if user creation fails
		return nil, fmt.Errorf("failed to create user profile: %w", err)
	}
	return newUser, nil
}

// UpdateUserRole updates a user's role in both the auth provider and the user repository.
func (s *authService) UpdateUserRole(ctx context.Context, uid string, role string) error {
	if !models.ValidRoles[role] {
		return fmt.Errorf("invalid role: %s", role)
	}

	updates := map[string]interface{}{
		"Role": role,
	}

	if err := s.userRepo.UpdateUser(ctx, uid, updates); err != nil {
		return fmt.Errorf("failed to update user role in repository: %w", err)
	}

	return nil
}

// DeleteUser deletes a user from the auth  provider and the user repository.
func (s *authService) DeleteUser(ctx context.Context, uid string) error {
	// 1. Delete user from Auth provider
	if err := s.authClient.DeleteUser(ctx, uid); err != nil {
		return fmt.Errorf("failed to delete user from auth provider: %w", err)
	}
	// 2. Delete user from repository
	if err := s.userRepo.DeleteUser(ctx, uid); err != nil {
		return fmt.Errorf("failed to delete user from repository: %w", err)
	}
	return nil
}

// ListUsers retrieves all users from the user repository.
func (s *authService) ListUsers(ctx context.Context) ([]*models.User, error) {
	users, err := s.userRepo.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

// ActivateUser activates a user by setting IsActive to true in the user repository.
func (s *authService) ActivateUser(ctx context.Context, uid string) error {
	updates := map[string]interface{}{
		"IsActive":  true,
		"UpdatedAt": time.Now(),
	}

	if err := s.userRepo.UpdateUser(ctx, uid, updates); err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}
	return nil
}

// DeactivateUser deactivates a user by setting IsActive to false in the user repository.
func (s *authService) DeactivateUser(ctx context.Context, uid string) error {
	updates := map[string]interface{}{
		"IsActive":  false,
		"UpdatedAt": time.Now(),
	}

	if err := s.userRepo.UpdateUser(ctx, uid, updates); err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}
	return nil
}
