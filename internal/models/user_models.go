package models

import "time"

// UserStatus is an enum for user account status.
// @Enum pending active suspended
type UserStatus string

const (
	StatusPending   UserStatus = "pending"
	StatusActive    UserStatus = "active"
	StatusSuspended UserStatus = "suspended"
)

// UserRole is an enum for user roles.
// @Enum admin user viewer unassigned
type UserRole string

const (
	RoleAdmin      UserRole = "admin"
	RoleUser       UserRole = "user"
	RoleViewer     UserRole = "viewer"
	RoleUnassigned UserRole = "unassigned"
)

// User represents the full user profile stored in the database.
// @Description Represents a user's full profile stored in the database.
type User struct {
	// @Property UID string "The unique identifier of the user"
	UID string `json:"uid" validate:"required"`
	// @Property Email string "The email address of the user"
	Email string `json:"email" validate:"required,email"`
	// @Property DisplayName string "The display name of the user (optional)"
	DisplayName string `json:"displayName,omitempty"`
	// @Property CreatedAt time.Time "Timestamp when the user account was created"
	CreatedAt time.Time `json:"createdAt"`
	// @Property UpdatedAt time.Time "Timestamp when the user account was last updated"
	UpdatedAt time.Time `json:"updatedAt"`
	// @Property Status UserStatus "The status of the user account"
	Status UserStatus `json:"status"`
	// @Property Role UserRole "The assigned role of the user"
	Role UserRole `json:"role"`
	// @Property AdminApproved bool "Indicates if the user was approved by an admin"
	AdminApproved bool `json:"adminApproved"`
	// @Property ApprovalDate time.Time "Timestamp when the user account was approved (optional)"
	ApprovalDate time.Time `json:"approvalDate,omitempty"`
}

// UpdateUserRequest represents the payload for updating user information.
// @Description Payload for updating existing user information (only specified fields are updated)
type UpdateUserRequest struct {
	// @Property DisplayName *string "The display name of the user (optional)"
	DisplayName *string `json:"displayName,omitempty"`
	// @Property Status *UserStatus "The status of the user account (optional)"
	Status *UserStatus `json:"status,omitempty"`
	// @Property Role *UserRole "The assigned role of the user (optional)"
	Role *UserRole `json:"role,omitempty"`
	// @Property AdminApproved *bool "Indicates if the user was approved by an admin (optional)"
	AdminApproved *bool `json:"adminApproved,omitempty"`
	// @Property ApprovalDate *time.Time "Timestamp when the user account was approved (optional)"
	ApprovalDate *time.Time `json:"approvalDate,omitempty"`
}
