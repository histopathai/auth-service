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
	RoleAdmin         UserRole = "admin"
	RolePathologist   UserRole = "pathologist"
	RoleDatascientist UserRole = "datascientist"
	RoleUnassigned    UserRole = "unassigned"
)

// Role names stored before the user groups existed. They are only ever read,
// through Normalize; nothing writes them any more.
const (
	legacyRoleUser   UserRole = "user"
	legacyRoleViewer UserRole = "viewer"
)

// Normalize maps a stored role to a current one, so that documents written
// before the role migration behave correctly without it having run.
func (r UserRole) Normalize() UserRole {
	switch r {
	case legacyRoleUser:
		return RolePathologist
	case legacyRoleViewer:
		return RoleDatascientist
	default:
		return r
	}
}

// IsAssignable reports whether an admin may give this role to a user.
// "unassigned" is where registration starts, not something to assign.
func (r UserRole) IsAssignable() bool {
	switch r {
	case RoleAdmin, RolePathologist, RoleDatascientist:
		return true
	default:
		return false
	}
}

type UpdateUser struct {
	DisplayName   *string
	Status        *UserStatus
	Role          *UserRole
	AdminApproved *bool
	ApprovalDate  *time.Time
	// DataAccess mirrors membership of the readers group. The group is the
	// source of truth; this field is a cached view of it for listing users.
	DataAccess   *bool
	DataAccessAt *time.Time
}

type User struct {
	UserID        string
	Email         string
	DisplayName   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Status        UserStatus
	Role          UserRole
	AdminApproved bool
	ApprovalDate  time.Time

	// DataAccess reports whether this user is in the readers group, which is
	// what grants read-only access to Firestore metadata and the processed
	// bucket from outside the platform (notebooks, dev-ingestor). It is
	// independent of Role: a platform "viewer" need not have data access, and
	// data access does not imply any platform privilege.
	DataAccess   bool
	DataAccessAt time.Time
}

func (u *User) GetID() string {
	return u.UserID
}

func (u *User) SetID(id string) {
	u.UserID = id
}

func (u *User) SetCreatedAt(t time.Time) {
	u.CreatedAt = t
}

func (u *User) SetUpdatedAt(t time.Time) {
	u.UpdatedAt = t
}
