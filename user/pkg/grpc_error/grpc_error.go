package grpc_error

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrNotFound        = errors.New("Not Found")
	ErrNoCtxMetaData   = errors.New("No ctx metadata")
	ErrInvlidSessionID = errors.New("Invalid session id")
	ErrEmailExists     = errors.New("Email Already Exists")
)

func ParseGRPCErrStatusCodes(err error) codes.Code {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return codes.NotFound
	case errors.Is(err, context.Canceled):
		return codes.Canceled
	case errors.Is(err, context.DeadlineExceeded):
		return codes.DeadlineExceeded
	case errors.Is(err, ErrEmailExists):
		return codes.AlreadyExists
	case errors.Is(err, ErrNoCtxMetaData):
		return codes.Unauthenticated
	case errors.Is(err, ErrInvlidSessionID):
		return codes.PermissionDenied
	case strings.Contains(err.Error(), "Validate"):
		return codes.InvalidArgument
	case strings.Contains(err.Error(), "redis"):
		return codes.NotFound
	}
	return codes.Internal
}

func MapGRPCErrCodeToHttpStatus(code codes.Code) int {
	switch code {
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.AlreadyExists:
		return http.StatusBadRequest
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.InvalidArgument:
		return http.StatusBadGateway
	}
	return http.StatusInternalServerError
}

func ErrorResponse(err error, msg string) error {
	return status.Errorf(ParseGRPCErrStatusCodes(err), fmt.Sprintf("%s: %v", msg, err))
}
