package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/internal/models"
	"github.com/histopathai/auth-service/internal/service"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register
// @Summary User Registration
// @Description Endpoint for user registration
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body models.UserRegistrationPayload true "User registration payload"
// @Success 201 {object} object{user=models.User,message=string} "User registered successfully"
// @Failure 400 {object} object{error=string,message=string,details=string} "Invalid request payload"
// @Failure 500 {object} object{error=string,message=string,details=string} "User registration failed"
// @Router /auth/register [post]
// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var payload models.UserRegistrationPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	user, err := h.authService.RegisterUser(c.Request.Context(), &payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "registration_failed",
			"message": "Failed to register user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":    user,
		"message": "User registered successfully",
	})

}

// VerifyToken
// @Summary Token Verification
// @Description Endpoint to verify a user token
// @Tags Auth
// @Accept json
// @Produce json
// @Param token body object{token=string} true "Token to verify"
// @Success 200 {object} object{user=models.User,message=string} "Token verified successfully"
// @Failure 400 {object} object{error=string,message=string,details=string} "Invalid request"
// @Failure 401 {object} object{error=string,message=string,details=string} "Invalid token"
// @Router /auth/verify [post]
// VerifyToken handles token verification
func (h *AuthHandler) VerifyToken(c *gin.Context) {
	type TokenRequest struct {
		Token string `json:"token" binding:"required"`
	}
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Token is required",
			"details": err.Error(),
		})
		return
	}

	user, err := h.authService.VerifyToken(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid_token",
			"message": "Token verification failed",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"message": "Token verified successfully",
	})
}

// ChangePasswordSelf
// @Summary Change Own Password
// @Description Endpoint for an authenticated user to change their own password.
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param new_password body object{new_password=string} true "New password"
// @Success 200 {object} object{message=string} "Password changed successfully"
// @Failure 400 {object} object{error=string,message=string,details=string} "Invalid request payload"
// @Failure 401 {object} object{error=string,message=string} "User not authenticated"
// @Failure 500 {object} object{error=string,message=string,details=string} "Failed to change password"
// @Router /user/password [put]
func (h *AuthHandler) ChangePasswordSelf(c *gin.Context) {
	type PasswordChangeRequest struct {
		NewPassword string `json:"new_password" binding:"required"`
	}

	// UID'yi kimlik doğrulama bağlamından alıyoruz (kullanıcı kendi şifresini değiştiriyor)
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated or user ID not found in context",
		})
		return
	}

	loggedInUser, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to retrieve user information from context",
		})
		return
	}
	uid := loggedInUser.UID

	var req PasswordChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	err := h.authService.ChangePassword(c.Request.Context(), uid, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "password_change_failed",
			"message": "Failed to change password",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// DeleteAccount
// @Summary Delete User Account
// @Description Endpoint for a user to delete their own account
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} object{message=string} "User account deleted successfully"
// @Failure 400 {object} object{error=string,message=string} "Invalid request"
// @Failure 500 {object} object{error=string,message=string,details=string} "Failed to delete user account"
// @Router /user/account [delete]
// DeleteAccount handles user account deletion
// This endpoint allows a user to delete their account.
func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	// UID'yi URL parametresinden değil, kimlik doğrulama bağlamından alıyoruz
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated or user ID not found in context",
		})
		return
	}

	loggedInUser, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to retrieve user information from context",
		})
		return
	}
	uid := loggedInUser.UID // Kullanıcı objesinden UID'yi alın

	err := h.authService.DeleteUser(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "account_deletion_failed",
			"message": "Failed to delete user account",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User account deleted successfully",
	})
}

// GetProfile
// @Summary Get User Profile
// @Description Returns the profile information of the authenticated user
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} object{user=models.User} "User profile retrieved successfully"
// @Failure 401 {object} object{error=string,message=string} "User not authenticated"
// @Router /user/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})

}
