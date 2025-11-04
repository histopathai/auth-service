package errors

import (
	"fmt"
)

type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "VALIDATION_ERROR"
	ErrorTypeNotFound     ErrorType = "NOT_FOUND_ERROR"
	ErrorTypeUnauthorized ErrorType = "UNAUTHORIZED_ERROR"
	ErrorTypeInternal     ErrorType = "INTERNAL_ERROR"
	ErrorTypeConflict     ErrorType = "CONFLICT_ERROR"
	ErrorTypeForbidden    ErrorType = "FORBIDDEN_ERROR"
)

type Err struct {
	Type    ErrorType
	Message string
	Details map[string]interface{}
	Err     error
}

func (e *Err) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *Err) Unwrap() error {
	return e.Err
}

// Helper fonksiyonlar
func NewValidationError(message string, details map[string]interface{}) *Err {
	return &Err{
		Type:    ErrorTypeValidation,
		Message: message,
		Details: details,
	}
}

func NewNotFoundError(message string) *Err {
	return &Err{
		Type:    ErrorTypeNotFound,
		Message: message,
	}
}

func NewConflictError(message string, details map[string]interface{}) *Err {
	return &Err{
		Type:    ErrorTypeConflict,
		Message: message,
		Details: details,
	}
}

func NewInternalError(message string, err error) *Err {
	return &Err{
		Type:    ErrorTypeInternal,
		Message: message,
		Err:     err,
	}
}

func NewUnauthorizedError(message string) *Err {
	return &Err{
		Type:    ErrorTypeUnauthorized,
		Message: message,
	}
}

func NewForbiddenError(message string) *Err {
	return &Err{
		Type:    ErrorTypeForbidden,
		Message: message,
	}
}
