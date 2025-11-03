package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	response "github.com/histopathai/auth-service/internal/api/http/dto/reponse"
	"github.com/histopathai/auth-service/internal/api/http/dto/request"
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
// @Description Retrieves a list of all registered users (Admin access required)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param        limit query int false "Number of items per page" default(20) minimum(1) maximum(100)
// @Param        offset query int false "Number of items to skip" default(0) minimum(0)
// @Param        sort_by query string false "Field to sort by" default(created_at) Enums(created_at, updated_at, name)
// @Param        sort_dir query string false "Sort direction" default(desc) Enums(asc, desc)
// @Success 200 {object} response.SuccessListResponse{data=[]models.User,pagination=response.PaginationResponse} "Successfully retrieved users"
// @Faılure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	var paginationDTO request.QueryPaginationRequest
	if err := c.ShouldBindQuery(&paginationDTO); err != nil {
		h.handleError(c, err)
		return
	}

	paginationDTO.ApplyDefaults()

	allowedSortFields := []string{"created_at", "updated_at", "name"}
	if err := paginationDTO.ValidateSortFields(allowedSortFields); err != nil {
		h.handleError(c, err)
		return
	}

	//DTO -> Service
	pagination := &query.Pagination{
		Limit:     paginationDTO.Limit,
		Offset:    paginationDTO.Offset,
		SortBy:    paginationDTO.SortBy,
		SortOrder: paginationDTO.SortOrder,
	}

	result, err := h.authService.ListUsers(c.Request.Context(), pagination)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Service -> DTO
	paginationResp := &response.PaginationResponse{
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
	}

	h.response.SuccessList(c, result.Data, paginationResp)
}

// GetUser
// @Summary Get User by UID
// @Description Retrieves a single user by their UID (Admin access required)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uid path string true "User UID"
// @Success 200 {object} response.Success{data=model.User} "User retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/users/{uid} [get]
func (h *AdminHandler) GetUser(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		h.handleError(c, errors.NewValidationError("Missing UID", nil))
		return
	}

	user, err := h.authService.GetUserByUID(c.Request.Context(), uid)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.response.Success(c, http.StatusOK, user)
}

// ApproveUser
// @Summary Approve User
// @Description Approves a user account (Admin access required)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uid path string true "User UID"
// @Success 204 {object} response.NoContent
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/users/{uid}/approve [post]
func (h *AdminHandler) ApproveUser(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		h.handleError(c, errors.NewValidationError("Missing UID", nil))
		return
	}

	err := h.authService.ApproveUser(c.Request.Context(), uid)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.response.NoContent(c)
}

// SuspendUser
// @Summary Suspend User
// @Description Suspends a user account (Admin access required)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uid path string true "User UID"
// @Success 204 {object} response.NoContent "User suspended successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/users/{uid}/suspend [post]

func (h *AdminHandler) SuspendUser(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		h.handleError(c, errors.NewValidationError("Missing UID", nil))
		return
	}

	err := h.authService.SuspendUser(c.Request.Context(), uid)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.response.NoContent(c)
}

// ChangePasswordForUser
// @Summary Change User Password (Admin)
// @Description Endpoint for an administrator to change a specific user's password.
// @Tags Admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param new_password body request.ChangePasswordRequest true "New password"
// @Success 204 {object} response.NoContent "User password changed successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/users/{uid}/change-password [post]
func (h *AdminHandler) ChangePasswordForUser(c *gin.Context) {

	var req request.ChangePasswordForUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.NewValidationError("Invalid request payload", nil))
		return
	}

	err := h.authService.ChangeUserPassword(c.Request.Context(), req.UID, req.NewPassword)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.response.NoContent(c)
}

// MakeAdmin
// @Summary Grant Admin Role to User
// @Description Grants admin role to a user (Admin access required)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uid path string true "User UID"
// @Success 204 {object} response.NoContent "User granted admin role successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/users/{uid}/make-admin [post]
func (h *AdminHandler) MakeAdmin(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		h.handleError(c, errors.NewValidationError("Missing UID", nil))
		return
	}

	err := h.authService.PromoteUserToAdmin(c.Request.Context(), uid)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.response.NoContent(c)
}
