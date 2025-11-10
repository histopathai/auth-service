package model

type RegisterUser struct {
	Email       string
	Password    string
	DisplayName string
}

type UserAuthInfo struct {
	UserID        string
	Email         string
	EmailVerified bool
	DisplayName   string
}
