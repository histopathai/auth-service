package models

// UserRegistrationPayload represents the payload for user registration.
type UserRegistrationPayload struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	DisplayName string `json:"displayName,omitempty"`
}

type UserAuthInfo struct {
	UID           string `json:"uid"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
	DisplayName   string `json:"displayName,omitempty"`
}
