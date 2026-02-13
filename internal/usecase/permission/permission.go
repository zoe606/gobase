// Package permission provides permission use cases.
package permission

import (
	"go-boilerplate/internal/repo"
)

// UseCase implements permission business logic.
type UseCase struct {
	permissionRepo repo.PermissionRepo
}

// New creates a new permission use case.
func New(permissionRepo repo.PermissionRepo) *UseCase {
	return &UseCase{
		permissionRepo: permissionRepo,
	}
}
