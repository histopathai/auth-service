package service

import (
	"context"
	"fmt"
	"time"

	"github.com/histopathai/auth-service/internal/models"
	"github.com/histopathai/auth-service/internal/repository"
	"github.com/histopathai/auth-service/internal/utils"
)

// Ensure AuthServiceImpl implements AuthService interface

var _ AuthService = &AuthServiceImpl{}

// AuthServiceImpl implements the AuthService interface.
type AuthServiceImpl struct {
	authRepo    repository.AuthRepository
	userRepo    repository.UserRepository
	mailService utils.EmailService
}

// NewAuthService creates a new instance of AuthServiceImpl.
func NewAuthService(authRepo repository.AuthRepository, userRepo repository.UserRepository, mailService utils.EmailService) *AuthServiceImpl {
	return &AuthServiceImpl{
		authRepo:    authRepo,
		userRepo:    userRepo,
		mailService: mailService,
	}
}

// RegisterUser handles the full registration process.
func (s *AuthServiceImpl) RegisterUser(ctx context.Context, payload *models.UserRegistrationPayload) (*models.User, error) {

	// 1. Creata Firebase user
	authInfo, err := s.authRepo.RegisterUser(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	// 2. Create user record in the database (initially pending)
	now := time.Now()
	user := &models.User{
		UID:           authInfo.UID,
		Email:         authInfo.Email,
		DisplayName:   authInfo.DisplayName,
		CreatedAt:     now,
		UpdatedAt:     now,
		Status:        models.StatusPending,
		Role:          models.RoleUnassigned,
		AdminApproved: false,
		ApprovalDate:  time.Time{},
	}

	// 3. Save user record
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		s.authRepo.DeleteAuthUser(ctx, authInfo.UID) // Rollback Firebase user creation
		return nil, fmt.Errorf("failed to create user record: %w", err)
	}

	return user, nil

}

// VerifyToken verifies the provided ID token and retrieves the full user profile from Firestore.
func (s *AuthServiceImpl) VerifyToken(ctx context.Context, idToken string) (*models.User, error) {

	// 1. Verify the ID token with Firebase
	authInfo, err := s.authRepo.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	// 2. Retrieve the user profile from the database
	user, err := s.userRepo.GetUserByUID(ctx, authInfo.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user profile: %w", err)
	}

	// 3. Check if the user is active
	if user.Status != models.StatusActive {
		return nil, fmt.Errorf("user account is not active: %s", user.Status)
	}

	return user, nil
}

// InitiatePasswordReset starts the password reset process for a user.
func (s *AuthServiceImpl) InitiatePasswordReset(ctx context.Context, uid string) error {
	// 1. Retrieve the user by UID
	user, err := s.userRepo.GetUserByUID(ctx, uid)
	if err != nil {
		return fmt.Errorf("failed to retrieve user: %w", err)
	}

	// 2. Create a password reset link using Firebase
	resetLink, err := s.authRepo.CreatePasswordResetLink(ctx, user.Email)
	if err != nil {
		return fmt.Errorf("failed to create password reset link: %w", err)
	}

	subject := "Password Reset Request"
	body := fmt.Sprintf(`
		<p>Hi %s,</p>
		<p>We received a request to reset your password. If you did not make this request, please ignore this email.</p>
		<p>To reset your password, please click the link below:</p>
		<p><a href="%s">Reset Password</a></p>
		<p>If you have any questions, feel free to contact our support team.</p>
		<p>Best regards,</p>
		`, user.DisplayName, resetLink)

	// 3. Send the reset link via email
	if err := s.mailService.SendEmail(ctx, user.Email, subject, body); err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}
	return nil
}

// ChangePassword changes the user's password.
func (s *AuthServiceImpl) ChangePassword(ctx context.Context, uid string, newPassword string) error {
	// 1. Change the password in Firebase Auth
	if err := s.authRepo.ChangeUserPassword(ctx, uid, newPassword); err != nil {
		return fmt.Errorf("failed to change user password: %w", err)
	}
	return nil
}

// DeleteUser deletes a user from both Firebase Auth and the database.
func (s *AuthServiceImpl) DeleteUser(ctx context.Context, uid string) error {
	// 1. Delete the user from Firebase Auth
	if err := s.authRepo.DeleteAuthUser(ctx, uid); err != nil {
		return fmt.Errorf("failed to delete user from Firebase Auth: %w", err)
	}

	// 2. Delete the user record from the database
	if err := s.userRepo.DeleteUser(ctx, uid); err != nil {
		return fmt.Errorf("failed to delete user record: %w", err)
	}

	return nil
}

// --- Admin-specific operations ---

// InitiateEmailVerification sends a verification email to the user.
func (s *AuthServiceImpl) InitiateEmailVerification(ctx context.Context, uid string) error {
	// 1. Retrieve the user by UID
	user, err := s.userRepo.GetUserByUID(ctx, uid)
	if err != nil {
		return fmt.Errorf("failed to retrieve user: %w", err)
	}
	// Ensure the user is pending and not already verified
	if user.Status != models.StatusPending {
		return fmt.Errorf("user is not pending verification: %s", user.Status)
	}

	// 2. Create a verification link using Firebase
	verificationLink, err := s.authRepo.CreateEmailVerificationLink(ctx, user.Email)
	if err != nil {
		return fmt.Errorf("failed to create email verification link: %w", err)
	}

	subject := "Email Verification Required"
	body := fmt.Sprintf(`
		<p>Hi %s,</p>
		<p>Thank you for registering with us. To complete your registration, please verify your email address by clicking the link below:</p>
		<p><a href="%s">Verify Email</a></p>
		<p>If you did not register, please ignore this email.</p>
		<p>Best regards,</p>
		`, user.DisplayName, verificationLink)

	// 3. Send the verification link via email
	if err := s.mailService.SendEmail(ctx, user.Email, subject, body); err != nil {
		return fmt.Errorf("failed to send email verification: %w", err)
	}
	return nil

}

// ApproveUser approves a pending user, assigns a role, and sets the approval date.
func (s *AuthServiceImpl) ApproveUser(ctx context.Context, uid string, role models.UserRole) error {
	// 1. Retrieve the user by UID
	user, err := s.userRepo.GetUserByUID(ctx, uid)
	if err != nil {
		return fmt.Errorf("failed to retrieve user: %w", err)
	}

	// 2. Ensure the user is pending approval or already approved
	if user.Status != models.StatusPending && user.AdminApproved {
		return fmt.Errorf("user is not pending approval or already approved: %s", user.Status)
	}

	// 3. Update user status, role, and approval date
	err = s.userRepo.SetUserRoleAndStatus(ctx, uid, role, models.StatusActive, true)

	if err != nil {
		return fmt.Errorf("failed to approve user: %w", err)
	}

	return s.InitiateEmailVerification(ctx, uid) // Send verification email
}

// SuspendUser suspends a user account.
func (s *AuthServiceImpl) SuspendUser(ctx context.Context, uid string) error {
	// 1. Retrieve the user by UID
	user, err := s.userRepo.GetUserByUID(ctx, uid)
	if err != nil {
		return fmt.Errorf("failed to retrieve user: %w", err)
	}

	// 2. Ensure the user is active before suspending
	if user.Status != models.StatusActive {
		return fmt.Errorf("user is not active: %s", user.Status)
	}

	// 3. Update user status to suspended
	if err := s.userRepo.SetUserRoleAndStatus(ctx, uid, user.Role, models.StatusSuspended, false); err != nil {
		return fmt.Errorf("failed to suspend user: %w", err)
	}

	return nil
}

// DeactivateUser deactivates a user account.
func (s *AuthServiceImpl) DeactivateUser(ctx context.Context, uid string) error {
	// 1. Retrieve the user by UID
	user, err := s.userRepo.GetUserByUID(ctx, uid)
	if err != nil {
		return fmt.Errorf("failed to retrieve user: %w", err)
	}

	// 2. Ensure the user is active or suspended before deactivating
	if user.Status != models.StatusActive && user.Status != models.StatusSuspended {
		return fmt.Errorf("user is not active or suspended: %s", user.Status)
	}

	// 3. Update user status to deactivated
	if err := s.userRepo.SetUserRoleAndStatus(ctx, uid, user.Role, models.StatusDeactivated, false); err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	return nil
}

// GetUser retrieves a user by their unique identifier.
func (s *AuthServiceImpl) GetUser(ctx context.Context, uid string) (*models.User, error) {
	// 1. Retrieve the user by UID
	user, err := s.userRepo.GetUserByUID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}

	return user, nil
}

// GetAllUsers retrives all users
func (s *AuthServiceImpl) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	// 1. Retrieve all users from the repository
	users, err := s.userRepo.GetAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all users: %w", err)
	}

	return users, nil
}

// ActivateUser activates a user account.
func (s *AuthServiceImpl) ActivateUser(ctx context.Context, uid string) error {
	// 1. Retrieve the user by UID
	user, err := s.userRepo.GetUserByUID(ctx, uid)
	if err != nil {
		return fmt.Errorf("failed to retrieve user: %w", err)
	}

	// 2. Ensure the user is suspended or deactivated before activating
	if user.Status != models.StatusSuspended && user.Status != models.StatusDeactivated {
		return fmt.Errorf("user is not suspended or deactivated: %s", user.Status)
	}

	// 3. Update user status to active
	if err := s.userRepo.SetUserRoleAndStatus(ctx, uid, user.Role, models.StatusActive, false); err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	return nil
}
