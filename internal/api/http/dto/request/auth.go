package request

// ConfirmRegisterRequest represents user registration confirmation request
type ConfirmRegisterRequest struct {
	Email       string `json:"email" binding:"required,email" example:"user@example.com"`
	Token       string `json:"token" binding:"required" example:"confirmation_token"`
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
