package article

import "time"

// CreateRequest represents the request to create a Article.
type CreateRequest struct {
	UserID       uint      `json:"user_id" validate:"required"`
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
	UserID       *uint      `json:"user_id,omitempty"`
	Title        *string    `json:"title,omitempty"`
	Slug         *string    `json:"slug,omitempty"`
	Content      *string    `json:"content,omitempty"`
	Excerpt      *string    `json:"excerpt,omitempty"`
	CoverMediaID *uint      `json:"cover_media_id,omitempty"`
	Status       *string    `json:"status,omitempty"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
	ViewCount    *int       `json:"view_count,omitempty"`
}

// ListRequest represents the request to list articles.
type ListRequest struct {
	Page     int `query:"page" validate:"omitempty,min=1"`
	PageSize int `query:"page_size" validate:"omitempty,min=1,max=100"`
}

// GetPageSize returns the page size with a default value.
func (r *ListRequest) GetPageSize() int {
	if r.PageSize <= 0 {
		return 20
	}
	return r.PageSize
}

// GetOffset returns the offset for pagination.
func (r *ListRequest) GetOffset() int {
	if r.Page <= 1 {
		return 0
	}
	return (r.Page - 1) * r.GetPageSize()
}
