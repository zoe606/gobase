package profile

import "errors"

// Errors.
var (
	ErrProfileNotFound = errors.New("profile not found")
	ErrInvalidMedia    = errors.New("invalid media: media does not exist or does not belong to user")
)
