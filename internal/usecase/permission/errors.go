package permission

import "errors"

var (
	// ErrPermissionExists is returned when a permission with the same name already exists.
	ErrPermissionExists = errors.New("permission already exists")
	// ErrPermissionNotFound is returned when a permission is not found.
	ErrPermissionNotFound = errors.New("permission not found")
	// ErrPermissionInUse is returned when trying to delete a permission that is assigned to roles.
	ErrPermissionInUse = errors.New("permission is assigned to one or more roles")
)
