package auth

import "errors"

// Common errors.
var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrEmailExists         = errors.New("email already exists")
	ErrUserNotActive       = errors.New("user account is not active")
	ErrInvalidToken        = errors.New("invalid or expired token")
	ErrDefaultRoleNotFound = errors.New("default role not found")

	// Email verification errors.
	ErrEmailAlreadyVerified = errors.New("email already verified")
	ErrVerificationNotFound = errors.New("verification token not found")
	ErrVerificationExpired  = errors.New("verification token expired")
	ErrVerificationUsed     = errors.New("verification token already used")

	// Password reset errors.
	ErrResetTokenNotFound = errors.New("password reset token not found")
	ErrResetTokenExpired  = errors.New("password reset token expired")
	ErrResetTokenUsed     = errors.New("password reset token already used")
)
