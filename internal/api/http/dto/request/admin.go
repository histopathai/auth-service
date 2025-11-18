package request

// ListUsersRequest represents query parameters for listing users
type ListUsersRequest struct {
	PaginationRequest
}

// Default sort field for users
const DefaultUserSortBy = "created_at"

// ApplyDefaults sets default values for list users request
func (r *ListUsersRequest) ApplyDefaults() {
	r.PaginationRequest.ApplyDefaults(DefaultUserSortBy)
}

// GetAllowedSortFields returns allowed sort fields for users
func (r *ListUsersRequest) GetAllowedSortFields() []string {
	return []string{"created_at", "updated_at", "email", "display_name"}
}

// ChangeUserPasswordRequest represents admin password change request
type ChangeUserPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8" example:"NewStrongP@ss123"`
}
