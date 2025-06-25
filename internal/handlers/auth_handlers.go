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

// InitatePasswordReset handles password reset initiation
func (h *AuthHandler) InitiatePasswordReset(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "User ID is required",
		})
		return
	}
	err := h.authService.InitiatePasswordReset(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "password_reset_failed",
			"message": "Failed to initiate password reset",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset initiated successfully",
	})
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	type PasswordChangeRequest struct {
		NewPassword string `json:"new_password" binding:"required"`
	}

	uid := c.Param("uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "User ID is required",
		})
		return
	}

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

// DeleteAccount handles user account deletion
func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "User ID is required",
		})
		return
	}

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

// GetProfile handles getting user profile
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
