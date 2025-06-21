package adapters

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/histopathai/auth-service/internal/models"
)

// MockUserRepository is a mock implementation of the UserRepository interface
type MockUserRepository struct {
	mock.Mock
}

// NewMockUserRepository creates a new instance of MockUserRepository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Mock: mock.Mock{},
	}
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
