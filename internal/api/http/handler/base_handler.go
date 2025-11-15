package handler

import (
	stderr "errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	response "github.com/histopathai/auth-service/internal/api/http/dto/response"
	"github.com/histopathai/auth-service/internal/shared/errors"
)

type ResponseHelper struct{}

func (rh *ResponseHelper) Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, gin.H{"data": data})
}

func (rh *ResponseHelper) Error(c *gin.Context, statusCode int, errType string, message string, details map[string]interface{}) {
	c.JSON(statusCode, response.ErrorResponse{
		ErrorType: errType,
		Message:   message,
		Details:   details,
	})
}

func (rh *ResponseHelper) Created(c *gin.Context, data interface{}) {
	rh.Success(c, http.StatusCreated, data)
}

func (rh *ResponseHelper) NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func (rh *ResponseHelper) SuccessList(c *gin.Context, data interface{}, pagination *response.PaginationResponse) {
	response := gin.H{
		"data": data,
	}
	if pagination != nil {
		response["pagination"] = pagination
	}
	c.JSON(http.StatusOK, response)
}

type BaseHandler struct {
	logger   *slog.Logger
	response *ResponseHelper
}

func NewBaseHandler(logger *slog.Logger) *BaseHandler {
	return &BaseHandler{
		logger:   logger,
		response: &ResponseHelper{},
	}
}

func (bh *BaseHandler) handleError(c *gin.Context, err error) {
	requestID, _ := c.Get("request_id")
	var customErr *errors.Err

	if stderr.As(err, &customErr) {
		statusCode, errResponse := bh.mapCustomError(customErr)

		bh.logger.Error("Request failed",
			slog.String("request_id", requestID.(string)),
			slog.String("error_type", string(customErr.Type)),
			slog.String("message", customErr.Message),
			slog.String("path", c.Request.URL.Path),
		)
		c.JSON(statusCode, errResponse)
		return
	}

	bh.logger.Error("Request failed",
		slog.String("request_id", requestID.(string)),
		slog.String("error_type", "unknown"),
		slog.String("message", err.Error()),
		slog.String("path", c.Request.URL.Path),
	)
	c.JSON(http.StatusInternalServerError, response.ErrorResponse{
		ErrorType: "unknown",
		Message:   "An unexpected error occurred",
	})
}

func (bh *BaseHandler) mapCustomError(err *errors.Err) (int, response.ErrorResponse) {
	statusMap := map[errors.ErrorType]int{
		errors.ErrorTypeValidation:   http.StatusBadRequest,
		errors.ErrorTypeNotFound:     http.StatusNotFound,
		errors.ErrorTypeConflict:     http.StatusConflict,
		errors.ErrorTypeUnauthorized: http.StatusUnauthorized,
		errors.ErrorTypeForbidden:    http.StatusForbidden,
		errors.ErrorTypeInternal:     http.StatusInternalServerError,
	}

	statusCode, exists := statusMap[err.Type]
	if !exists {
		statusCode = http.StatusInternalServerError
	}

	return statusCode, response.ErrorResponse{
		ErrorType: string(err.Type),
		Message:   err.Message,
		Details:   err.Details,
	}
}
