package auth

import "errors"

// Common errors.
var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrEmailExists         = errors.New("email already exists")
	ErrUserNotActive       = errors.New("user account is not active")
	ErrInvalidToken        = errors.New("invalid or expired token")
	ErrDefaultRoleNotFound = errors.New("default role not found")
)
