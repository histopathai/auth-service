package repository

import "context"

// DirectoryRepository manages membership of the Google group that carries
// read-only access to the research data (Firestore metadata and the processed
// GCS bucket). Group membership is the only thing that grants that access —
// the IAM policy binds the group once and never changes per person.
type DirectoryRepository interface {
	// AddMember adds an email to the readers group. Adding an address that is
	// already a member is not an error.
	AddMember(ctx context.Context, email string) error

	// RemoveMember removes an email from the readers group. Removing an address
	// that is not a member is not an error.
	RemoveMember(ctx context.Context, email string) error

	// HasMember reports whether the email is currently a member.
	HasMember(ctx context.Context, email string) (bool, error)
}
