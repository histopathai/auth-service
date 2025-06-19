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
	Role        string `json:"role" binding:"required"` // Should be validated ("patolog", "admin", "viewer")
	Institution string `json:"institution,omitempty"`
}

// UserUpdateRequest is used for updating user profile
type UserUpdateRequest struct {
	DisplayName *string `json:"displayName,omitempty"`
	Role        *string `json:"role,omitempty"`
	Institution *string `json:"institution,omitempty"`
	IsActive    *bool   `json:"isActive,omitempty"`
}
