package media

import (
	"time"

	"go-boilerplate/internal/entity"
)

// MediaResponse represents a media item in API responses.
type MediaResponse struct {
	ID             uint                   `json:"id"`
	AttachableType string                 `json:"attachable_type"`
	AttachableID   uint                   `json:"attachable_id"`
	Collection     string                 `json:"collection"`
	Filename       string                 `json:"filename"`
	OriginalName   string                 `json:"original_name"`
	MimeType       string                 `json:"mime_type"`
	Size           int64                  `json:"size"`
	Type           entity.MediaType       `json:"type"`
	Width          *int                   `json:"width,omitempty"`
	Height         *int                   `json:"height,omitempty"`
	Variants       map[string]interface{} `json:"variants,omitempty"`
	URL            string                 `json:"url,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// PresignedURLResponse contains the presigned upload URL.
type PresignedURLResponse struct {
	UploadURL string `json:"upload_url"`
	Path      string `json:"path"`
	ExpiresIn int    `json:"expires_in"` // seconds
}

// MediaListResponse contains a list of media items.
type MediaListResponse struct {
	Items []*MediaResponse `json:"items"`
	Total int              `json:"total"`
}

// FromEntity converts an entity.Media to MediaResponse.
func FromEntity(m *entity.Media) *MediaResponse {
	return &MediaResponse{
		ID:             m.ID,
		AttachableType: m.AttachableType,
		AttachableID:   m.AttachableID,
		Collection:     m.Collection,
		Filename:       m.Filename,
		OriginalName:   m.OriginalName,
		MimeType:       m.MimeType,
		Size:           m.Size,
		Type:           m.Type,
		Width:          m.Width,
		Height:         m.Height,
		Variants:       m.Variants,
		CreatedAt:      m.CreatedAt,
	}
}

// FromEntities converts a slice of entity.Media to MediaListResponse.
func FromEntities(media []*entity.Media) *MediaListResponse {
	items := make([]*MediaResponse, len(media))
	for i, m := range media {
		items[i] = FromEntity(m)
	}
	return &MediaListResponse{
		Items: items,
		Total: len(items),
	}
}
