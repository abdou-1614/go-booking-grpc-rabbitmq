package http_error

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/labstack/echo/v4"
)

const (
	ErrBadRequest       = "Bad request"
	ErrAlreadyExists    = "Already exists"
	ErrNoSuchUser       = "User not found"
	ErrWrongCredentials = "Wrong Credentials"
	ErrNotFound         = "Not Found"
	ErrUnauthorized     = "Unauthorized"
	ErrForbidden        = "Forbidden"
	ErrBadQueryParams   = "Invalid query params"
	ErrRequestTimeout   = "Request Timeout"
	ErrInvalidEmail     = "Invalid email"
	ErrInvalidPassword  = "Invalid password"
	ErrInvalidField     = "Invalid field"
)

var (
	BadRequest            = errors.New("Bad request")
	WrongCredentials      = errors.New("Wrong Credentials")
	NotFound              = errors.New("Not Found")
	Unauthorized          = errors.New("Unauthorized")
	Forbidden             = errors.New("Forbidden")
	PermissionDenied      = errors.New("Permission Denied")
	NotRequiredFields     = errors.New("No such required fields")
	BadQueryParams        = errors.New("Invalid query params")
	InternalServerError   = errors.New("Internal Server Error")
	RequestTimeoutError   = errors.New("Request Timeout")
	ExistsEmailError      = errors.New("User with given email already exists")
	InvalidJWTToken       = errors.New("Invalid JWT token")
	InvalidJWTClaims      = errors.New("Invalid JWT claims")
	NotAllowedImageHeader = errors.New("Not allowed image header")
	NoCookie              = errors.New("not found cookie header")
	InvalidUUID           = errors.New("invalid uuid")
)

type RestError struct {
	ErrStatus int         `json:"status,omitempty"`
	ErrError  string      `json:"error,omitempty"`
	ErrCause  interface{} `json:"cause,omitempty"`
}

type RestErr interface {
	Status() int
	Error() string
	Cause() interface{}
	ErrBody() RestError
}

func (e RestError) ErrBody() RestError {
	return e
}

func (e RestError) Error() string {
	return fmt.Sprintf("status: %v, error: %v, cause: %v", e.ErrStatus, e.ErrError, e.ErrCause)
}

func (e RestError) Status() int {
	return e.ErrStatus
}

func (e RestError) Cause() interface{} {
	return e.ErrCause
}

func NewRestErr(status int, err string, causes interface{}) RestErr {
	return RestError{
		ErrStatus: status,
		ErrError:  err,
		ErrCause:  causes,
	}
}

func NewRestErrWithMessages(status int, err string, cause interface{}) RestErr {
	return RestError{
		ErrStatus: status,
		ErrError:  err,
		ErrCause:  cause,
	}
}

func NewRestErrFromBytes(bytes []byte) (RestErr, error) {
	var apiError RestError

	if err := json.Unmarshal(bytes, &apiError); err != nil {
		return nil, errors.New("Invalid Json")
	}

	return apiError, nil
}

func NewBadRequestErr(cause interface{}) RestErr {
	return RestError{
		ErrStatus: http.StatusBadGateway,
		ErrError:  BadRequest.Error(),
		ErrCause:  cause,
	}
}

func NewNotFoundErr(cause interface{}) RestErr {
	return RestError{
		ErrStatus: http.StatusNotFound,
		ErrError:  NotFound.Error(),
		ErrCause:  cause,
	}
}

func NewUnauthorizedError(cause interface{}) RestErr {
	return RestError{
		ErrStatus: http.StatusUnauthorized,
		ErrError:  Unauthorized.Error(),
		ErrCause:  cause,
	}
}

func NewForbiddenError(cause interface{}) RestErr {
	return RestError{
		ErrStatus: http.StatusForbidden,
		ErrError:  Forbidden.Error(),
		ErrCause:  cause,
	}
}

func NewInternalServerError(cause interface{}) RestErr {
	result := RestError{
		ErrStatus: http.StatusInternalServerError,
		ErrError:  InternalServerError.Error(),
		ErrCause:  cause,
	}
	return result
}

func ParseErrors(err error) RestErr {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return NewRestErr(http.StatusNotFound, ErrNotFound, nil)
	case errors.Is(err, context.DeadlineExceeded):
		return NewRestErr(http.StatusRequestTimeout, ErrRequestTimeout, nil)
	case errors.Is(err, Unauthorized):
		return NewRestErr(http.StatusUnauthorized, ErrUnauthorized, nil)
	case errors.Is(err, WrongCredentials):
		return NewRestErr(http.StatusUnauthorized, ErrUnauthorized, nil)
	case strings.Contains(strings.ToLower(err.Error()), "sqlstate"):
		return ParseSqlError(err)
	case strings.Contains(strings.ToLower(err.Error()), "field validation"):
		return ParseValidationError(err)
	case strings.Contains(strings.ToLower(err.Error()), "unmarshal"):
		return NewRestErr(http.StatusBadRequest, ErrBadRequest, err)
	case strings.Contains(strings.ToLower(err.Error()), "uuid"):
		return NewRestErr(http.StatusBadRequest, ErrBadRequest, err)
	case strings.Contains(strings.ToLower(err.Error()), "cookie"):
		return NewRestErr(http.StatusUnauthorized, ErrUnauthorized, err)
	case strings.Contains(strings.ToLower(err.Error()), "token"):
		return NewRestErr(http.StatusUnauthorized, ErrUnauthorized, err)
	case strings.Contains(strings.ToLower(err.Error()), "bcrypt"):
		return NewRestErr(http.StatusBadRequest, ErrBadRequest, nil)
	default:
		if restErr, ok := err.(RestErr); ok {
			return restErr
		}
		return NewInternalServerError(err)
	}
}

func ParseSqlError(err error) RestErr {
	if strings.Contains(err.Error(), pgerrcode.UniqueViolation) {
		return NewRestErr(http.StatusBadRequest, ErrAlreadyExists, err)
	}

	return NewRestErr(http.StatusBadRequest, ErrBadRequest, err)
}

func ParseValidationError(err error) RestErr {
	if strings.Contains(err.Error(), "Password") {
		return NewRestErr(http.StatusBadRequest, ErrInvalidPassword, err)
	}

	if strings.Contains(err.Error(), "Email") {
		return NewRestErr(http.StatusBadRequest, ErrInvalidEmail, err)
	}

	return NewRestErr(http.StatusBadRequest, ErrInvalidField, err)
}

func ErrorResponse(err error) (int, interface{}) {
	return ParseErrors(err).Status(), ParseErrors(err)
}

func ErrorCtxResponse(ctx echo.Context, err error) error {
	restErr := ParseErrors(err)
	return ctx.JSON(restErr.Status(), restErr.ErrBody())
}
