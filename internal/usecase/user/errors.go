package user

import "errors"

var (
	// ErrUserNotFound is returned when a user is not found.
	ErrUserNotFound = errors.New("user not found")
	// ErrEmailExists is returned when the email already exists.
	ErrEmailExists = errors.New("email already exists")
	// ErrRoleNotFound is returned when the specified role is not found.
	ErrRoleNotFound = errors.New("role not found")
	// ErrCannotDeleteSelf is returned when a user tries to delete themselves.
	ErrCannotDeleteSelf = errors.New("cannot delete your own account")
)
