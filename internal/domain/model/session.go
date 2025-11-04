package model

import "time"

type Session struct {
	SessionID    string
	UserID       string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	LastUsedAt   time.Time
	RequestCount int64
	Metadata     map[string]interface{}
}

func (s *Session) GetID() string {
	return s.SessionID
}

func (s *Session) SetID(id string) {
	s.SessionID = id
}

func (s *Session) SetCreatedAt(t time.Time) {
	s.CreatedAt = t
}

func (s *Session) SetUpdatedAt(t time.Time) {
	s.LastUsedAt = t
}
