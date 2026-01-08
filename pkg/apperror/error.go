package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents an application error with code, message, and HTTP status.
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
	Err     error  `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithError wraps an underlying error.
func (e *AppError) WithError(err error) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Status:  e.Status,
		Err:     err,
	}
}

// WithMessage returns a new AppError with a custom message.
func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: msg,
		Status:  e.Status,
		Err:     e.Err,
	}
}

// Predefined errors for common scenarios.
var (
	ErrBadRequest = &AppError{
		Code:    CodeBadRequest,
		Message: "Bad request",
		Status:  http.StatusBadRequest,
	}

	ErrValidation = &AppError{
		Code:    CodeValidationError,
		Message: "Validation failed",
		Status:  http.StatusBadRequest,
	}

	ErrUnauthorized = &AppError{
		Code:    CodeUnauthorized,
		Message: "Unauthorized",
		Status:  http.StatusUnauthorized,
	}

	ErrForbidden = &AppError{
		Code:    CodeForbidden,
		Message: "Forbidden",
		Status:  http.StatusForbidden,
	}

	ErrNotFound = &AppError{
		Code:    CodeNotFound,
		Message: "Resource not found",
		Status:  http.StatusNotFound,
	}

	ErrConflict = &AppError{
		Code:    CodeConflict,
		Message: "Resource conflict",
		Status:  http.StatusConflict,
	}

	ErrRateLimited = &AppError{
		Code:    CodeRateLimited,
		Message: "Too many requests",
		Status:  http.StatusTooManyRequests,
	}

	ErrRequestTimeout = &AppError{
		Code:    CodeRequestTimeout,
		Message: "Request timeout",
		Status:  http.StatusGatewayTimeout,
	}

	ErrInternal = &AppError{
		Code:    CodeInternalError,
		Message: "Internal server error",
		Status:  http.StatusInternalServerError,
	}

	ErrServiceUnavailable = &AppError{
		Code:    CodeServiceUnavailable,
		Message: "Service unavailable",
		Status:  http.StatusServiceUnavailable,
	}
)

// New creates a new AppError with the given code, message, and status.
func New(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// IsAppError checks if an error is an AppError and returns it.
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
