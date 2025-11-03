package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/histopathai/auth-service/internal/domain/model"
	"github.com/histopathai/auth-service/internal/domain/repository"
	"github.com/histopathai/auth-service/internal/shared/errors"
)

const (
	DefaultSessionDuration = 30 * time.Minute
	MaxSessionsPerUser     = 3
)

type SessionService struct {
	sessionRepo repository.SessionRepository
	authService AuthService
	logger      *slog.Logger
}

func NewSessionService(sessionRepo repository.SessionRepository, authService AuthService, logger *slog.Logger) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		authService: authService,
		logger:      logger,
	}
}

func (s *SessionService) CreateSession(ctx context.Context, userID string) (string, error) {

	sessionID, err := s.generateSessionID(32)
	if err != nil {
		return "", errors.NewInternalError("failed to generate session ID", err)
	}

	now := time.Now()
	session := &model.Session{
		SessionID:    sessionID,
		UserID:       userID,
		CreatedAt:    now,
		ExpiresAt:    now.Add(DefaultSessionDuration),
		LastUsedAt:   now,
		RequestCount: 0,
		Metadata:     make(map[string]interface{}),
	}

	if err := s.enforceMaxSessions(ctx, userID, MaxSessionsPerUser); err != nil {
		return "", err
	}

	createdID, err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return createdID, nil
}

func (s *SessionService) ValidateSession(ctx context.Context, sessionID string) (*model.Session, error) {
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.sessionRepo.Delete(ctx, sessionID)
		return nil, errors.NewNotFoundError("session_expired")
	}
	session.LastUsedAt = time.Now()
	session.RequestCount++

	if err := s.sessionRepo.Update(ctx, sessionID, session); err != nil {
		s.logger.Warn("failed to update session usage", "sessionID", sessionID, "error", err)
	}

	return session, nil
}

func (s *SessionService) ExtendSession(ctx context.Context, sessionID string) error {
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	session.ExpiresAt = time.Now().Add(DefaultSessionDuration)

	if err := s.sessionRepo.Update(ctx, sessionID, session); err != nil {
		return errors.NewInternalError("failed to extend session", err)
	}

	return nil
}

func (s *SessionService) RevokeSession(ctx context.Context, sessionID string) error {
	if err := s.sessionRepo.Delete(ctx, sessionID); err != nil {
		return errors.NewInternalError("failed to revoke session", err)
	}
	return nil
}

func (s *SessionService) RevokeAllUserSessions(ctx context.Context, userID string) error {
	if err := s.sessionRepo.DeleteByUser(ctx, userID); err != nil {
		return errors.NewInternalError("failed to revoke user sessions", err)
	}
	return nil
}

func (s *SessionService) GetUserSessionStats(ctx context.Context, userID string) (map[string]interface{}, error) {
	sessions, err := s.sessionRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, errors.NewInternalError("failed to list user sessions", err)
	}

	stats := map[string]interface{}{
		"active_sessions": len(sessions),
		"sessions":        make([]map[string]interface{}, 0, len(sessions)),
	}

	sessionList := stats["sessions"].([]map[string]interface{})
	for _, session := range sessions {
		sessionList = append(sessionList, map[string]interface{}{
			"session_id":    session.SessionID,
			"created_at":    session.CreatedAt,
			"expires_at":    session.ExpiresAt,
			"last_used":     session.LastUsedAt,
			"request_count": session.RequestCount,
		})
	}
	stats["sessions"] = sessionList

	return stats, nil
}

func (s *SessionService) GetActiveSessionCount(ctx context.Context, userID string) (int, error) {
	sessions, err := s.sessionRepo.ListByUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	return len(sessions), nil
}

func (s *SessionService) generateSessionID(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *SessionService) enforceMaxSessions(ctx context.Context, userID string, maxSessions int) error {
	sessions, err := s.sessionRepo.ListByUser(ctx, userID)
	if err != nil {
		return err
	}
	if len(sessions) < maxSessions {
		return nil
	}

	toDelete := len(sessions) - maxSessions + 1
	oldestSessions := s.findOldestSessions(sessions, toDelete)
	for _, session := range oldestSessions {
		if err := s.sessionRepo.Delete(ctx, session.SessionID); err != nil {
			s.logger.Warn("failed to delete old session", "sessionID", session.SessionID, "error", err)
		}
	}

	return nil
}

func (s *SessionService) findOldestSessions(sessions []*model.Session, count int) []*model.Session {
	if len(sessions) <= count {
		return sessions
	}

	sorted := make([]*model.Session, len(sessions))
	copy(sorted, sessions)

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			time1 := sorted[i].LastUsedAt
			time2 := sorted[j].LastUsedAt

			if time1.IsZero() {
				time1 = sorted[i].CreatedAt
			}
			if time2.IsZero() {
				time2 = sorted[j].CreatedAt
			}

			if time1.After(time2) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted[:count]
}

func (s *SessionService) ValidateAndExtend(ctx context.Context, sessionID string) (*model.Session, error) {
	session, err := s.ValidateSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	timeLeft := time.Until(session.ExpiresAt)
	if timeLeft < DefaultSessionDuration/2 {
		if err := s.ExtendSession(ctx, sessionID); err != nil {
			s.logger.Warn("failed to auto-extend session", "sessionID", sessionID, "error", err)
		}
	}

	return session, nil
}
