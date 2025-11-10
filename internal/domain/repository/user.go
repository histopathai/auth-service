package repository

import (
	"context"

	"github.com/histopathai/auth-service/internal/domain/model"
	"github.com/histopathai/auth-service/internal/shared/query"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error

	GetByUserID(ctx context.Context, userID string) (*model.User, error)

	Update(ctx context.Context, userID string, updates *model.UpdateUser) error

	Delete(ctx context.Context, userID string) error

	List(ctx context.Context, pagination *query.Pagination) (*query.Result[*model.User], error)
}
