package model

type ConfirmRegisterUser struct {
	Email       string
	Token       string
	DisplayName string
}

type UserAuthInfo struct {
	UserID        string
	Email         string
	EmailVerified bool
	DisplayName   string
}
