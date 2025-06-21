package models

import (
	"time"
)

var ValidRoles = map[string]bool{
	"patolog": true,
	"admin":   true,
	"viewer":  true,
}

// User represents a user in the system.
type User struct {
	UID         string    `json:"uid" firestore:"uid"`
	Email       string    `json:"email" firestore:"email"`
	DisplayName string    `json:"displayName,omitempty" firestore:"displayName,omitempty"`
	Role        string    `json:"role" firestore:"role"`
	Institution string    `json:"institution,omitempty" firestore:"institution,omitempty"`
	CreatedAt   time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" firestore:"updatedAt"`
	IsActive    bool      `json:"isActive" firestore:"isActive"`
}

// UserCreateRequest is used for creating a new user
type UserCreateRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"displayName,omitempty"`
	Institution string `json:"institution,omitempty"`
	Role        string `json:"role,omitempty" `
}

// UserUpdateRequest is used for updating user profile
type UserUpdateRequest struct {
	DisplayName     *string `json:"displayName,omitempty"`
	Role            *string `json:"role,omitempty"`
	Institution     *string `json:"institution,omitempty"`
	IsActive        *bool   `json:"isActive,omitempty"`
	IsEmailVerified *bool   `json:"isEmailVerified,omitempty"` // Custom claim to check if email is verified
}

// TokenClaims represents the essential claims extracted fom a verified authentication token.
// This is a generic representation, abstracting away provider-specific token structures.
type TokenClaims struct {
	UID             string
	Email           string
	Role            string //Custom claim that we expect to be present for authorization
	IsEmailVerified bool   // Custom claim to check if email is verified
}
