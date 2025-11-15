package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	dtoRequest "github.com/histopathai/auth-service/internal/api/http/dto/request"
	dtoResponse "github.com/histopathai/auth-service/internal/api/http/dto/response"
	"github.com/histopathai/auth-service/internal/service"
	"github.com/histopathai/auth-service/internal/shared/errors"
	"github.com/histopathai/auth-service/internal/shared/query"
)

type AdminHandler struct {
	authService service.AuthService
	BaseHandler
}

func NewAdminHandler(authService service.AuthService, logger *slog.Logger) *AdminHandler {
	return &AdminHandler{
		authService: authService,
		BaseHandler: BaseHandler{logger: logger, response: &ResponseHelper{}},
	}
}

// ListUsers
// @Summary List Users
// @Description Get paginated list of all users (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Param offset query int false "Items to skip" default(0) minimum(0)
// @Param sort_by query string false "Sort field" default(created_at) Enums(created_at, updated_at, email, display_name)
// @Param sort_order query string false "Sort direction" default(desc) Enums(asc, desc)
// @Param status query string false "Filter by status" Enums(pending, active, suspended)
// @Param role query string false "Filter by role" Enums(user, admin)
// @Param search query string false "Search in email and display name"
// @Success 200 {object} response.UserListResponse "Users retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	var req dtoRequest.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.handleError(c, errors.NewValidationError("Invalid query parameters", nil))
		return
	}

	req.ApplyDefaults()

	allowedFields := req.GetAllowedSortFields()
	isValid := false
	for _, field := range allowedFields {
		if field == req.SortBy {
			isValid = true
			break
		}
	}
	if !isValid {
		h.handleError(c, errors.NewValidationError("Invalid sort field", map[string]interface{}{
			"sort_by": "must be one of: created_at, updated_at, email, display_name",
		}))
		return
	}

	pagination := &query.Pagination{
		Limit:     req.Limit,
		Offset:    req.Offset,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	}

	result, err := h.authService.ListUsers(c.Request.Context(), pagination)
	if err != nil {
		h.handleError(c, err)
		return
	}

	users := make([]dtoResponse.UserResponse, len(result.Data))
	for i, user := range result.Data {
		users[i] = mapToUserResponse(user)
	}

	response := dtoResponse.UserListResponse{
		Data: users,
		Pagination: dtoResponse.PaginationResponse{
			Limit:   result.Limit,
			Offset:  result.Offset,
			HasMore: result.HasMore,
		},
	}

	h.response.SuccessList(c, response.Data, &response.Pagination)
}

// GetUser
// @Summary Get User by ID
// @Description Get detailed user information by ID (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path string true "User UserID"
// @Success 200 {object} response.UserDetailResponse "User retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid UserID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/users/{user_id} [get]
func (h *AdminHandler) GetUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		h.handleError(c, errors.NewValidationError("Missing UserID", nil))
		return
	}

	user, err := h.authService.GetUserByUserID(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := dtoResponse.UserDetailResponse{
		UserResponse: mapToUserResponse(user),
	}

	h.response.Success(c, http.StatusOK, response)
}

// ApproveUser
// @Summary Approve User
// @Description Approve pending user account (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path string true "User UserID"
// @Success 200 {object} response.UserActionResponse "User approved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid UserID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/users/{user_id}/approve [post]
func (h *AdminHandler) ApproveUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		h.handleError(c, errors.NewValidationError("Missing UserID", nil))
		return
	}

	err := h.authService.ApproveUser(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	user, _ := h.authService.GetUserByUserID(c.Request.Context(), userID)

	response := dtoResponse.UserActionResponse{
		Message: "User approved successfully",
		User:    mapToUserResponse(user),
	}

	h.response.Success(c, http.StatusOK, response)
}

// ChangePasswordForUser
// @Summary Change User Password (Admin)
// @Description Admin changes specific user's password
// @Tags Admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path string true "User UserID"
// @Param payload body request.ChangeUserPasswordRequest true "New password"
// @Success 204 "Password changed successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/users/{user_id}/change-password [post]
func (h *AdminHandler) ChangePasswordForUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		h.handleError(c, errors.NewValidationError("Missing UserID", nil))
		return
	}

	var req dtoRequest.ChangeUserPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.NewValidationError("Invalid request payload", nil))
		return
	}

	err := h.authService.ChangeUserPassword(c.Request.Context(), userID, req.NewPassword)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.response.NoContent(c)
}

// SuspendUser
// @Summary Suspend User
// @Description Suspend an active user account (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path string true "User UserID"
// @Success 200 {object} response.UserActionResponse "User suspended successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid UserID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/users/{user_id}/suspend [post]
func (h *AdminHandler) SuspendUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		h.handleError(c, errors.NewValidationError("Missing UserID", nil))
		return
	}

	err := h.authService.SuspendUser(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Updated user'ı getir
	user, _ := h.authService.GetUserByUserID(c.Request.Context(), userID)

	response := dtoResponse.UserActionResponse{
		Message: "User suspended successfully",
		User:    mapToUserResponse(user),
	}

	h.response.Success(c, http.StatusOK, response)
}

// MakeAdmin
// @Summary Make User Admin
// @Description Grant admin role to a user (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path string true "User UserID"
// @Success 200 {object} response.UserActionResponse "User granted admin role successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid UserID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/users/{user_id}/make-admin [post]
func (h *AdminHandler) MakeAdmin(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		h.handleError(c, errors.NewValidationError("Missing UserID", nil))
		return
	}

	err := h.authService.PromoteUserToAdmin(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Updated user'ı getir
	user, _ := h.authService.GetUserByUserID(c.Request.Context(), userID)

	response := dtoResponse.UserActionResponse{
		Message: "User granted admin role successfully",
		User:    mapToUserResponse(user),
	}

	h.response.Success(c, http.StatusOK, response)
}
