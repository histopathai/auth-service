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

	for key, value := range doc.Data() {
		switch key {
		case "email":
			user.Email = value.(string)
		case "display_name":
			user.DisplayName = value.(string)
		case "created_at":
			user.CreatedAt = value.(time.Time)
		case "updated_at":
			user.UpdatedAt = value.(time.Time)
		case "status":
			user.Status = model.UserStatus(value.(string))
		case "role":
			user.Role = model.UserRole(value.(string))
		case "admin_approved":
			user.AdminApproved = value.(bool)
		case "approval_date":
			user.ApprovalDate = value.(time.Time)
		}
	}
	user.UserID = doc.Ref.ID
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
