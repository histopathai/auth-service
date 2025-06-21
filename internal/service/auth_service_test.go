package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/histopathai/auth-service/internal/adapter"
	"github.com/histopathai/auth-service/internal/models"
	"github.com/histopathai/auth-service/internal/service"
)

// MockUserRepository is a mock implementation of the UserRepository interface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByUID(ctx context.Context, uid string) (*models.User, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *MockUserRepository) UpdateUser(ctx context.Context, uid string, updates map[string]interface{}) error {
	args := m.Called(ctx, uid, updates)
	return args.Error(0)
}
func (m *MockUserRepository) DeleteUser(ctx context.Context, uid string) error {
	args := m.Called(ctx, uid)
	return args.Error(0)
}
func (m *MockUserRepository) ListUsers(ctx context.Context) ([]*models.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}
func TestAuthService_VerifyIDToken(t *testing.T) {
	mockAuthClient := adapter.NewMockAuthClient(adapter.MockAuthClientConfig{})
	mockUserRepo := new(MockUserRepository)
	authService := service.NewAuthService(mockAuthClient, mockUserRepo)

	ctx := context.Background()

	t.Run("succesful token verification", func(t *testing.T) {
		expectedToken := &models.TokenClaims{
			UID:             "test-uid",
			Email:           "example@example.com",
			IsEmailVerified: true,
			Role:            "viewer",
		}

		mockAuthClient.On("VerifyIDToken", ctx, "some-token").Return(expectedToken, nil).Once()

		token, err := authService.VerifyIDToken(ctx, "Bearer some-token")

		assert.NoError(t, err)
		assert.Equal(t, expectedToken, token)
		mockAuthClient.AssertExpectations(t)
	})

	t.Run("failed token verification", func(t *testing.T) {
		mockAuthClient.On("VerifyIDToken", ctx, "invalid-token").Return(&models.TokenClaims{}, errors.New("invalid token")).Once()

		token, err := authService.VerifyIDToken(ctx, "invalid-token")
		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Contains(t, err.Error(), "invalid token")
		mockAuthClient.AssertExpectations(t)
	})
}
func TestAuthService_CreateUser(t *testing.T) {

	setupTest := func() (*adapter.MockAuthClient, *MockUserRepository, service.AuthService, context.Context) {
		mockAuthClient := adapter.NewMockAuthClient(adapter.MockAuthClientConfig{})
		mockUserRepo := new(MockUserRepository)
		authService := service.NewAuthService(mockAuthClient, mockUserRepo)
		ctx := context.Background()
		return mockAuthClient, mockUserRepo, authService, ctx
	}

	req := &models.UserCreateRequest{
		Email:       "newuser@example.com",
		Password:    "Password123",
		DisplayName: "New User",
		Role:        "viewer",
		Institution: "Test Institution",
	}

	t.Run("successful user creation", func(t *testing.T) {

		mockAuthClient, mockUserRepo, authService, ctx := setupTest()
		//Prepare the exoected input for AuthClient.CreateUser
		expectedAuthClientReq := &service.AuthClientCreateUserRequest{
			Email:       req.Email,
			Password:    req.Password,
			DisplayName: req.DisplayName,
		}

		// Prepare the expected output from AuthClient.CreateUser
		mockAuthClientUserRecord := &service.AuthClientUserRecord{
			UID:         "mock-uid-123",
			Email:       req.Email,
			DisplayName: req.DisplayName,
		}

		//Mock AuthClient.CreateUser call
		mockAuthClient.On("CreateUser", ctx, expectedAuthClientReq).
			Return(mockAuthClientUserRecord, nil).
			Once()

		// Mock UserRepository.CreateUser call.
		//Used mock.AnythingOfType to avoid passing the exact user object

		mockUserRepo.On("CreateUser", ctx, mock.AnythingOfType("*models.User")).
			Return(nil).
			Once()

		user, err := authService.CreateUser(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, mockAuthClientUserRecord.UID, user.UID) // Check if the UID matches
		assert.Equal(t, req.Email, user.Email)                  // Check if the email matches
		assert.NotZero(t, user.CreatedAt)                       // Check if CreatedAt is set
		assert.NotZero(t, user.UpdatedAt)                       // Check if UpdatedAt is set
		assert.False(t, user.IsActive)                          // Default isActive should be false
		assert.Equal(t, "viewer", user.Role)                    // Default role should be "viewer"
		mockAuthClient.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("auth provider user creation fails", func(t *testing.T) {
		mockAuthClient, mockUserRepo, authService, ctx := setupTest()

		expectedAuthClientReq := &service.AuthClientCreateUserRequest{
			Email:       req.Email,
			Password:    req.Password,
			DisplayName: req.DisplayName,
		}

		// Make sure to return nil for the user record when there's an error
		mockAuthClient.On("CreateUser", ctx, expectedAuthClientReq).
			Return((*service.AuthClientUserRecord)(nil), errors.New("auth provider error")).
			Once()

		user, err := authService.CreateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "failed to create user in auth provider: auth provider error")
		mockAuthClient.AssertExpectations(t)
		mockUserRepo.AssertNotCalled(t, "CreateUser", mock.Anything, mock.AnythingOfType("*models.User"))

	})

	t.Run("user repository creation fails, rollback auth provider", func(t *testing.T) {
		mockAuthClient, mockUserRepo, authService, ctx := setupTest()

		// Prepare the expected input for AuthClient.CreateUser
		expectedAuthClientReq := &service.AuthClientCreateUserRequest{
			Email:       req.Email,
			Password:    req.Password,
			DisplayName: req.DisplayName,
		}
		// Mock the AuthClient.CreateUser call
		mockAuthClientUserRecord := &service.AuthClientUserRecord{
			UID:         "mock-uid-rollback",
			Email:       req.Email,
			DisplayName: req.DisplayName,
		}

		// Mock the AuthClient.CreateUser call
		mockAuthClient.On("CreateUser", ctx, expectedAuthClientReq).
			Return(mockAuthClientUserRecord, nil).
			Once()

		// Mock the UserRepository.CreateUser call to return an error
		mockUserRepo.On("CreateUser", ctx, mock.AnythingOfType("*models.User")).
			Return(errors.New("user repository error")).
			Once()

		// Check if DeleteUser is called on the AuthClient
		mockAuthClient.On("DeleteUser", ctx, mockAuthClientUserRecord.UID).
			Return(nil).
			Once()

		user, err := authService.CreateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "failed to create user profile: user repository error")
		mockAuthClient.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)

	})

	t.Run("user repository creation fails, rollback auth provider also fails", func(t *testing.T) {
		mockAuthClient, mockUserRepo, authService, ctx := setupTest()

		// Prepare the expected input for AuthClient.CreateUser
		expectedAuthClientReq := &service.AuthClientCreateUserRequest{
			Email:       req.Email,
			Password:    "Password123",
			DisplayName: req.DisplayName,
		}
		// Mock the AuthClient.CreateUser call
		mockAuthClientUserRecord := &service.AuthClientUserRecord{
			UID:         "mock-uid-rollback-fail",
			Email:       req.Email,
			DisplayName: req.DisplayName,
		}

		mockAuthClient.On("CreateUser", ctx, expectedAuthClientReq).
			Return(mockAuthClientUserRecord, nil).
			Once()

		mockUserRepo.On("CreateUser", ctx, mock.AnythingOfType("*models.User")).
			Return(errors.New("user repository error")).
			Once()

		mockAuthClient.On("DeleteUser", ctx, mockAuthClientUserRecord.UID).
			Return(errors.New("auth provider rollback error")).
			Once()

		user, err := authService.CreateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "failed to create user profile: user repository error")

		mockAuthClient.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)

	})
}
func TestAuthService_UpdateUserRole(t *testing.T) {
	setupTest := func() (*adapter.MockAuthClient, *MockUserRepository, service.AuthService, context.Context) {
		mockAuthClient := adapter.NewMockAuthClient(adapter.MockAuthClientConfig{})
		mockUserRepo := new(MockUserRepository)
		authService := service.NewAuthService(mockAuthClient, mockUserRepo)
		ctx := context.Background()
		return mockAuthClient, mockUserRepo, authService, ctx
	}

	uid := "test-uid"
	newRole := "admin"

	t.Run("succesful role update", func(t *testing.T) {
		_, mockUserRepo, authService, ctx := setupTest()

		updates := map[string]interface{}{"Role": newRole}

		mockUserRepo.On("UpdateUser", mock.Anything, uid, updates).
			Return(nil).
			Once()

		err := authService.UpdateUserRole(ctx, uid, newRole)
		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("invalied role", func(t *testing.T) {
		invalidRole := "invalid_role"

		_, mockUserRepo, authService, ctx := setupTest()

		err := authService.UpdateUserRole(ctx, uid, invalidRole)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid role: invalid_role")
		mockUserRepo.AssertNotCalled(t, "UpdateUser", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("repository update fails", func(t *testing.T) {
		_, mockUserRepo, authService, ctx := setupTest()

		updates := map[string]interface{}{"Role": newRole}

		mockUserRepo.On("UpdateUser", mock.Anything, uid, updates).
			Return(errors.New("repository error")).
			Once()

		err := authService.UpdateUserRole(ctx, uid, newRole)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update user role in repository: repository error")
		mockUserRepo.AssertExpectations(t)
	})

}
