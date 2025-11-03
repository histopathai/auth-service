package request

import (
	"fmt"

	"github.com/histopathai/auth-service/internal/shared/errors"
)

const (
	DefaultLimit     = 20
	MaxLimit         = 100
	DefaultOffset    = 0
	DefaultSortBy    = "created_at"
	DefaultSortOrder = "desc"
)

type RegisterUserRequest struct {
	Email       string `json:"email" binding:"required,email" example:"user@example.com"`
	Password    string `json:"password" binding:"required,min=8" example:"strongpassword123"`
	DisplayName string `json:"display_name" binding:"required" example:"John Doe"`
}

type ChangePasswordForUserRequest struct {
	UID         string `json:"uid" binding:"required,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	NewPassword string `json:"new_password" binding:"required,min=8" example:"newstrongpassword123"`
}
type ChangePasswordSelfRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8" example:"newstrongpassword123"`
}

type VerifyTokenRequest struct {
	Token string `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
type RevokeSessionsRequest struct {
	SessionID string `json:"session_id" binding:"omitempty,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type QueryPaginationRequest struct {
	Limit     int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset    int    `form:"offset" binding:"omitempty,min=0"`
	SortBy    string `form:"sort_by" binding:"omitempty,oneof=created_at updated_at email"`
	SortOrder string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

func (qpr *QueryPaginationRequest) ApplyDefaults() {
	if qpr.Limit > MaxLimit {
		qpr.Limit = MaxLimit
	}
	if qpr.Offset < 0 {
		qpr.Offset = DefaultOffset
	}
	if qpr.SortBy == "" {
		qpr.SortBy = DefaultSortBy
	}
	if qpr.SortOrder == "" {
		qpr.SortOrder = DefaultSortOrder
	}
}

func (qpr *QueryPaginationRequest) ValidateSortFields(allowedFields []string) error {
	for _, field := range allowedFields {
		if field == qpr.SortBy {
			return nil
		}
	}
	return errors.NewValidationError("invali sort field", map[string]interface{}{
		"sort_by": fmt.Sprintf("must be one of: %v", allowedFields),
	})
}
