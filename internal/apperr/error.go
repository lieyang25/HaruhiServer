package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

type Code string

const (
	// CodeOK marks successful responses.
	CodeOK               Code = "OK"
	// CodeInvalidArgument marks bad client input.
	CodeInvalidArgument  Code = "INVALID_ARGUMENT"
	// CodeNotFound marks missing resources.
	CodeNotFound         Code = "NOT_FOUND"
	// CodeMethodNotAllowed marks unsupported HTTP methods.
	CodeMethodNotAllowed Code = "METHOD_NOT_ALLOWED"
	// CodeConflict marks state conflicts (for example duplicates).
	CodeConflict         Code = "CONFLICT"
	// CodeInternal marks unexpected server-side failures.
	CodeInternal         Code = "INTERNAL"
)

// Error is the project-level typed error used for API responses.
type Error struct {
	Code    Code
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}

	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}

	return string(e.Code)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func New(code Code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func Wrap(code Code, message string, err error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// As extracts *Error from wrapped errors.
func As(err error) *Error {
	if err == nil {
		return nil
	}

	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}

	return nil
}

// HTTPStatus maps business error codes to transport status codes.
func HTTPStatus(err error) int {
	appErr := As(err)

	if appErr == nil {
		return http.StatusInternalServerError
	}

	switch appErr.Code {
	case CodeInvalidArgument:
		return http.StatusBadRequest
	case CodeNotFound:
		return http.StatusNotFound
	case CodeMethodNotAllowed:
		return http.StatusMethodNotAllowed
	case CodeConflict:
		return http.StatusConflict
	case CodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
