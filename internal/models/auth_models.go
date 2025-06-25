package models

// UserRegistrationPayload represents the payload for user registration.
type UserRegistrationPayload struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	DisplayName string `json:"displayName,omitempty"`
}

// UserLoginPayload represents the payload for user login.
type UserLoginPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// TokenResponse represents the response containing authentication tokens.
type TokenResponse struct {
	IDToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

// RefreshTokenPayload represents the payload for refreshing tokens.
type RefreshTokenPayload struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// SendVerificationEmailPayload represents the payload for sending a verification email.
// This is used to trigger the email verification process for a user.
type SendVerificationEmailPayload struct {
	Email string `json:"email" validate:"required,email"`
}

// PasswordResetRequestPayload represents the payload for requesting a password reset.
type PasswordResetRequestPayload struct {
	Email string `json:"email" validate:"required,email"`
}

// ConfirmPasswordResetPayload represents the payload for confirming a password reset.
type ConfirmPasswordResetPayload struct {
	OobCode     string `json:"oobCode" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=8"`
}

type UserAuthInfo struct {
	UID           string `json:"uid"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
	DisplayName   string `json:"displayName,omitempty"`
}
