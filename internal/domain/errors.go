package domain

import (
	"errors"
	"fmt"
)

type ErrorCode string

const (
	ErrInvalidArgument ErrorCode = "INVALID_ARGUMENT"
	ErrInvalidState    ErrorCode = "INVALID_STATE"
	ErrConflict        ErrorCode = "CONFLICT"
	ErrForbidden       ErrorCode = "FORBIDDEN"
	ErrNotFound        ErrorCode = "NOT_FOUND"
)

type DomainError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return string(e.Code)
}

func (e *DomainError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewDomainError(code ErrorCode, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

func WrapDomainError(code ErrorCode, message string, err error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func AsDomainError(err error) *DomainError {
	if err == nil {
		return nil
	}

	var de *DomainError
	if errors.As(err, &de) {
		return de
	}
	return nil
}
