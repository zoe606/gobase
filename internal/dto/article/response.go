package article

import (
	"time"

	"go-boilerplate/internal/entity"
)

// Response represents a Article response.
type Response struct {
	ID           uint       `json:"id"`
	UserID       uint       `json:"user_id,omitempty"`
	Title        string     `json:"title,omitempty"`
	Slug         string     `json:"slug,omitempty"`
	Content      *string    `json:"content,omitempty"`
	Excerpt      *string    `json:"excerpt,omitempty"`
	CoverMediaID *uint      `json:"cover_media_id,omitempty"`
	Status       *string    `json:"status,omitempty"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
	ViewCount    *int       `json:"view_count,omitempty"`
	CreatedAt    time.Time  `json:"created_at,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at,omitempty"`
}

// NewResponse creates a Response from an entity.Article.
func NewResponse(article *entity.Article) *Response {
	if article == nil {
		return nil
	}
	return &Response{
		ID:           article.ID,
		UserID:       article.UserID,
		Title:        article.Title,
		Slug:         article.Slug,
		Content:      article.Content,
		Excerpt:      article.Excerpt,
		CoverMediaID: article.CoverMediaID,
		Status:       article.Status,
		PublishedAt:  article.PublishedAt,
		ViewCount:    article.ViewCount,
		CreatedAt:    article.CreatedAt,
		UpdatedAt:    article.UpdatedAt,
	}
}

// ListResponse represents a list of Article responses.
type ListResponse struct {
	Data       []*Response `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// NewListResponse creates a ListResponse from a slice of Articles.
func NewListResponse(articles []*entity.Article, total int64, page, pageSize int) *ListResponse {
	data := make([]*Response, len(articles))
	for i, article := range articles {
		data[i] = NewResponse(article)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &ListResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
