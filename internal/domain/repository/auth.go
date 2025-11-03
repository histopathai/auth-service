package repository

import (
	"context"

	"github.com/histopathai/auth-service/internal/domain/model"
)

type AuthRepository interface {
	Register(ctx context.Context, payload *model.RegisterUser) (*model.User, error)

	VerifyIDToken(ctx context.Context, idToken string) (*model.User, error)

	ChangePassword(ctx context.Context, uid, newPassword string) error

	Delete(ctx context.Context, uid string) error

	GetAuthInfo(ctx context.Context, uid string) (*model.User, error)
}
