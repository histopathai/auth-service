package dto

// CreateSessionRequest represents session creation request
type CreateSessionRequest struct {
	Token string `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// ExtendSessionRequest represents session extension request (optional, can use path param only)
type ExtendSessionRequest struct {
	SessionID string `json:"session_id" binding:"required" example:"abc123def456"`
}

// RevokeSessionRequest represents session revocation request (optional, can use path param only)
type RevokeSessionRequest struct {
	SessionID string `json:"session_id" binding:"required" example:"abc123def456"`
}
