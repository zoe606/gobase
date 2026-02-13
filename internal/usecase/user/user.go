// Package user provides user management use cases.
package user

import (
	"go-boilerplate/internal/repo"
	"go-boilerplate/pkg/hasher"
)

// UseCase implements user management business logic.
type UseCase struct {
	userRepo repo.UserRepo
	roleRepo repo.RoleRepo
}

// New creates a new user use case.
func New(userRepo repo.UserRepo, roleRepo repo.RoleRepo) *UseCase {
	return &UseCase{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// hashPassword hashes a password using bcrypt.
func hashPassword(password string) (string, error) {
	return hasher.Hash(password)
}
