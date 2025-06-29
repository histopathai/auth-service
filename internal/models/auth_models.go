package models

// UserRegistrationPayload represents the payload for user registration.
// @Description Represents the required payload for user registration.
type UserRegistrationPayload struct {
	// @Property Email string "The email address of the user" minLength(1) format(email)
	Email string `json:"email" validate:"required,email"`
	// @Property Password string "The password for the user account" minLength(8)
	Password string `json:"password" validate:"required,min=8"`
	// @Property DisplayName string "The display name of the user (optional)"
	DisplayName string `json:"displayName,omitempty"`
}

// UserAuthInfo represents basic user authentication information.
// @Description Represents basic user information returned after authentication.
type UserAuthInfo struct {
	// @Property UID string "The unique identifier of the user"
	UID string `json:"uid"`
	// @Property Email string "The email address of the user"
	Email string `json:"email"`
	// @Property EmailVerified bool "Indicates if the user's email address is verified"
	EmailVerified bool `json:"emailVerified"`
	// @Property DisplayName string "The display name of the user (optional)"
	DisplayName string `json:"displayName,omitempty"`
}
