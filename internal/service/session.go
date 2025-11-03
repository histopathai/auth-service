package service

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"sync"
	"time"

	"github.com/histopathai/auth-service/internal/shared/errors"
)

type SessionScope string

const (
	ScopeImageServe SessionScope = "image-serve"
	ScopeAdminOps   SessionScope = "admin-ops"
)

type Session struct {
	Scope        SessionScope
	UserID       string
	Role         string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	LastUsed     time.Time
	RequestCount int64
	Metadata     map[string]interface{}
}

type ScopeConfig struct {
	Expiration        time.Duration
	MaxSessionPerUser int
	AllowedRoles      []string
}

type Config struct {
	DefaultExpiration  time.Duration
	MaxSessionsPerUser int
	CleanupInterval    time.Duration
	ScopeConfigs       map[SessionScope]ScopeConfig
}

type ScopedSessionService struct {
	sessions     map[SessionScope]map[string]*Session
	userSessions map[SessionScope]map[string][]string // userID -> sessionIDs
	mutex        sync.RWMutex
	logger       *slog.Logger
	config       Config
}

func NewScopeSessionService(config Config, logger *slog.Logger) *ScopedSessionService {

	if config.CleanupInterval <= 0 {
		config.CleanupInterval = 5 * time.Minute
	}

	if config.DefaultExpiration <= 0 {
		config.DefaultExpiration = 30 * time.Minute
	}

	if config.MaxSessionsPerUser <= 0 {
		config.MaxSessionsPerUser = 3
	}
	if config.ScopeConfigs == nil {
		config.ScopeConfigs = make(map[SessionScope]ScopeConfig)
	}

	service := &ScopedSessionService{
		sessions:     make(map[SessionScope]map[string]*Session),
		userSessions: make(map[SessionScope]map[string][]string),
		logger:       logger,
		config:       config,
	}

	go service.cleanupExpiredSessions()
	return service
}

func (s *ScopedSessionService) CreateSession(scope SessionScope, userID, role string, metadata map[string]interface{}) (string, *Session, error) {
	scopeConfig, hasConfig := s.config.ScopeConfigs[scope]
	expiration := s.config.DefaultExpiration
	maxSessions := s.config.MaxSessionsPerUser

	if hasConfig {
		if scopeConfig.Expiration > 0 {
			expiration = scopeConfig.Expiration
		}
		if scopeConfig.MaxSessionPerUser > 0 {
			maxSessions = scopeConfig.MaxSessionPerUser
		}

		// Role control(optional)
		if len(scopeConfig.AllowedRoles) > 0 && !contains(scopeConfig.AllowedRoles, role) {
			return "", nil, errors.NewForbiddenError("role not allowed for this scope")
		}
	}

	// Generate session ID
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", nil, errors.NewInternalError("session_id_generation_failed", err)
	}

	sessionID := hex.EncodeToString(bytes) // 32 characters

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.sessions[scope] == nil {
		s.sessions[scope] = make(map[string]*Session)
	}
	if s.userSessions[scope] == nil {
		s.userSessions[scope] = make(map[string][]string)
	}

	s.cleanupUserSessionsUnsafe(scope, userID, maxSessions)

	now := time.Now()
	session := &Session{
		Scope:        scope,
		UserID:       userID,
		Role:         role,
		CreatedAt:    now,
		ExpiresAt:    now.Add(expiration),
		LastUsed:     now,
		RequestCount: 0,
		Metadata:     metadata,
	}

	s.sessions[scope][sessionID] = session
	s.userSessions[scope][userID] = append(s.userSessions[scope][userID], sessionID)

	s.logger.Info("Session created",
		"scope", scope,
		"user_id", userID,
		"session_id", sessionID,
		"expires_at", expiration,
	)

	return sessionID, session, nil
}

func (s *ScopedSessionService) ValidateSession(scope SessionScope, sessionID string) (*Session, bool) {

	s.mutex.Lock()
	defer s.mutex.Unlock()

	scopeSessions, exists := s.sessions[scope]
	if !exists {
		return nil, false
	}

	session, exists := scopeSessions[sessionID]
	if !exists {
		return nil, false
	}

	if time.Now().After(session.ExpiresAt) {
		s.removeSessionUnsafe(scope, sessionID)
		return nil, false
	}

	session.LastUsed = time.Now()
	session.RequestCount++
	return session, true
}

func (s *ScopedSessionService) ExtendSession(scope SessionScope, sessionID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	scopeSessions, exists := s.sessions[scope]
	if !exists {
		return errors.NewNotFoundError("session not found")
	}

	session, exists := scopeSessions[sessionID]
	if !exists {
		return errors.NewNotFoundError("session not found")
	}

	scopeConfig, hasConfig := s.config.ScopeConfigs[scope]
	expiration := s.config.DefaultExpiration
	if hasConfig && scopeConfig.Expiration > 0 {
		expiration = scopeConfig.Expiration
	}

	session.ExpiresAt = time.Now().Add(expiration)
	return nil
}

func (s *ScopedSessionService) RevokeSession(scope SessionScope, sessionID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.removeSessionUnsafe(scope, sessionID)
	return nil
}

func (s *ScopedSessionService) RevokeAllUserSessions(scope SessionScope, userID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.userSessions[scope] == nil {
		return nil
	}

	sessionIDs := s.userSessions[scope][userID]
	for _, sessionID := range sessionIDs {
		delete(s.sessions[scope], sessionID)
	}
	delete(s.userSessions[scope], userID)
	return nil
}

func (s *ScopedSessionService) GetUserSessionStats(scope SessionScope, userID string) map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.userSessions[scope] == nil {
		return map[string]interface{}{
			"scope":           scope,
			"active_sessions": 0,
			"sessions":        []map[string]interface{}{},
		}
	}

	sessionIDs := s.userSessions[scope][userID]
	stats := map[string]interface{}{
		"scope":           scope,
		"active_sessions": len(sessionIDs),
		"sessions":        []map[string]interface{}{},
	}

	for _, sessionID := range sessionIDs {
		if session, exists := s.sessions[scope][sessionID]; exists {
			stats["sessions"] = append(stats["sessions"].([]map[string]interface{}), map[string]interface{}{
				"session_id":    sessionID,
				"created_at":    session.CreatedAt,
				"expires_at":    session.ExpiresAt,
				"last_used":     session.LastUsed,
				"request_count": session.RequestCount,
				"metadata":      session.Metadata,
			})
		}
	}
	return stats
}

func (s *ScopedSessionService) GetAllUserStats(userID string) map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	allStats := map[string]interface{}{
		"user_id": userID,
		"scopes":  []map[string]interface{}{},
	}

	for scope := range s.sessions {
		if s.userSessions[scope] == nil {
			continue
		}
		sessionIDs := s.userSessions[scope][userID]
		if len(sessionIDs) == 0 {
			continue
		}

		scopeStats := map[string]interface{}{
			"scope":           scope,
			"active_sessions": len(sessionIDs),
			"sessions":        []map[string]interface{}{},
		}

		for _, sessionID := range sessionIDs {
			if session, exists := s.sessions[scope][sessionID]; exists {
				scopeStats["sessions"] = append(scopeStats["sessions"].([]map[string]interface{}), map[string]interface{}{
					"session_id":    sessionID,
					"created_at":    session.CreatedAt,
					"expires_at":    session.ExpiresAt,
					"last_used":     session.LastUsed,
					"request_count": session.RequestCount,
				})
			}
		}
		allStats["scopes"] = append(allStats["scopes"].([]map[string]interface{}), scopeStats)
	}

	return allStats
}

// --- Internal helpers ---
func (s *ScopedSessionService) cleanupExpiredSessions() {
	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		now := time.Now()
		totalExpired := 0

		for scope, scopeSessions := range s.sessions {
			expiredCount := 0
			for sessionID, session := range scopeSessions {
				if now.After(session.ExpiresAt) {
					s.removeSessionUnsafe(scope, sessionID)
					expiredCount++
				}
			}
			totalExpired += expiredCount
			if expiredCount > 0 {
				s.logger.Info("Cleaned up expired sessions",
					"scope", scope,
					"count", expiredCount,
				)
			}
		}
		s.mutex.Unlock()
	}
}

func (s *ScopedSessionService) cleanupUserSessionsUnsafe(scope SessionScope, userID string, maxSessions int) {
	sessionIDs := s.userSessions[scope][userID]
	if len(sessionIDs) < maxSessions {
		return
	}

	sessionsToRemove := len(sessionIDs) - maxSessions + 1
	if sessionsToRemove <= 0 {
		return
	}

	for i := 0; i < sessionsToRemove; i++ {
		sessionID := sessionIDs[i]
		delete(s.sessions[scope], sessionID)
		s.logger.Info("Removed old session",
			"scope", scope,
			"session_id", sessionID,
			"user_id", userID,
		)
	}
	s.userSessions[scope][userID] = sessionIDs[sessionsToRemove:]
}

func (s *ScopedSessionService) removeSessionUnsafe(scope SessionScope, sessionID string) {
	scopeSessions, exists := s.sessions[scope]
	if !exists {
		return
	}

	session, exists := scopeSessions[sessionID]
	if !exists {
		return
	}

	userID := session.UserID
	delete(s.sessions[scope], sessionID)

	sessionIDs := s.userSessions[scope][userID]
	for i, id := range sessionIDs {
		if id == sessionID {
			s.userSessions[scope][userID] = append(sessionIDs[:i], sessionIDs[i+1:]...)
			break
		}
	}

	if len(s.userSessions[scope][userID]) == 0 {
		delete(s.userSessions[scope], userID)
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
