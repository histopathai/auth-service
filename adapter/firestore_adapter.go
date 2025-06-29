package adapter

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/auth-service/internal/models"
	"github.com/histopathai/auth-service/internal/repository"
	"google.golang.org/api/iterator"
	// option paketini import edin
)

// Ensure FirestoreAdapter implements UserRepository interface
var _ repository.UserRepository = &FirestoreAdapter{}

// FirestoreAdapter implements UserRepository using Firestore as the backend.
type FirestoreAdapter struct {
	client     *firestore.Client
	collection *firestore.CollectionRef
}

// NewFirestoreAdapter creates a new FirestoreAdapter instance.
func NewFirestoreAdapter(userClient *firestore.Client, userCollection string) (*FirestoreAdapter, error) {

	return &FirestoreAdapter{
		client:     userClient,
		collection: userClient.Collection(userCollection),
	}, nil
}

func (fa *FirestoreAdapter) CreateUser(ctx context.Context, user *models.User) error {

	_, err := fa.collection.Doc(user.UID).Set(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (fa *FirestoreAdapter) GetUserByUID(ctx context.Context, uid string) (*models.User, error) {

	doc, err := fa.collection.Doc(uid).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by UID %s: %w", uid, err)
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, fmt.Errorf("failed to convert document data to User model: %w", err)
	}
	return &user, nil
}

func (fa *FirestoreAdapter) UpdateUser(ctx context.Context, uid string, payload *models.UpdateUserRequest) (*models.User, error) {

	updates := make([]firestore.Update, 0)

	if payload.DisplayName != nil {
		updates = append(updates, firestore.Update{Path: "displayName", Value: *payload.DisplayName})
	}
	if payload.Status != nil {
		updates = append(updates, firestore.Update{Path: "status", Value: *payload.Status})
	}
	if payload.Role != nil {
		updates = append(updates, firestore.Update{Path: "role", Value: *payload.Role})
	}

	if payload.AdminApproved != nil {
		updates = append(updates, firestore.Update{Path: "adminApproved", Value: *payload.AdminApproved})
		if *payload.AdminApproved {
			now := time.Now()
			updates = append(updates, firestore.Update{Path: "approvalDate", Value: now})
		} else {
			updates = append(updates, firestore.Update{Path: "approvalDate", Value: nil})
		}
	}

	if len(updates) == 0 {
		return fa.GetUserByUID(ctx, uid) // No updates, return current user
	}

	_, err := fa.collection.Doc(uid).Update(ctx, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update user %s: %w", uid, err)
	}

	return fa.GetUserByUID(ctx, uid)

}

func (fa *FirestoreAdapter) DeleteUser(ctx context.Context, uid string) error {

	_, err := fa.collection.Doc(uid).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete user %s: %w", uid, err)
	}
	return nil
}

func (fa *FirestoreAdapter) SetUserRoleAndStatus(ctx context.Context, uid string, role models.UserRole, status models.UserStatus, adminApproved bool) error {

	updates := []firestore.Update{
		{Path: "Role", Value: role},
		{Path: "Status", Value: status},
		{Path: "AdminApproved", Value: adminApproved},
	}

	if adminApproved {
		now := time.Now()
		updates = append(updates, firestore.Update{Path: "ApprovalDate", Value: now})
	} else {
		updates = append(updates, firestore.Update{Path: "ApprovalDate", Value: nil})
	}

	_, err := fa.collection.Doc(uid).Update(ctx, updates)
	if err != nil {
		return fmt.Errorf("failed to set role and status for user %s: %w", uid, err)
	}
	return nil
}

func (fa *FirestoreAdapter) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	iter := fa.collection.Documents(ctx)
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve users: %w", err)
		}

		var user models.User
		if err := doc.DataTo(&user); err != nil {
			return nil, fmt.Errorf("failed to convert document data to User model: %w", err)
		}
		users = append(users, &user)
	}
	return users, nil
}
