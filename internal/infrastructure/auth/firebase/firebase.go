package firebase

import (
	"context"

	"firebase.google.com/go/auth"
	"github.com/histopathai/auth-service/internal/domain/model"
)

type FirebaseAuthRepositoryImpl struct {
	client *auth.Client
}

func NewFirebaseAuthRepository(client *auth.Client) *FirebaseAuthRepositoryImpl {
	return &FirebaseAuthRepositoryImpl{
		client: client,
	}
}

func (far *FirebaseAuthRepositoryImpl) VerifyIDToken(ctx context.Context, idToken string) (*model.UserAuthInfo, error) {
	token, err := far.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, MapFirebaseAuthError(err)
	}

	authUser := &model.UserAuthInfo{
		UserID:        token.UID,
		Email:         getStringClaim(token.Claims, "email"),
		EmailVerified: getBoolClaim(token.Claims, "email_verified"),
		DisplayName:   getStringClaim(token.Claims, "name"), // ✅ Güvenli
	}

	return authUser, nil
}

func (far *FirebaseAuthRepositoryImpl) ChangePassword(ctx context.Context, userID string, newPassword string) error {

	_, err := far.client.UpdateUser(ctx, userID, (&auth.UserToUpdate{}).Password(newPassword))
	if err != nil {
		return MapFirebaseAuthError(err)
	}

	return nil
}

func (far *FirebaseAuthRepositoryImpl) Delete(ctx context.Context, userID string) error {

	err := far.client.DeleteUser(ctx, userID)
	if err != nil {
		return MapFirebaseAuthError(err)
	}

	return nil

}

func (far *FirebaseAuthRepositoryImpl) GetAuthInfo(ctx context.Context, userID string) (*model.UserAuthInfo, error) {
	u, err := far.client.GetUser(ctx, userID)
	if err != nil {
		return nil, MapFirebaseAuthError(err)
	}

	authUser := &model.UserAuthInfo{
		UserID:        u.UID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		DisplayName:   u.DisplayName,
	}

	return authUser, nil
}

func getStringClaim(claims map[string]interface{}, key string) string {
	if val, ok := claims[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getBoolClaim(claims map[string]interface{}, key string) bool {
	if val, ok := claims[key]; ok && val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}
