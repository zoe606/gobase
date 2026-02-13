package role

// CreateRequest represents the request to create a Role.
type CreateRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=50"`
	Description string `json:"description" validate:"max=255"`
}

// UpdateRequest represents the request to update a Role.
type UpdateRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=2,max=50"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=255"`
}

// AssignPermissionsRequest represents the request to assign permissions to a role.
type AssignPermissionsRequest struct {
	PermissionIDs []uint `json:"permission_ids" validate:"required"`
}
