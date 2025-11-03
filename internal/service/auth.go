package service

import (
	"context"
	"fmt"
	"time"

	"github.com/histopathai/auth-service/internal/domain/model"
	"github.com/histopathai/auth-service/internal/domain/repository"
	"github.com/histopathai/auth-service/internal/shared/errors"
	"github.com/histopathai/auth-service/internal/shared/query"
)

type AuthService struct {
	authRepo repository.AuthRepository
	userRepo repository.UserRepository
}

func NewAuthService(authrepo repository.AuthRepository, userRepo repository.UserRepository) *AuthService {
	return &AuthService{
		authRepo: authrepo,
		userRepo: userRepo,
	}
}

func (s *AuthService) RegisterUser(ctx context.Context, register *model.RegisterUser) (*model.User, error) {

	// 1. Creata Firebase user
	authInfo, err := s.authRepo.Register(ctx, register)
	if err != nil {
		return nil, err
	}

	// 2. Create user record in the database (initially pending)
	user := &model.User{
		UID:         authInfo.UID,
		Email:       authInfo.Email,
		DisplayName: authInfo.DisplayName,
		Status:      model.StatusPending,
		Role:        model.RoleUnassigned,
	}

	// 3. Save user record
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.authRepo.Delete(ctx, authInfo.UID) // Rollback Firebase user creation
		return nil, fmt.Errorf("failed to create user record: %w", err)
	}

	return user, nil

}

func (s *AuthService) VerifyToken(ctx context.Context, idToken string) (*model.User, error) {

	// 1. Verify ID Token with Firebase
	authUser, err := s.authRepo.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}

	// 2. Retrieve full user profile from Firestore
	user, err := s.userRepo.GetByUID(ctx, authUser.UID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) ChangeUserPassword(ctx context.Context, uid string, newPassword string) error {
	return s.authRepo.ChangePassword(ctx, uid, newPassword)
}

func (s *AuthService) DeleteUser(ctx context.Context, uid string) error {
	if err := s.userRepo.Delete(ctx, uid); err != nil {
		return errors.NewInternalError("failed to delete user from database", err)
	}

	if err := s.authRepo.Delete(ctx, uid); err != nil {
		return errors.NewInternalError(fmt.Sprintf("CRITICAL: User deleted from DB but FAILED to delete from Auth. UID: %s", uid), err)
	}

	return nil
}

func (s *AuthService) GetUserByUID(ctx context.Context, uid string) (*model.User, error) {
	return s.userRepo.GetByUID(ctx, uid)
}

func (s *AuthService) ApproveUser(ctx context.Context, uid string) error {

	// 1. Retrieve the user by UID
	user, err := s.userRepo.GetByUID(ctx, uid)
	if err != nil {
		return err
	}

	//2. Ensure user is in pending approval or already approved state
	if user.Status == model.StatusActive && user.AdminApproved {
		detail := map[string]interface{}{
			"uid":    uid,
			"status": user.Status,
		}
		return errors.NewConflictError("user is already active and approved", detail)
	}

	targetRole := user.Role
	if user.Role == model.RoleUnassigned {
		targetRole = model.RoleUser
	}

	// 3. Update user status to active, set role and approval date
	err = s.SetUserRoleAndStatus(ctx, uid, targetRole, model.StatusActive, true)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) SuspendUser(ctx context.Context, uid string) error {

	// 1. Retrieve the user by UID
	user, err := s.userRepo.GetByUID(ctx, uid)
	if err != nil {
		detail := map[string]interface{}{
			"uid": uid,
		}
		return errors.NewValidationError("failed to retrieve user for suspension", detail)
	}
	// 2. Ensure user is active before suspending
	if user.Status != model.StatusActive {
		detail := map[string]interface{}{
			"uid":    uid,
			"status": user.Status,
		}
		return errors.NewConflictError("user is not active and cannot be suspended", detail)
	}

	// 3. Update user status to suspended
	err = s.SetUserRoleAndStatus(ctx, uid, user.Role, model.StatusSuspended, false)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) ActivateUser(ctx context.Context, uid string) error {

	// 1. Retrieve the user by UID
	user, err := s.userRepo.GetByUID(ctx, uid)
	if err != nil {
		return err
	}

	// 2. Ensure user is suspended before activating
	if user.Status != model.StatusSuspended {
		detail := map[string]interface{}{
			"uid":    uid,
			"status": user.Status,
		}
		return errors.NewConflictError("user is not suspended and cannot be activated", detail)
	}

	// 3. Update user status to active
	err = s.SetUserRoleAndStatus(ctx, uid, user.Role, model.StatusActive, true)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) PromoteUserToAdmin(ctx context.Context, uid string) error {
	user, err := s.userRepo.GetByUID(ctx, uid)
	if err != nil {
		return err
	}

	// 2. Ensure user is activated
	if user.Status != model.StatusActive {
		detail := map[string]interface{}{
			"uid":    uid,
			"status": user.Status,
		}
		return errors.NewConflictError("user is not active and cannot be promoted to admin", detail)
	}

	// 3. Check if user is already an admin
	if user.Role == model.RoleAdmin {
		detail := map[string]interface{}{
			"uid":  uid,
			"role": user.Role,
		}
		return errors.NewConflictError("user is already an admin", detail)
	}

	// 4. Update user role to admin
	err = s.SetUserRoleAndStatus(ctx, uid, model.RoleAdmin, user.Status, user.AdminApproved)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) SetUserRoleAndStatus(ctx context.Context, uid string, role model.UserRole, status model.UserStatus, adminApproved bool) error {

	updates := &model.UpdateUser{
		Role:          &role,
		Status:        &status,
		AdminApproved: &adminApproved,
	}
	if adminApproved {
		t := time.Now()
		updates.ApprovalDate = &t
	} else {
		updates.ApprovalDate = nil
	}

	err := s.userRepo.Update(ctx, uid, updates)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) ListUsers(ctx context.Context, pagination *query.Pagination) (*query.Result[*model.User], error) {
	return s.userRepo.List(ctx, pagination)
}
