package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"
)

type ImageSession struct {
	UserID       string
	Role         string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	LastUsed     time.Time
	RequestCount int64
}

type ImageSessionService struct {
	sessions     map[string]*ImageSession
	userSessions map[string][]string // userID -> sessionIDs
	mutex        sync.RWMutex
	authService  AuthService
}

func NewImageSessionService(authService AuthService) *ImageSessionService {
	service := &ImageSessionService{
		sessions:     make(map[string]*ImageSession),
		userSessions: make(map[string][]string),
		authService:  authService,
	}

	// Auto cleanup every 5 minutes
	go service.cleanupExpiredSessions()

	return service
}

// CreateSession - Kullanıcı için yeni session oluştur
func (s *ImageSessionService) CreateSession(userID, role string) (string, error) {
	// Generate cryptographically secure session ID
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("session ID generation failed: %w", err)
	}
	sessionID := hex.EncodeToString(bytes) // 32 karakter

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Kullanıcının eski sessionlarını temizle (max 3 session per user)
	s.cleanupUserSessions(userID, 3)

	now := time.Now()
	session := &ImageSession{
		UserID:       userID,
		Role:         role,
		CreatedAt:    now,
		ExpiresAt:    now.Add(30 * time.Minute), // 30 dakika
		LastUsed:     now,
		RequestCount: 0,
	}

	s.sessions[sessionID] = session
	s.userSessions[userID] = append(s.userSessions[userID], sessionID)

	log.Printf("✅ Image session created: %s for user: %s", sessionID, userID)
	return sessionID, nil
}

// ValidateSession - Session'ı doğrula ve kullanım bilgisini güncelle
func (s *ImageSessionService) ValidateSession(sessionID string) (*ImageSession, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, false
	}

	// Expire kontrolü
	if time.Now().After(session.ExpiresAt) {
		s.removeSessionUnsafe(sessionID)
		return nil, false
	}

	// Kullanım bilgisini güncelle
	session.LastUsed = time.Now()
	session.RequestCount++

	return session, true
}

// ExtendSession - Session süresini uzat (aktif kullanımda)
func (s *ImageSessionService) ExtendSession(sessionID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	// 30 dakika daha uzat
	session.ExpiresAt = time.Now().Add(30 * time.Minute)
	return nil
}

// RevokeSession - Session'ı zorla iptal et
func (s *ImageSessionService) RevokeSession(sessionID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.removeSessionUnsafe(sessionID)
	return nil
}

// RevokeAllUserSessions - Kullanıcının tüm sessionlarını iptal et
func (s *ImageSessionService) RevokeAllUserSessions(userID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	sessionIDs := s.userSessions[userID]
	for _, sessionID := range sessionIDs {
		delete(s.sessions, sessionID)
	}
	delete(s.userSessions, userID)

	log.Printf("🗑️ All sessions revoked for user: %s", userID)
	return nil
}

// GetUserSessionStats - Kullanıcının session istatistikleri
func (s *ImageSessionService) GetUserSessionStats(userID string) map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	sessionIDs := s.userSessions[userID]
	stats := map[string]interface{}{
		"active_sessions": len(sessionIDs),
		"sessions":        []map[string]interface{}{},
	}

	for _, sessionID := range sessionIDs {
		if session, exists := s.sessions[sessionID]; exists {
			stats["sessions"] = append(stats["sessions"].([]map[string]interface{}), map[string]interface{}{
				"session_id":    sessionID,
				"created_at":    session.CreatedAt,
				"expires_at":    session.ExpiresAt,
				"last_used":     session.LastUsed,
				"request_count": session.RequestCount,
			})
		}
	}

	return stats
}

// Cleanup functions
func (s *ImageSessionService) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		now := time.Now()
		expiredCount := 0

		for sessionID, session := range s.sessions {
			if now.After(session.ExpiresAt) {
				s.removeSessionUnsafe(sessionID)
				expiredCount++
			}
		}

		if expiredCount > 0 {
			log.Printf("🧹 Cleaned up %d expired sessions", expiredCount)
		}
		s.mutex.Unlock()
	}
}

func (s *ImageSessionService) cleanupUserSessions(userID string, maxSessions int) {
	sessionIDs := s.userSessions[userID]
	if len(sessionIDs) < maxSessions {
		return
	}

	// En eski sessionları kaldır
	for i := 0; i < len(sessionIDs)-maxSessions+1; i++ {
		sessionID := sessionIDs[i]
		delete(s.sessions, sessionID)
		log.Printf("🗑️ Removed old session: %s for user: %s", sessionID, userID)
	}

	// Listeyi güncelle
	s.userSessions[userID] = sessionIDs[len(sessionIDs)-maxSessions+1:]
}

func (s *ImageSessionService) removeSessionUnsafe(sessionID string) {
	if session, exists := s.sessions[sessionID]; exists {
		userID := session.UserID
		delete(s.sessions, sessionID)

		// User sessions listesinden de kaldır
		sessionIDs := s.userSessions[userID]
		for i, id := range sessionIDs {
			if id == sessionID {
				s.userSessions[userID] = append(sessionIDs[:i], sessionIDs[i+1:]...)
				break
			}
		}

		if len(s.userSessions[userID]) == 0 {
			delete(s.userSessions, userID)
		}
	}
}
