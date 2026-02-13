package user

import "go-boilerplate/pkg/pagination"

// CreateRequest represents the request to create a User.
type CreateRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
	RoleID   uint   `json:"role_id" validate:"required"`
	Active   *bool  `json:"active,omitempty"`
}

// UpdateRequest represents the request to update a User.
type UpdateRequest struct {
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=8"`
	Name     *string `json:"name,omitempty"`
	RoleID   *uint   `json:"role_id,omitempty"`
	Active   *bool   `json:"active,omitempty"`
}

// ListRequest represents the request to list users with filters.
type ListRequest struct {
	pagination.Params
	Search string `query:"search"`
	RoleID uint   `query:"role_id"`
	Active *bool  `query:"active"`
}
