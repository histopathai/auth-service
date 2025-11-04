package request

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email" example:"user@example.com"`
	Password    string `json:"password" binding:"required,min=8" example:"StrongP@ss123"`
	DisplayName string `json:"display_name" binding:"required,min=2,max=100" example:"John Doe"`
}

// VerifyTokenRequest represents token verification request
type VerifyTokenRequest struct {
	Token string `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8" example:"NewStrongP@ss123"`
}
