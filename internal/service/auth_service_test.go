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
