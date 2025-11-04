package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/histopathai/auth-service/internal/domain/model"
	"github.com/histopathai/auth-service/internal/shared/errors"
)

const (
	DefaultMaxSessionsPerUser = 5
	DefaultCleanupInterval    = 5 * time.Minute
)

type inMemorySessionRepository struct {
	sessions           map[string]*model.Session
	userSessions       map[string]map[string]bool
	mutex              sync.RWMutex
	cleanupOnce        sync.Once
	maxSessionsPerUser int
}

func NewInMemorySessionRepository(maxSessionsPerUser int) *inMemorySessionRepository {
	if maxSessionsPerUser <= 0 {
		maxSessionsPerUser = DefaultMaxSessionsPerUser
	}

	repo := &inMemorySessionRepository{
		sessions:           make(map[string]*model.Session),
		userSessions:       make(map[string]map[string]bool),
		maxSessionsPerUser: maxSessionsPerUser,
	}

	repo.cleanupOnce.Do(func() {
		go repo.cleanupExpiredSessions()
	})

	return repo
}

func (r *inMemorySessionRepository) Create(ctx context.Context, session *model.Session) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Session ID yoksa oluştur
	if session.SessionID == "" {
		session.SessionID = uuid.New().String()
	}

	sessionID := session.SessionID
	userID := session.UserID

	if r.userSessions[userID] == nil {
		r.userSessions[userID] = make(map[string]bool)
	}

	if len(r.userSessions[userID]) >= r.maxSessionsPerUser {
		oldestSessionID := r.findOldestSessionUnsafe(userID)
		if oldestSessionID != "" {
			r.deleteSessionUnsafe(oldestSessionID)
		}
	}

	r.sessions[sessionID] = session
	r.userSessions[userID][sessionID] = true

	return sessionID, nil
}

func (r *inMemorySessionRepository) Get(ctx context.Context, sessionID string) (*model.Session, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	session, exists := r.sessions[sessionID]
	if !exists {
		return nil, errors.NewNotFoundError("session_not_found")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, errors.NewNotFoundError("session_expired")
	}

	return session, nil
}

func (r *inMemorySessionRepository) Update(ctx context.Context, sessionID string, session *model.Session) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.sessions[sessionID]; !exists {
		return errors.NewNotFoundError("session_not_found")
	}

	session.SessionID = sessionID
	r.sessions[sessionID] = session
	return nil
}

func (r *inMemorySessionRepository) Delete(ctx context.Context, sessionID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.deleteSessionUnsafe(sessionID)
}

func (r *inMemorySessionRepository) deleteSessionUnsafe(sessionID string) error {
	session, exists := r.sessions[sessionID]
	if !exists {
		return errors.NewNotFoundError("session_not_found")
	}

	userID := session.UserID
	delete(r.sessions, sessionID)

	if userSessions, ok := r.userSessions[userID]; ok {
		delete(userSessions, sessionID)

		if len(userSessions) == 0 {
			delete(r.userSessions, userID)
		}
	}

	return nil
}

func (r *inMemorySessionRepository) DeleteByUser(ctx context.Context, userID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	userSessions, exists := r.userSessions[userID]
	if !exists {
		return nil
	}

	toDelete := make([]string, 0)
	for sessionID := range userSessions {
		toDelete = append(toDelete, sessionID)
	}

	for _, sessionID := range toDelete {
		r.deleteSessionUnsafe(sessionID)
	}

	return nil
}

func (r *inMemorySessionRepository) ListByUser(ctx context.Context, userID string) ([]*model.Session, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	userSessions, exists := r.userSessions[userID]
	if !exists {
		return []*model.Session{}, nil
	}

	sessions := make([]*model.Session, 0, len(userSessions))
	for sessionID := range userSessions {
		if session, ok := r.sessions[sessionID]; ok {
			if time.Now().Before(session.ExpiresAt) {
				sessions = append(sessions, session)
			}
		}
	}

	return sessions, nil
}

func (r *inMemorySessionRepository) findOldestSessionUnsafe(userID string) string {
	userSessions, exists := r.userSessions[userID]
	if !exists || len(userSessions) == 0 {
		return ""
	}

	var oldestSessionID string
	var oldestTime time.Time

	for sessionID := range userSessions {
		session, exists := r.sessions[sessionID]
		if !exists {
			continue
		}

		compareTime := session.LastUsedAt
		if compareTime.IsZero() {
			compareTime = session.CreatedAt
		}

		if oldestSessionID == "" || compareTime.Before(oldestTime) {
			oldestSessionID = sessionID
			oldestTime = compareTime
		}
	}

	return oldestSessionID
}

func (r *inMemorySessionRepository) cleanupExpiredSessions() {
	ticker := time.NewTicker(DefaultCleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		r.mutex.Lock()

		now := time.Now()
		toDelete := make([]string, 0)

		for sessionID, session := range r.sessions {
			if now.After(session.ExpiresAt) {
				toDelete = append(toDelete, sessionID)
			}
		}

		for _, sessionID := range toDelete {
			r.deleteSessionUnsafe(sessionID)
		}

		r.mutex.Unlock()
	}
}

func (r *inMemorySessionRepository) GetStats() map[string]interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return map[string]interface{}{
		"total_sessions":        len(r.sessions),
		"total_users":           len(r.userSessions),
		"max_sessions_per_user": r.maxSessionsPerUser,
	}
}
