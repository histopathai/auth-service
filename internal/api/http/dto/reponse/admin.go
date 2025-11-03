package dto

// UserListResponse represents paginated user list response
type UserListResponse struct {
	Data       []UserResponse     `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// UserDetailResponse represents detailed user response
type UserDetailResponse struct {
	UserResponse
	LastLoginAt    *string `json:"last_login_at,omitempty" example:"2023-10-15T14:30:00Z"`
	FailedAttempts int     `json:"failed_attempts,omitempty" example:"0"`
}

// UserActionResponse represents response after user action (approve, suspend, etc.)
type UserActionResponse struct {
	Message string       `json:"message" example:"User approved successfully"`
	User    UserResponse `json:"user"`
}
