package dto

import "time"

// SessionResponse represents a single session
type SessionResponse struct {
	SessionID    string                 `json:"session_id" example:"abc123def456"`
	UserID       string                 `json:"user_id" example:"user-123"`
	CreatedAt    time.Time              `json:"created_at" example:"2023-10-15T14:30:00Z"`
	ExpiresAt    time.Time              `json:"expires_at" example:"2023-10-15T15:00:00Z"`
	LastUsedAt   time.Time              `json:"last_used_at" example:"2023-10-15T14:45:00Z"`
	RequestCount int64                  `json:"request_count" example:"42"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// CreateSessionResponse represents session creation response
type CreateSessionResponse struct {
	SessionID string          `json:"session_id" example:"abc123def456"`
	ExpiresAt time.Time       `json:"expires_at" example:"2023-10-15T15:00:00Z"`
	Message   string          `json:"message" example:"Session created successfully"`
	Session   SessionResponse `json:"session"`
}

// SessionListResponse represents user's active sessions
type SessionListResponse struct {
	ActiveSessions int               `json:"active_sessions" example:"3"`
	Sessions       []SessionResponse `json:"sessions"`
}

// SessionStatsResponse represents session statistics
type SessionStatsResponse struct {
	ActiveSessions int                    `json:"active_sessions" example:"3"`
	TotalRequests  int64                  `json:"total_requests" example:"150"`
	Sessions       []SessionDetailedStats `json:"sessions"`
	Summary        map[string]interface{} `json:"summary,omitempty"`
}

// SessionDetailedStats represents detailed statistics for a session
type SessionDetailedStats struct {
	SessionID    string                 `json:"session_id" example:"abc123def456"`
	CreatedAt    time.Time              `json:"created_at" example:"2023-10-15T14:30:00Z"`
	ExpiresAt    time.Time              `json:"expires_at" example:"2023-10-15T15:00:00Z"`
	LastUsedAt   time.Time              `json:"last_used_at" example:"2023-10-15T14:45:00Z"`
	RequestCount int64                  `json:"request_count" example:"42"`
	TimeLeft     string                 `json:"time_left" example:"15m30s"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// RevokeSessionResponse represents session revocation response
type RevokeSessionResponse struct {
	Message string `json:"message" example:"Session revoked successfully"`
}

// RevokeAllSessionsResponse represents bulk session revocation response
type RevokeAllSessionsResponse struct {
	Message         string `json:"message" example:"All sessions revoked successfully"`
	RevokedSessions int    `json:"revoked_sessions" example:"3"`
}

// ExtendSessionResponse represents session extension response
type ExtendSessionResponse struct {
	SessionID string    `json:"session_id" example:"abc123def456"`
	ExpiresAt time.Time `json:"expires_at" example:"2023-10-15T15:30:00Z"`
	Message   string    `json:"message" example:"Session extended successfully"`
}
