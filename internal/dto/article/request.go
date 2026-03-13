package articledto

import (
	"time"

	"go-boilerplate/pkg/pagination"
)

// CreateRequest represents the request to create a Article.
type CreateRequest struct {
	Title        string    `json:"title" validate:"required"`
	Slug         string    `json:"slug" validate:"required"`
	Content      string    `json:"content" validate:"required"`
	Excerpt      string    `json:"excerpt" validate:"required"`
	CoverMediaID uint      `json:"cover_media_id" validate:"required"`
	Status       string    `json:"status" validate:"required"`
	PublishedAt  time.Time `json:"published_at" validate:"required"`
	ViewCount    int       `json:"view_count" validate:"required"`
}

// UpdateRequest represents the request to update a Article.
type UpdateRequest struct {
	Title        *string    `json:"title,omitempty"`
	Slug         *string    `json:"slug,omitempty"`
	Content      *string    `json:"content,omitempty"`
	Excerpt      *string    `json:"excerpt,omitempty"`
	CoverMediaID *uint      `json:"cover_media_id,omitempty"`
	Status       *string    `json:"status,omitempty"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
	ViewCount    *int       `json:"view_count,omitempty"`
}

// ListRequest represents the request to list articles with filters.
type ListRequest struct {
	pagination.Params
	Status string `query:"status" validate:"omitempty,oneof=draft published"`
	UserID uint   `query:"user_id"`
	Search string `query:"search"`
}
