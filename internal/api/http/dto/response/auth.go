package response

// RegisterResponse represents user registration response
type RegisterResponse struct {
	User    UserResponse `json:"user"`
	Message string       `json:"message" example:"Registration successful. Please wait for admin approval."`
}

// VerifyTokenResponse represents token verification response
type VerifyTokenResponse struct {
	Valid bool         `json:"valid" example:"true"`
	User  UserResponse `json:"user"`
}

// ProfileResponse represents user profile response (same as UserResponse but can be extended)
type ProfileResponse struct {
	User UserResponse `json:"user"`
}
