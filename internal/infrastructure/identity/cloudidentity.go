// Package identity talks to the Cloud Identity Groups API.
//
// The service account this runs as must be a MANAGER (or OWNER) of the target
// group — that permission lives in Workspace/Cloud Identity, not in GCP IAM, so
// no project-level role grants it. On Cloud Run the credentials come from the
// attached service account via ADC; no key file is involved.
package identity

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"google.golang.org/api/cloudidentity/v1"
	"google.golang.org/api/googleapi"
)

// GroupDirectory manages members of one specific group.
type GroupDirectory struct {
	svc        *cloudidentity.Service
	groupEmail string

	// The API addresses groups by resource name ("groups/{id}"), which has to be
	// looked up from the email once and does not change afterwards. Only a
	// successful lookup is cached: a transient failure must not disable the
	// feature until the next restart.
	mu        sync.Mutex
	groupName string
}

func NewGroupDirectory(ctx context.Context, groupEmail string) (*GroupDirectory, error) {
	if groupEmail == "" {
		return nil, fmt.Errorf("identity: group email is not configured")
	}

	svc, err := cloudidentity.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("identity: failed to create Cloud Identity client: %w", err)
	}

	return &GroupDirectory{svc: svc, groupEmail: groupEmail}, nil
}

// resolveGroup turns the group email into its resource name, caching the result
// after the first success.
func (d *GroupDirectory) resolveGroup(ctx context.Context) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.groupName != "" {
		return d.groupName, nil
	}

	resp, err := d.svc.Groups.Lookup().
		GroupKeyId(d.groupEmail).
		Context(ctx).
		Do()
	if err != nil {
		return "", fmt.Errorf(
			"identity: cannot resolve group %q — check that it exists and that "+
				"this service account is a manager of it: %w", d.groupEmail, err)
	}

	d.groupName = resp.Name
	return d.groupName, nil
}

func (d *GroupDirectory) AddMember(ctx context.Context, email string) error {
	groupName, err := d.resolveGroup(ctx)
	if err != nil {
		return err
	}

	membership := &cloudidentity.Membership{
		PreferredMemberKey: &cloudidentity.EntityKey{Id: email},
		Roles:              []*cloudidentity.MembershipRole{{Name: "MEMBER"}},
	}

	_, err = d.svc.Groups.Memberships.Create(groupName, membership).Context(ctx).Do()
	if err != nil {
		// Already a member — the desired state is what the caller asked for.
		if isStatus(err, http.StatusConflict) {
			return nil
		}
		return fmt.Errorf("identity: failed to add %s to %s: %w", email, d.groupEmail, err)
	}
	return nil
}

func (d *GroupDirectory) RemoveMember(ctx context.Context, email string) error {
	membershipName, found, err := d.lookupMembership(ctx, email)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	if _, err := d.svc.Groups.Memberships.Delete(membershipName).Context(ctx).Do(); err != nil {
		if isStatus(err, http.StatusNotFound) {
			return nil
		}
		return fmt.Errorf("identity: failed to remove %s from %s: %w", email, d.groupEmail, err)
	}
	return nil
}

func (d *GroupDirectory) HasMember(ctx context.Context, email string) (bool, error) {
	_, found, err := d.lookupMembership(ctx, email)
	return found, err
}

func (d *GroupDirectory) lookupMembership(ctx context.Context, email string) (string, bool, error) {
	groupName, err := d.resolveGroup(ctx)
	if err != nil {
		return "", false, err
	}

	resp, err := d.svc.Groups.Memberships.Lookup(groupName).
		MemberKeyId(email).
		Context(ctx).
		Do()
	if err != nil {
		if isStatus(err, http.StatusNotFound) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("identity: failed to look up %s in %s: %w", email, d.groupEmail, err)
	}
	return resp.Name, true, nil
}

func isStatus(err error, code int) bool {
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		return apiErr.Code == code
	}
	return false
}
