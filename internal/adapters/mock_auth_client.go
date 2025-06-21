package adapters

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/histopathai/auth-service/internal/models"
	"github.com/histopathai/auth-service/internal/service"
)

// MockAuthClient is a mock implementation of the AuthClient interface
type MockAuthClient struct {
	mock.Mock
}

type MockAuthClientConfig struct {
	// Add any configuration fields needed for the mock
}

// NewMockAuthClient creates a new instance of MockAuthClient
func NewMockAuthClient(config MockAuthClientConfig) *MockAuthClient {
	return &MockAuthClient{
		Mock: mock.Mock{},
	}
}

func (m *MockAuthClient) VerifyIDToken(ctx context.Context, idToken string) (*models.TokenClaims, error) {
	args := m.Called(ctx, idToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TokenClaims), args.Error(1)
}

func (m *MockAuthClient) CreateUser(ctx context.Context, req *service.AuthClientCreateUserRequest) (*service.AuthClientUserRecord, error) {
	args := m.Called(ctx, req)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*service.AuthClientUserRecord), args.Error(1)
}

func (m *MockAuthClient) DeleteUser(ctx context.Context, uid string) error {
	args := m.Called(ctx, uid)
	return args.Error(0)
}
