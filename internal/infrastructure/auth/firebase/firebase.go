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

func (far *FirebaseAuthRepositoryImpl) Register(ctx context.Context, payload *model.RegisterUser) (*model.UserAuthInfo, error) {
	params := (&auth.UserToCreate{}).
		Email(payload.Email).
		Password(payload.Password).
		DisplayName(payload.DisplayName).
		EmailVerified(false).
		Disabled(false)

	u, err := far.client.CreateUser(ctx, params)
	if err != nil {
		return nil, MapFirebaseAuthError(err)
	}

	UserAuthInfo := &model.UserAuthInfo{
		UID:           u.UID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		DisplayName:   u.DisplayName,
	}
	return UserAuthInfo, nil
}

func (far *FirebaseAuthRepositoryImpl) VerifyIDToken(ctx context.Context, idToken string) (*model.UserAuthInfo, error) {

	token, err := far.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, MapFirebaseAuthError(err)
	}

	authUser := &model.UserAuthInfo{
		UID:           token.UID,
		Email:         token.Claims["email"].(string),
		EmailVerified: token.Claims["email_verified"].(bool),
		DisplayName:   token.Claims["name"].(string),
	}

	return authUser, nil
}

func (far *FirebaseAuthRepositoryImpl) ChangePassword(ctx context.Context, uid string, newPassword string) error {

	_, err := far.client.UpdateUser(ctx, uid, (&auth.UserToUpdate{}).Password(newPassword))
	if err != nil {
		return MapFirebaseAuthError(err)
	}

	return nil
}

func (far *FirebaseAuthRepositoryImpl) Delete(ctx context.Context, uid string) error {

	err := far.client.DeleteUser(ctx, uid)
	if err != nil {
		return MapFirebaseAuthError(err)
	}

	return nil

}

func (far *FirebaseAuthRepositoryImpl) GetAuthInfo(ctx context.Context, uid string) (*model.UserAuthInfo, error) {
	u, err := far.client.GetUser(ctx, uid)
	if err != nil {
		return nil, MapFirebaseAuthError(err)
	}

	authUser := &model.UserAuthInfo{
		UID:           u.UID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		DisplayName:   u.DisplayName,
	}

	return authUser, nil
}
