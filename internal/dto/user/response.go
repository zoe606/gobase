package user

import (
	"time"

	"go-boilerplate/internal/entity"
	"go-boilerplate/pkg/pagination"
)

// Response represents a User response.
type Response struct {
	ID              uint       `json:"id"`
	Email           string     `json:"email"`
	Name            string     `json:"name"`
	RoleID          uint       `json:"role_id"`
	RoleName        string     `json:"role_name"`
	Active          bool       `json:"active"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// NewResponse creates a Response from an entity.User.
func NewResponse(user *entity.User) *Response {
	if user == nil {
		return nil
	}
	return &Response{
		ID:              user.ID,
		Email:           user.Email,
		Name:            user.Name,
		RoleID:          user.RoleID,
		RoleName:        user.Role.Name,
		Active:          user.Active,
		EmailVerifiedAt: user.EmailVerifiedAt,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}

// ListResponse represents a paginated list of User responses.
type ListResponse struct {
	Data []*Response      `json:"data"`
	Meta *pagination.Meta `json:"meta"`
}

// NewListResponse creates a ListResponse from a slice of Users.
func NewListResponse(users []*entity.User, total int64, params pagination.Params) *ListResponse {
	data := make([]*Response, len(users))
	for i, user := range users {
		data[i] = NewResponse(user)
	}

	return &ListResponse{
		Data: data,
		Meta: pagination.NewMeta(params.Page, params.Limit, total),
	}
}
