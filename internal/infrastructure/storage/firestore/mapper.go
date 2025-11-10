package firestore

import (
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/auth-service/internal/domain/model"
)

func UserToFirestoreMap(user *model.User) map[string]interface{} {
	return map[string]interface{}{
		"user_id":        user.UserID,
		"email":          user.Email,
		"display_name":   user.DisplayName,
		"created_at":     user.CreatedAt,
		"updated_at":     user.UpdatedAt,
		"status":         string(user.Status),
		"role":           string(user.Role),
		"admin_approved": user.AdminApproved,
		"approval_date":  user.ApprovalDate,
	}
}

func UserFromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.User, error) {
	var user model.User

	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func UpdateUserToFirestoreUpdates(update *model.UpdateUser) []firestore.Update {
	updates := make([]firestore.Update, 0)

	if update.DisplayName != nil {
		updates = append(updates, firestore.Update{Path: "display_name", Value: *update.DisplayName})
	}
	if update.Status != nil {
		updates = append(updates, firestore.Update{Path: "status", Value: string(*update.Status)})
	}
	if update.Role != nil {
		updates = append(updates, firestore.Update{Path: "role", Value: string(*update.Role)})
	}
	if update.AdminApproved != nil {
		updates = append(updates, firestore.Update{Path: "admin_approved", Value: *update.AdminApproved})
	}
	if update.ApprovalDate != nil {
		updates = append(updates, firestore.Update{Path: "approval_date", Value: *update.ApprovalDate})
	}

	updates = append(updates, firestore.Update{Path: "updated_at", Value: time.Now()})

	return updates
}
