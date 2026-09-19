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
	// directory is nil when no readers group is configured; the data-access
	// operations then report a clear error instead of failing obscurely.
	directory repository.DirectoryRepository
}

func NewAuthService(
	authrepo repository.AuthRepository,
	userRepo repository.UserRepository,
	directory repository.DirectoryRepository,
) *AuthService {
	return &AuthService{
		authRepo:  authrepo,
		userRepo:  userRepo,
		directory: directory,
	}
}

// ── data access ───────────────────────────────────────────────────────────────
//
// Read-only access to the research data is carried by membership of a Google
// group that the IAM policy binds once. Granting is therefore a group
// membership change, not an IAM change, and it is deliberately independent of
// the platform role.

func (s *AuthService) GrantDataAccess(ctx context.Context, userID string) (*model.User, error) {
	return s.setDataAccess(ctx, userID, true)
}

func (s *AuthService) RevokeDataAccess(ctx context.Context, userID string) (*model.User, error) {
	return s.setDataAccess(ctx, userID, false)
}

func (s *AuthService) setDataAccess(ctx context.Context, userID string, grant bool) (*model.User, error) {
	if s.directory == nil {
		return nil, errors.NewInternalError(
			"data access is not configured: set READERS_GROUP_EMAIL", nil)
	}

	user, err := s.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.Email == "" {
		return nil, errors.NewValidationError("user has no email address",
			map[string]interface{}{"userID": userID})
	}

	// The group is changed first: if it fails the stored flag stays as it was,
	// so the record never claims access the group does not actually carry.
	if grant {
		err = s.directory.AddMember(ctx, user.Email)
	} else {
		err = s.directory.RemoveMember(ctx, user.Email)
	}
	if err != nil {
		return nil, errors.NewInternalError("failed to update group membership", err)
	}

	now := time.Now()
	updates := &model.UpdateUser{DataAccess: &grant, DataAccessAt: &now}
	if err := s.userRepo.Update(ctx, userID, updates); err != nil {
		return nil, err
	}

	user.DataAccess = grant
	user.DataAccessAt = now
	return user, nil
}

// SyncDataAccess re-reads the group and returns the authoritative answer,
// repairing the stored flag when the two have drifted (someone edited the group
// directly in the admin console).
func (s *AuthService) SyncDataAccess(ctx context.Context, userID string) (*model.User, error) {
	if s.directory == nil {
		return nil, errors.NewInternalError(
			"data access is not configured: set READERS_GROUP_EMAIL", nil)
	}

	user, err := s.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	member, err := s.directory.HasMember(ctx, user.Email)
	if err != nil {
		return nil, errors.NewInternalError("failed to read group membership", err)
	}

	if member != user.DataAccess {
		updates := &model.UpdateUser{DataAccess: &member}
		if err := s.userRepo.Update(ctx, userID, updates); err != nil {
			return nil, err
		}
		user.DataAccess = member
	}
	return user, nil
}

func (s *AuthService) RegisterUser(ctx context.Context, register *model.ConfirmRegisterUser) (*model.User, error) {

	// 1. Verify Firebase Auth
	authInfo, err := s.authRepo.VerifyIDToken(ctx, register.Token)
	if err != nil {
		return nil, err
	}

	if authInfo.Email != register.Email {
		return nil, errors.NewUnauthorizedError("email in token does not match registration email")
	}

	existingUser, err := s.userRepo.GetByEmail(ctx, register.Email)
	if err != nil {
		return nil, errors.NewInternalError("failed to check existing user by email", err)
	}
	if existingUser != nil {
		detail := map[string]interface{}{
			"email": register.Email,
		}
		return nil, errors.NewConflictError("user with this email already exists", detail)
	}

	// 2. Create user record in the database (initially pending)
	user := &model.User{
		UserID:      authInfo.UserID,
		Email:       authInfo.Email,
		DisplayName: register.DisplayName,
		Status:      model.StatusPending,
		Role:        model.RoleUnassigned,
	}

	// 3. Save user record
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.authRepo.Delete(ctx, authInfo.UserID) // Rollback Firebase user creation
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
	user, err := s.userRepo.GetByUserID(ctx, authUser.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) ChangeUserPassword(ctx context.Context, userID string, newPassword string) error {
	return s.authRepo.ChangePassword(ctx, userID, newPassword)
}

func (s *AuthService) DeleteUser(ctx context.Context, userID string) error {
	// Data access outlives the platform account: the readers group grants GCP
	// read access and knows nothing about this deletion. Remove the membership
	// first, and refuse to delete if that fails — an orphaned group member keeps
	// reading the data with no account left to audit it against.
	if s.directory != nil {
		user, err := s.userRepo.GetByUserID(ctx, userID)
		if err != nil {
			return err
		}
		if user.Email != "" {
			if err := s.directory.RemoveMember(ctx, user.Email); err != nil {
				return errors.NewInternalError(
					"refusing to delete the user: failed to remove them from the readers group, "+
						"which would leave their data access in place", err)
			}
		}
	}

	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return errors.NewInternalError("failed to delete user from database", err)
	}

	if err := s.authRepo.Delete(ctx, userID); err != nil {
		return errors.NewInternalError(fmt.Sprintf("CRITICAL: User deleted from DB but FAILED to delete from Auth. GetByUserID: %s", userID), err)
	}

	return nil
}

func (s *AuthService) GetUserByUserID(ctx context.Context, userID string) (*model.User, error) {
	return s.userRepo.GetByUserID(ctx, userID)
}

// ApproveUser activates a pending or suspended user. role is the group the
// admin puts a new user in, required while they have none; it is ignored for
// anyone who already has a role. Admin is never granted here — that is
// SetUserRole, a separate and deliberate step.
func (s *AuthService) ApproveUser(ctx context.Context, userID string, role model.UserRole) error {

	// 1. Retrieve the user by GetByUserID
	user, err := s.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	//2. Ensure user is in pending approval or already approved state
	if user.Status == model.StatusActive && user.AdminApproved {
		detail := map[string]interface{}{
			"userID": userID,
			"status": user.Status,
		}
		return errors.NewConflictError("user is already active and approved", detail)
	}

	// role only fills an empty slot. Someone who already has a role keeps it,
	// whatever the request says: a browser on the old bundle sends "user" with
	// every approval, and that must not turn a suspended datascientist into a
	// pathologist on reactivation.
	targetRole := user.Role
	if user.Role == model.RoleUnassigned {
		if role != model.RolePathologist && role != model.RoleDatascientist {
			return errors.NewValidationError("approving this user needs a role: pathologist or datascientist",
				map[string]interface{}{"userID": userID, "role": role})
		}
		targetRole = role
	}

	// 3. Update user status to active, set role and approval date
	err = s.SetUserRoleAndStatus(ctx, userID, targetRole, model.StatusActive, true)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) SuspendUser(ctx context.Context, userID string) error {

	// 1. Retrieve the user by GetByUserID
	user, err := s.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		detail := map[string]interface{}{
			"userID": userID,
		}
		return errors.NewValidationError("failed to retrieve user for suspension", detail)
	}
	// 2. Ensure user is active before suspending
	if user.Status != model.StatusActive {
		detail := map[string]interface{}{
			"userID": userID,
			"status": user.Status,
		}
		return errors.NewConflictError("user is not active and cannot be suspended", detail)
	}

	// 3. Update user status to suspended
	err = s.SetUserRoleAndStatus(ctx, userID, user.Role, model.StatusSuspended, false)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) ActivateUser(ctx context.Context, userID string) error {

	// 1. Retrieve the user by GetByUserID
	user, err := s.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// 2. Ensure user is suspended before activating
	if user.Status != model.StatusSuspended {
		detail := map[string]interface{}{
			"userID": userID,
			"status": user.Status,
		}
		return errors.NewConflictError("user is not suspended and cannot be activated", detail)
	}

	// 3. Update user status to active
	err = s.SetUserRoleAndStatus(ctx, userID, user.Role, model.StatusActive, true)
	if err != nil {
		return err
	}
	return nil
}

// SetUserRole moves an active user to another group. actorID is the admin
// making the change: nobody changes their own role, so the last admin cannot
// lock the platform out of its admin panel by demoting themselves.
func (s *AuthService) SetUserRole(ctx context.Context, actorID, userID string, role model.UserRole) error {
	if !role.IsAssignable() {
		return errors.NewValidationError("role must be admin, pathologist or datascientist",
			map[string]interface{}{"role": role})
	}
	if actorID == userID {
		return errors.NewValidationError("you cannot change your own role",
			map[string]interface{}{"userID": userID})
	}

	user, err := s.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if user.Status != model.StatusActive {
		detail := map[string]interface{}{
			"userID": userID,
			"status": user.Status,
		}
		return errors.NewConflictError("user is not active; approve them first", detail)
	}
	if user.Role == role {
		detail := map[string]interface{}{
			"userID": userID,
			"role":   user.Role,
		}
		return errors.NewConflictError("user already has this role", detail)
	}

	// Only the role: status, approval and its date stay as they are.
	return s.userRepo.Update(ctx, userID, &model.UpdateUser{Role: &role})
}

// PromoteUserToAdmin is SetUserRole(admin), kept for the make-admin endpoint.
func (s *AuthService) PromoteUserToAdmin(ctx context.Context, actorID, userID string) error {
	return s.SetUserRole(ctx, actorID, userID, model.RoleAdmin)
}

func (s *AuthService) SetUserRoleAndStatus(ctx context.Context, userID string, role model.UserRole, status model.UserStatus, adminApproved bool) error {

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

	err := s.userRepo.Update(ctx, userID, updates)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) ListUsers(ctx context.Context, pagination *query.Pagination) (*query.Result[*model.User], error) {
	return s.userRepo.List(ctx, pagination)
}
