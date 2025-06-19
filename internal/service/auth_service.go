package service

import (
	"context"
	"fmt"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/histopathai/auth-service/internal/repository"
	"github.com/histopathai/auth-service/pkg/models"
)

// AuthService defines the interface for authentication and user management operations.
type AuthService interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
	CreateUser(ctx context.Context, req *models.UserCreateRequest) (*models.User, error)
	UpdateUserRole(ctx context.Context, uid string, newRole string) error
	DeleteUser(ctx context.Context, uid string) error
	ListUsers(ctx context.Context) ([]*models.User, error)
	ActivateUser(ctx context.Context, uid string) error
	DeactivateUser(ctx context.Context, uid string) error
}

type authService struct {
	firebaseAuth *auth.Client
	userRepo     repository.UserRepository
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(fbAuth *auth.Client, repo repository.UserRepository) AuthService {
	return &authService{firebaseAuth: fbAuth, userRepo: repo}
}

// VerifyIDToken verifies a Firebase ID Token
func (s *authService) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	// Remove "Bearer " prefix if present
	idToken = strings.TrimPrefix(idToken, "Bearer ")

	token, err := s.firebaseAuth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}
	return token, nil
}

// Createuser creates a new user in Firebase Auth and FireStore
func (s *authService) CreateUser(ctx context.Context, req *models.UserCreateRequest) (*models.User, error) {

	//1. Create user in Firebase Authentication
	params := (&auth.UserToCreate{}).
		Email(req.Email).
		Password(req.Password).
		DisplayName(req.DisplayName).
		EmailVerified(false) // E-mail verification can be handled later

	userRecord, err := s.firebaseAuth.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create user in Firebase: %w", err)
	}

	// 2. Save user role and profile to Firestore
	newUser := &models.User{
		UID:         userRecord.UID,
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Role:        req.Role, // Should be assigned by admin or system
		Institution: req.Institution,
		IsActive:    false, // Default to false, can be activated later by admin
	}

	if err := s.userRepo.CreateUser(ctx, newUser); err != nil {
		// if Firestore fails, delete the user from Firebase
		s.firebaseAuth.DeleteUser(ctx, userRecord.UID) //Rollback
		return nil, fmt.Errorf("failed to create user in Firestore: %w", err)
	}
	return newUser, nil
}

// UpdateUserRole updates a user's role in Firestore.
// This typically requires admin privileges.
func (s *authService) UpdateUserRole(ctx context.Context, uid string, newRole string) error {
	// Validate the new role (e.g., "admin", "viewer", etc.)
	if !models.ValidRoles[newRole] {
		return fmt.Errorf("invalid role specified: %s", newRole)
	}

	updates := map[string]interface{}{
		"Role": newRole,
	}

	if err := s.userRepo.UpdateUser(ctx, uid, updates); err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}
	return nil
}

// DeleteUser deletes a user from Firebase Auth and Firestore.
func (s *authService) DeleteUser(ctx context.Context, uid string) error {
	// 1. Delete user from Firebase Authentication
	if err := s.firebaseAuth.DeleteUser(ctx, uid); err != nil {
		return fmt.Errorf("failed to delete user from Firebase: %w", err)
	}

	// 2. Delete user from Firestore
	if err := s.userRepo.DeleteUser(ctx, uid); err != nil {
		return fmt.Errorf("failed to delete user from Firestore: %w", err)
	}

	return nil
}

// ListUsers retrieves all users from Firestore.
func (s *authService) ListUsers(ctx context.Context) ([]*models.User, error) {
	users, err := s.userRepo.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

// ActivateUser activates a user by setting IsActive to true.
func (s *authService) ActivateUser(ctx context.Context, uid string) error {
	updates := map[string]interface{}{
		"IsActive": true,
	}

	if err := s.userRepo.UpdateUser(ctx, uid, updates); err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}
	return nil
}

// DeactivateUser deactivates a user by setting IsActive to false.
func (s *authService) DeactivateUser(ctx context.Context, uid string) error {
	updates := map[string]interface{}{
		"IsActive": false,
	}

	if err := s.userRepo.UpdateUser(ctx, uid, updates); err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}
	return nil
}
