package firestore

import (
	"errors"

	sharedErrors "github.com/histopathai/auth-service/internal/shared/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MapFirestoreError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, iterator.Done) {
		return nil // No more documents
	}

	st, ok := status.FromError(err)
	if !ok {
		return sharedErrors.NewInternalError("Firestore Operation Failed", err)
	}

	switch st.Code() {
	case codes.NotFound:
		return sharedErrors.NewNotFoundError("Document not found")
	case codes.AlreadyExists:
		return sharedErrors.NewConflictError("Document already exists", nil)
	case codes.Aborted:
		return sharedErrors.NewConflictError("Transaction aborted, please retry", nil)
	case codes.PermissionDenied:
		return sharedErrors.NewForbiddenError("Permission denied for Firestore operation")
	default:
		return sharedErrors.NewInternalError("Firestore internal error", err)
	}
}
