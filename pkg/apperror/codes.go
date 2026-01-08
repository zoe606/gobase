// Package apperror provides standardized error codes and types for the application.
package apperror

// Error codes for client errors (4xx).
const (
	CodeBadRequest      = "BAD_REQUEST"
	CodeValidationError = "VALIDATION_ERROR"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeForbidden       = "FORBIDDEN"
	CodeNotFound        = "NOT_FOUND"
	CodeConflict        = "CONFLICT"
	CodeRateLimited     = "RATE_LIMITED"
	CodeRequestTimeout  = "REQUEST_TIMEOUT"
)

// Error codes for server errors (5xx).
const (
	CodeInternalError      = "INTERNAL_ERROR"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	CodeTimeout            = "TIMEOUT"
)
