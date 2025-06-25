package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/internal/models"
	"github.com/histopathai/auth-service/internal/service"
)

// AdminHandler handles administrative tasks
type AdminHandler struct {
	authService service.AuthService
}

// NewAdminHandler creates a new AdminHandler instance
func NewAdminHandler(authService service.AuthService) *AdminHandler {
	return &AdminHandler{
		authService: authService,
	}
}

// GetAllUsers handles the retrieval of all users
func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	users, err := h.authService.GetAllUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_retrieve_users",
			"message": "An error occurred while retrieving users",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}

// GetUser handles the retrieval of a single user by UID
func (h *AdminHandler) GetUser(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_uid",
			"message": "User UID is required",
		})
		return
	}

	user, err := h.authService.GetUser(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_retrieve_user",
			"message": "An error occurred while retrieving the user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// ApproveUser handles the approval of a user by UID
func (h *AdminHandler) ApproveUser(c *gin.Context) {
	type ApproveUserRequest struct {
		Role models.UserRole `json:"role" binding:"required"`
	}

	uid := c.Param("uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_uid",
			"message": "User UID is required",
		})
		return
	}
	var req ApproveUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request payload",
			"details": err.Error(),
		})
		return
	}
	err := h.authService.ApproveUser(c.Request.Context(), uid, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_approve_user",
			"message": "An error occurred while approving the user",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User approved successfully",
	})
}

// SuspendUser handles the suspension(admin only)
func (h *AdminHandler) SuspendUser(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_uid",
			"message": "User UID is required",
		})
		return
	}

	err := h.authService.SuspendUser(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_suspend_user",
			"message": "An error occurred while suspending the user",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User suspended successfully",
	})
}

// DeactivateUser handles the deactivation of a user
func (h *AdminHandler) DeactivateUser(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_uid",
			"message": "User UID is required",
		})
		return
	}

	err := h.authService.DeactivateUser(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_deactivate_user",
			"message": "An error occurred while deactivating the user",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User deactivated successfully",
	})
}

// InitiateEmailVerification handles the initiation of email verification for a user
func (h *AdminHandler) InitiateEmailVerification(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_uid",
			"message": "User UID is required",
		})
		return
	}

	err := h.authService.InitiateEmailVerification(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_initiate_email_verification",
			"message": "An error occurred while initiating email verification",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Email verification initiated successfully",
	})
}

// ActivateUser handles the activation of a user account
func (h *AdminHandler) ActivateUser(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_uid",
			"message": "User UID is required",
		})
		return
	}

	err := h.authService.ActivateUser(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_activate_user",
			"message": "An error occurred while activating the user",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User activated successfully",
	})
}
