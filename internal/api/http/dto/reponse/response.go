package response

import "time"

type ErrorResponse struct {
	ErrorType string      `json:"error" example:"invalid_request"`
	Message   string      `json:"message" example:"The request payload is invalid."`
	Details   interface{} `json:"details,omitempty" example:"Detailed error information."`
}

type PaginationResponse struct {
	Limit   int  `json:"limit" example:"20"`
	Offset  int  `json:"offset" example:"0"`
	HasMore bool `json:"has_more" example:"true"`
}

type UserResponse struct {
	UID           string    `json:"uid" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email         string    `json:"email" example:"user@example.com"`
	DisplayName   string    `json:"display_name" example:"John Doe"`
	Status        string    `json:"status" example:"active"`
	Role          string    `json:"role" example:"user"`
	AdminApproved bool      `json:"admin_approved" example:"true"`
	ApprovalDate  time.Time `json:"approval_date" example:"2023-10-01T12:00:00Z"`
	CreatedAt     time.Time `json:"created_at" example:"2023-09-01T12:00:00Z"`
	UpdatedAt     time.Time `json:"updated_at" example:"2023-09-15T12:00:00Z"`
}

type UserListResponse struct {
	Data       []UserResponse     `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}
