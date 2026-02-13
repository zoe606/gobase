// Package role provides role management use cases.
package role

import (
	"go-boilerplate/internal/repo"
)

// UseCase implements role management business logic.
type UseCase struct {
	roleRepo       repo.RoleRepo
	permissionRepo repo.PermissionRepo
}

// New creates a new role use case.
func New(roleRepo repo.RoleRepo, permissionRepo repo.PermissionRepo) *UseCase {
	return &UseCase{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}
