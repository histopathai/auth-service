package model

import "time"

type UserStatus string

const (
	StatusPending   UserStatus = "pending"
	StatusActive    UserStatus = "active"
	StatusSuspended UserStatus = "suspended"
)

type UserRole string

const (
	RoleAdmin      UserRole = "admin"
	RoleUser       UserRole = "user"
	RoleViewer     UserRole = "viewer"
	RoleUnassigned UserRole = "unassigned"
)

type UpdateUser struct {
	DisplayName   *string
	Status        *UserStatus
	Role          *UserRole
	AdminApproved *bool
	ApprovalDate  *time.Time
}

type User struct {
	UID           string
	Email         string
	DisplayName   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Status        UserStatus
	Role          UserRole
	AdminApproved bool
	ApprovalDate  time.Time
}

func (u *User) GetID() string {
	return u.UID
}

func (u *User) SetID(id string) {
	u.UID = id
}

func (u *User) SetCreatedAt(t time.Time) {
	u.CreatedAt = t
}

func (u *User) SetUpdatedAt(t time.Time) {
	u.UpdatedAt = t
}
