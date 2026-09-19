package service

import (
	"context"
	stderrors "errors"
	"testing"
	"time"

	"github.com/histopathai/auth-service/internal/domain/model"
	"github.com/histopathai/auth-service/internal/shared/errors"
	"github.com/histopathai/auth-service/internal/shared/query"
)

// fakeUsers is an in-memory UserRepository holding just what the role logic reads and writes.
type fakeUsers struct {
	users map[string]*model.User
}

func (f *fakeUsers) Create(_ context.Context, u *model.User) error {
	f.users[u.UserID] = u
	return nil
}

func (f *fakeUsers) GetByUserID(_ context.Context, id string) (*model.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, errors.NewNotFoundError("user not found")
	}
	copy := *u
	return &copy, nil
}

func (f *fakeUsers) GetByEmail(context.Context, string) (*model.User, error) {
	return nil, errors.NewNotFoundError("user not found")
}

func (f *fakeUsers) Update(_ context.Context, id string, up *model.UpdateUser) error {
	u, ok := f.users[id]
	if !ok {
		return errors.NewNotFoundError("user not found")
	}
	if up.Role != nil {
		u.Role = *up.Role
	}
	if up.Status != nil {
		u.Status = *up.Status
	}
	if up.AdminApproved != nil {
		u.AdminApproved = *up.AdminApproved
	}
	if up.ApprovalDate != nil {
		u.ApprovalDate = *up.ApprovalDate
	}
	return nil
}

func (f *fakeUsers) Delete(_ context.Context, id string) error {
	delete(f.users, id)
	return nil
}

func (f *fakeUsers) List(context.Context, *query.Pagination) (*query.Result[*model.User], error) {
	return &query.Result[*model.User]{}, nil
}

func newRoleService(users ...*model.User) (*AuthService, *fakeUsers) {
	repo := &fakeUsers{users: map[string]*model.User{}}
	for _, u := range users {
		repo.users[u.UserID] = u
	}
	return NewAuthService(nil, repo, nil), repo
}

func errType(t *testing.T, err error) errors.ErrorType {
	t.Helper()
	var e *errors.Err
	if !stderrors.As(err, &e) {
		t.Fatalf("expected *errors.Err, got %v", err)
	}
	return e.Type
}

func TestApproveUser(t *testing.T) {
	approved := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	cases := []struct {
		name     string
		user     model.User
		role     model.UserRole
		wantErr  errors.ErrorType
		wantRole model.UserRole
	}{
		{"new user becomes pathologist", model.User{Status: model.StatusPending, Role: model.RoleUnassigned},
			model.RolePathologist, "", model.RolePathologist},
		{"new user becomes datascientist", model.User{Status: model.StatusPending, Role: model.RoleUnassigned},
			model.RoleDatascientist, "", model.RoleDatascientist},
		{"new user needs a role", model.User{Status: model.StatusPending, Role: model.RoleUnassigned},
			"", errors.ErrorTypeValidation, model.RoleUnassigned},
		{"admin is not granted by approval", model.User{Status: model.StatusPending, Role: model.RoleUnassigned},
			model.RoleAdmin, errors.ErrorTypeValidation, model.RoleUnassigned},
		{"reactivation keeps the role", model.User{Status: model.StatusSuspended, Role: model.RoleDatascientist},
			"", "", model.RoleDatascientist},
		// The old bundle sends "user" (normalized to pathologist) with every approval.
		{"reactivation ignores a role in the request", model.User{Status: model.StatusSuspended, Role: model.RoleDatascientist},
			model.RolePathologist, "", model.RoleDatascientist},
		{"reactivated admin stays admin", model.User{Status: model.StatusSuspended, Role: model.RoleAdmin},
			model.RolePathologist, "", model.RoleAdmin},
		{"already active", model.User{Status: model.StatusActive, AdminApproved: true, Role: model.RolePathologist, ApprovalDate: approved},
			"", errors.ErrorTypeConflict, model.RolePathologist},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := tc.user
			u.UserID = "u1"
			svc, repo := newRoleService(&u)

			err := svc.ApproveUser(context.Background(), "u1", tc.role)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected %s, got nil", tc.wantErr)
				}
				if got := errType(t, err); got != tc.wantErr {
					t.Fatalf("error type = %s, want %s", got, tc.wantErr)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := repo.users["u1"]
			if got.Role != tc.wantRole {
				t.Errorf("role = %q, want %q", got.Role, tc.wantRole)
			}
			if tc.wantErr == "" && (got.Status != model.StatusActive || !got.AdminApproved) {
				t.Errorf("user not activated: status=%q approved=%v", got.Status, got.AdminApproved)
			}
		})
	}
}

func TestSetUserRole(t *testing.T) {
	approved := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	active := func(role model.UserRole) model.User {
		return model.User{Status: model.StatusActive, AdminApproved: true, Role: role, ApprovalDate: approved}
	}
	cases := []struct {
		name     string
		actor    string
		user     model.User
		role     model.UserRole
		wantErr  errors.ErrorType
		wantRole model.UserRole
	}{
		{"pathologist to datascientist", "admin1", active(model.RolePathologist), model.RoleDatascientist, "", model.RoleDatascientist},
		{"datascientist to admin", "admin1", active(model.RoleDatascientist), model.RoleAdmin, "", model.RoleAdmin},
		{"admin demoted by another admin", "admin1", active(model.RoleAdmin), model.RolePathologist, "", model.RolePathologist},
		{"own role", "u1", active(model.RoleAdmin), model.RolePathologist, errors.ErrorTypeValidation, model.RoleAdmin},
		{"unassigned is not assignable", "admin1", active(model.RolePathologist), model.RoleUnassigned, errors.ErrorTypeValidation, model.RolePathologist},
		{"legacy name is not assignable", "admin1", active(model.RolePathologist), "viewer", errors.ErrorTypeValidation, model.RolePathologist},
		{"same role", "admin1", active(model.RolePathologist), model.RolePathologist, errors.ErrorTypeConflict, model.RolePathologist},
		{"pending user", "admin1", model.User{Status: model.StatusPending, Role: model.RoleUnassigned}, model.RolePathologist,
			errors.ErrorTypeConflict, model.RoleUnassigned},
		{"suspended user", "admin1", model.User{Status: model.StatusSuspended, Role: model.RolePathologist}, model.RoleDatascientist,
			errors.ErrorTypeConflict, model.RolePathologist},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := tc.user
			u.UserID = "u1"
			svc, repo := newRoleService(&u)

			err := svc.SetUserRole(context.Background(), tc.actor, "u1", tc.role)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected %s, got nil", tc.wantErr)
				}
				if got := errType(t, err); got != tc.wantErr {
					t.Fatalf("error type = %s, want %s", got, tc.wantErr)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := repo.users["u1"]
			if got.Role != tc.wantRole {
				t.Errorf("role = %q, want %q", got.Role, tc.wantRole)
			}
			// A role change is not an approval: status and its date stay put.
			if tc.wantErr == "" && (got.Status != model.StatusActive || !got.ApprovalDate.Equal(approved)) {
				t.Errorf("status/approval changed: status=%q date=%v", got.Status, got.ApprovalDate)
			}
		})
	}
}

func TestSetUserRoleUnknownUser(t *testing.T) {
	svc, _ := newRoleService()
	err := svc.SetUserRole(context.Background(), "admin1", "missing", model.RolePathologist)
	if err == nil || errType(t, err) != errors.ErrorTypeNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}
