package adapters

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/auth-service/internal/models"
	"github.com/histopathai/auth-service/internal/repository"
	"google.golang.org/api/iterator"
)

// FirestoreUserRepository implements UserRepository for Firestore.
type FirestoreUserRepository struct {
	client *firestore.Client
}

func NewFirestoreUserRepository(client *firestore.Client) repository.UserRepository {
	return &FirestoreUserRepository{client: client}
}

func (r *FirestoreUserRepository) GetUserByUID(ctx context.Context, uid string) (*models.User, error) {
	docRef := r.client.Collection("users").Doc(uid)

	docSnap, err := docRef.Get(ctx)
	if err != nil {
		if err == iterator.Done { // Not found case for single document
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	var user models.User
	if err := docSnap.DataTo(&user); err != nil {
		return nil, fmt.Errorf("failed to convert user data: %w", err)
	}
	return &user, nil
}

func (r *FirestoreUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	user.UpdatedAt = time.Now()

	_, err := r.client.Collection("users").Doc(user.UID).Set(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *FirestoreUserRepository) UpdateUser(ctx context.Context, uid string, updates map[string]interface{}) error {
	updates["UpdatedAt"] = time.Now() // Update timestamp

	_, err := r.client.Collection("users").Doc(uid).Set(ctx, updates, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *FirestoreUserRepository) DeleteUser(ctx context.Context, uid string) error {
	_, err := r.client.Collection("users").Doc(uid).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (r *FirestoreUserRepository) ListUsers(ctx context.Context) ([]*models.User, error) {
	var users []*models.User

	iter := r.client.Collection("users").Documents(ctx)
	defer iter.Stop()

	for {
		docSnap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list users: %w", err)
		}

		var user models.User
		if err := docSnap.DataTo(&user); err != nil {
			return nil, fmt.Errorf("failed to convert user data: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}
