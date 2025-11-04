package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	dtoRequest "github.com/histopathai/auth-service/internal/api/http/dto/request"
	dtoResponse "github.com/histopathai/auth-service/internal/api/http/dto/response"
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
// @Description Register a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.RegisterRequest true "Registration details"
// @Success 201 {object} dto.RegisterResponse "User registered successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 409 {object} dto.ErrorResponse "Email already exists"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dtoRequest.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.NewValidationError("Invalid request payload", nil))
		return
	}

	user, err := h.authService.RegisterUser(c.Request.Context(), &model.RegisterUser{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := dtoResponse.RegisterResponse{
		User:    mapToUserResponse(user),
		Message: "User registered successfully",
	}

	h.response.Success(c, http.StatusCreated, response)
}

// VerifyToken
// @Summary Verify Token
// @Description Verify authentication token validity
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.VerifyTokenRequest true "Token to verify"
// @Success 200 {object} dto.VerifyTokenResponse "Token is valid"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Invalid or expired token"
// @Router /auth/verify [post]
func (h *AuthHandler) VerifyToken(c *gin.Context) {
	var req dtoRequest.VerifyTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.NewValidationError("Invalid request payload", nil))
		return
	}

	user, err := h.authService.VerifyToken(c.Request.Context(), req.Token)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := dtoResponse.VerifyTokenResponse{
		Valid: true,
		User:  mapToUserResponse(user),
	}

	h.response.Success(c, http.StatusOK, response)
}

// ChangePasswordSelf
// @Summary Change Own Password
// @Description Change authenticated user's password
// @Tags Auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param payload body dto.ChangePasswordRequest true "New password"
// @Success 204 "Password changed successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/password [put]
func (h *AuthHandler) ChangePasswordSelf(c *gin.Context) {
	var req dtoRequest.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.NewValidationError("Invalid request payload", nil))
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
//
//	@Description Get authenticated user's profile
//
// @Tags Auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} dto.ProfileResponse "Profile retrieved successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
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

	response := dtoResponse.ProfileResponse{
		UserResponse: mapToUserResponse(user),
	}

	h.response.Success(c, http.StatusOK, response)
}

func mapToUserResponse(user *model.User) dtoResponse.UserResponse {
	return dtoResponse.UserResponse{
		UID:           user.UID,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		Status:        user.Status,
		Role:          user.Role,
		AdminApproved: user.AdminApproved,
		ApprovalDate:  user.ApprovalDate,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}
