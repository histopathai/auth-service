package firestore

import (
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/auth-service/internal/domain/model"
)

func UserToFirestoreMap(user *model.User) map[string]interface{} {
	return map[string]interface{}{
		"UID":           user.UID,
		"Email":         user.Email,
		"DisplayName":   user.DisplayName,
		"CreatedAt":     user.CreatedAt,
		"UpdatedAt":     user.UpdatedAt,
		"Status":        string(user.Status),
		"Role":          string(user.Role),
		"AdminApproved": user.AdminApproved,
		"ApprovalDate":  user.ApprovalDate,
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
		updates = append(updates, firestore.Update{Path: "DisplayName", Value: *update.DisplayName})
	}
	if update.Status != nil {
		updates = append(updates, firestore.Update{Path: "Status", Value: string(*update.Status)})
	}
	if update.Role != nil {
		updates = append(updates, firestore.Update{Path: "Role", Value: string(*update.Role)})
	}
	if update.AdminApproved != nil {
		updates = append(updates, firestore.Update{Path: "AdminApproved", Value: *update.AdminApproved})
	}
	if update.ApprovalDate != nil {
		updates = append(updates, firestore.Update{Path: "ApprovalDate", Value: *update.ApprovalDate})
	}

	updates = append(updates, firestore.Update{Path: "UpdatedAt", Value: time.Now()})

	return updates
}
