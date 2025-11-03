package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/internal/api/http/dto/request"
	"github.com/histopathai/auth-service/internal/domain/model"
	"github.com/histopathai/auth-service/internal/service"
	"github.com/histopathai/auth-service/internal/shared/errors"
)

type AuthHandler struct {
	authService service.AuthService
	BaseHandler
}

func NewAuthHandler(authService service.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		BaseHandler: BaseHandler{logger: logger, response: &ResponseHelper{}},
	}
}

// Register
// @Summary User Registration
// @Description Endpoint for user registration
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body request.RegisterUserRequest true "User registration payload"
// @Success 201 {object} response.SuccessResponse{data=models.User} "User registered successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "User registration failed"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, err)
		return
	}

	// DTO -> Service

	userRegister := &model.RegisterUser{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	}

	user, err := h.authService.RegisterUser(c.Request.Context(), userRegister)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.response.Success(c, http.StatusCreated, user)
}

// VerifyToken
// @Summary Verify Token
// @Description Endpoint to verify authentication token
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body request.VerifyTokenRequest true "Token verification payload"
// @Success 200 {object} response.SuccessResponse{data=string} "Token is valid"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Invalid or expired token"
// @Router /auth/verify [post]
func (h *AuthHandler) VerifyToken(c *gin.Context) {
	var req request.VerifyTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, err)
		return
	}

	user, err := h.authService.VerifyToken(c.Request.Context(), req.Token)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.response.Success(c, http.StatusOK, user)
}

// ChangePasswordSelf
// @Summary Change Own Password
// @Description Endpoint for a user to change their own password.
// @Tags Auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param payload body request.ChangePasswordSelfRequest true "Change password payload"
// @Success 204 {object} response.NoContent "Password changed successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Password change failed"
// @Router /auth/password [put]
func (h *AuthHandler) ChangePasswordSelf(c *gin.Context) {
	var req request.ChangePasswordSelfRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, err)
		return
	}

	userID, exist := c.Get("user_id")
	if !exist {
		h.handleError(c, errors.NewUnauthorizedError("User not authenticated"))
		return
	}

	err := h.authService.ChangeUserPassword(c.Request.Context(), userID.(string), req.NewPassword)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.response.NoContent(c)
}

// DeleteAccount
// @Summary Delete Own Account
// @Description Endpoint for a user to delete their own account.
// @Tags Auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 204 {object} response.NoContent "Account deleted successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Account deletion failed"
// @Router /user/account [delete]
func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		h.handleError(c, errors.NewUnauthorizedError("User not authenticated"))
		return
	}

	err := h.authService.DeleteUser(c.Request.Context(), userID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.response.NoContent(c)
}

// GetProfile
// @Summary Get User Profile
// @Description Endpoint to retrieve the authenticated user's profile.
// @Tags Auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.SuccessResponse{data=models.User} "User profile retrieved successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Failed to retrieve user profile"
// @Router /user/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		h.handleError(c, errors.NewUnauthorizedError("User not authenticated"))
		return
	}

	user, err := h.authService.GetUserByUID(c.Request.Context(), userID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.response.Success(c, http.StatusOK, user)
}
