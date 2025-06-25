package models

import "time"

type UserStatus string

const (
	StatusPending     UserStatus = "pending"
	StatusActive      UserStatus = "active"
	StatusSuspended   UserStatus = "suspended"
	StatusDeactivated UserStatus = "deactivated"
)

// User Role
type UserRole string

const (
	RoleAdmin      UserRole = "admin"
	RoleUser       UserRole = "user"
	RoleViewer     UserRole = "viewer"
	RoleUnassigned UserRole = "unassigned"
)

type User struct {
	UID           string     `json:"uid" validate:"required"`
	Email         string     `json:"email" validate:"required,email"`
	DisplayName   string     `json:"displayName,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	Status        UserStatus `json:"status"`
	Role          UserRole   `json:"role"`
	AdminApproved bool       `json:"adminApproved"`
	ApprovalDate  time.Time  `json:"approvalDate,omitempty"`
}

type UpdateUserRequest struct {
	DisplayName   *string     `json:"displayName,omitempty"`
	Status        *UserStatus `json:"status,omitempty"`
	Role          *UserRole   `json:"role,omitempty"`
	AdminApproved *bool       `json:"adminApproved,omitempty"`
	ApprovalDate  *time.Time  `json:"approvalDate,omitempty"`
}
