package role

import "errors"

var (
	// ErrRoleNotFound is returned when a role is not found.
	ErrRoleNotFound = errors.New("role not found")
	// ErrRoleNameExists is returned when the role name already exists.
	ErrRoleNameExists = errors.New("role name already exists")
	// ErrCannotDeleteRoleInUse is returned when trying to delete a role that is assigned to users.
	ErrCannotDeleteRoleInUse = errors.New("cannot delete role that is assigned to users")
)
