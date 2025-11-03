package firebase

import (
	"strings"

	"firebase.google.com/go/auth"
	sharedErrors "github.com/histopathai/auth-service/internal/shared/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MapFirebaseAuthError(err error) error {
	if err == nil {
		return nil
	}

	// Check for email already exists
	if auth.IsEmailAlreadyExists(err) {
		return sharedErrors.NewConflictError("Email already in use", nil)
	}

	// Check for user not found
	if auth.IsUserNotFound(err) {
		return sharedErrors.NewNotFoundError("User not found")
	}

	// Check gRPC status codes for other errors
	st, ok := status.FromError(err)
	if ok {
		switch st.Code() {
		case codes.InvalidArgument:
			errMsg := strings.ToLower(st.Message())
			if strings.Contains(errMsg, "password") {
				return sharedErrors.NewValidationError("Invalid password", map[string]interface{}{"password": "Password does not meet the requirements"})
			}
			if strings.Contains(errMsg, "token") || strings.Contains(errMsg, "id token") {
				return sharedErrors.NewUnauthorizedError("Invalid or expired token")
			}
		case codes.Unauthenticated:
			return sharedErrors.NewUnauthorizedError("Invalid or expired token")
		case codes.PermissionDenied:
			return sharedErrors.NewForbiddenError("Permission denied for Firebase Auth operation")
		}
	}

	// Check error message for token-related issues
	errMsg := strings.ToLower(err.Error())
	if strings.Contains(errMsg, "token") && (strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "expired")) {
		return sharedErrors.NewUnauthorizedError("Invalid or expired token")
	}

	return sharedErrors.NewInternalError("Firebase Auth operation failed", err)
}
