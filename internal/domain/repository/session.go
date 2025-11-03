package repository

import (
	"context"

	"github.com/histopathai/auth-service/internal/domain/model"
)

type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) (string, error)
	Get(ctx context.Context, sessionID string) (*model.Session, error)
	Update(ctx context.Context, sessionID string, session *model.Session) error
	Delete(ctx context.Context, sessionID string) error
	DeleteByUser(ctx context.Context, userID string) error
	ListByUser(ctx context.Context, userID string) ([]*model.Session, error)
	GetStats() map[string]interface{}
}
