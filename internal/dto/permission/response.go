package permission

import (
	"time"

	"go-boilerplate/internal/entity"
)

// Response represents a permission response.
type Response struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Resource  string    `json:"resource"`
	Action    string    `json:"action"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewResponse creates a Response from an entity.Permission.
func NewResponse(p *entity.Permission) *Response {
	if p == nil {
		return nil
	}
	return &Response{
		ID:        p.ID,
		Name:      p.Name,
		Resource:  p.Resource,
		Action:    p.Action,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

// NewListResponse creates a slice of Responses from a slice of Permissions.
func NewListResponse(permissions []*entity.Permission) []*Response {
	data := make([]*Response, len(permissions))
	for i, p := range permissions {
		data[i] = NewResponse(p)
	}
	return data
}
