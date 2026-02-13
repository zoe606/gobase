package role

import (
	"time"

	"go-boilerplate/internal/entity"
)

// PermissionResponse represents a permission in a role response.
type PermissionResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// Response represents a Role response.
type Response struct {
	ID          uint                  `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Permissions []*PermissionResponse `json:"permissions"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// NewResponse creates a Response from an entity.Role.
func NewResponse(role *entity.Role) *Response {
	if role == nil {
		return nil
	}

	permissions := make([]*PermissionResponse, len(role.Permissions))
	for i, p := range role.Permissions {
		permissions[i] = &PermissionResponse{
			ID:       p.ID,
			Name:     p.Name,
			Resource: p.Resource,
			Action:   p.Action,
		}
	}

	return &Response{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Permissions: permissions,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

// ListResponse represents a list of Role responses.
type ListResponse struct {
	Data []*Response `json:"data"`
}

// NewListResponse creates a ListResponse from a slice of Roles.
func NewListResponse(roles []*entity.Role) *ListResponse {
	data := make([]*Response, len(roles))
	for i, role := range roles {
		data[i] = NewResponse(role)
	}

	return &ListResponse{
		Data: data,
	}
}
