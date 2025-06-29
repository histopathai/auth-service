package handlers

import (
	"log/slog"
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

// GetAllUsers
// @Summary Get All Users
// @Description Retrieves a list of all registered users (Admin access required)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} object{users=[]models.User,count=int} "Successfully retrieved users"
// @Failure 500 {object} object{error=string,message=string,details=string} "Failed to retrieve users"
// @Router /admin/users [get]
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

// GetUser
// @Summary Get User by UID
// @Description Retrieves a single user by their UID (Admin access required)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uid path string true "User UID"
// @Success 200 {object} object{user=models.User} "User retrieved successfully"
// @Failure 400 {object} object{error=string,message=string} "User UID is required"
// @Failure 500 {object} object{error=string,message=string,details=string} "Failed to retrieve the user"
// @Router /admin/users/{uid} [get]
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

// ApproveUser
// @Summary Approve User
// @Description Approves a pending user account and assigns a role (Admin access required)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uid path string true "User UID"
// @Param payload body object{role=models.UserRole} true "User Role"
// @Success 200 {object} object{message=string} "User approved successfully"
// @Failure 400 {object} object{error=string,message=string,details=string} "Invalid request payload or missing UID"
// @Failure 500 {object} object{error=string,message=string,details=string} "Failed to approve user"
// @Router /admin/users/{uid}/approve [post]
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

	err := h.authService.ApproveUser(c.Request.Context(), uid, req.Role) // Bu satırda hata dönüyor
	if err != nil {
		// BURAYI EKLEYİN: Hata detayını slog ile loglayın
		slog.Error("Failed to approve user in authService", "user_id", uid, "role", req.Role, "error", err)
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

// SuspendUser
// @Summary Suspend User
// @Description Suspends a user account (Admin access required)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uid path string true "User UID"
// @Success 200 {object} object{message=string} "User suspended successfully"
// @Failure 400 {object} object{error=string,message=string} "User UID is required"
// @Failure 500 {object} object{error=string,message=string,details=string} "Failed to suspend user"
// @Router /admin/users/{uid}/suspend [post]
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

// ChangePasswordForUser
// @Summary Change User Password (Admin)
// @Description Endpoint for an administrator to change a specific user's password.
// @Tags Admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uid path string true "User UID"
// @Param new_password body object{new_password=string} true "New password"
// @Success 200 {object} object{message=string} "User password changed successfully"
// @Failure 400 {object} object{error=string,message=string,details=string} "Invalid request payload or User ID missing"
// @Failure 401 {object} object{error=string,message=string} "Unauthorized"
// @Failure 403 {object} object{error=string,message=string} "Forbidden (Admin role required)"
// @Failure 500 {object} object{error=string,message=string,details=string} "Failed to change user password"
// @Router /admin/users/{uid}/password [put]
func (h *AdminHandler) ChangePasswordForUser(c *gin.Context) {
	type PasswordChangeRequest struct {
		NewPassword string `json:"new_password" binding:"required"`
	}

	// UID'yi URL parametresinden alıyoruz (admin başka bir kullanıcının şifresini değiştiriyor)
	uid := c.Param("uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "User ID is required from URL parameter",
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
			"message": "Failed to change user password",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User password changed successfully",
	})
}

// MakeAdmin
// @Summary Promote User to Admin
// @Description Promotes a user to the 'admin' role (Admin access required)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uid path string true "User UID"
// @Success 200 {object} object{message=string} "User successfully promoted to admin"
// @Failure 400 {object} object{error=string,message=string} "User UID is required"
// @Failure 403 {object} object{error=string,message=string} "Forbidden (Admin role required)"
// @Failure 500 {object} object{error=string,message=string,details=string} "Failed to promote user to admin"
// @Router /admin/users/{uid}/make-admin [post]
func (h *AdminHandler) MakeAdmin(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_uid",
			"message": "User UID is required",
		})
		return
	}

	err := h.authService.PromoteUserToAdmin(c.Request.Context(), uid)
	if err != nil {
		// Hata mesajına göre dönüş yapabiliriz, örneğin zaten admin ise 400 Bad Request
		if err.Error() == "user is already an admin" { // AuthServiceImpl'den gelen hataya göre
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "already_admin",
				"message": "User is already an admin",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_promote_user",
			"message": "An error occurred while promoting user to admin",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User successfully promoted to admin",
	})
}
