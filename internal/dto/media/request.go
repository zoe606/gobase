package media

import "io"

// UploadRequest contains parameters for uploading a file.
type UploadRequest struct {
	File           io.Reader
	Filename       string
	Size           int64
	MimeType       string
	AttachableType string // e.g., "users", "posts"
	AttachableID   uint
	Collection     string // e.g., "avatar", "gallery", "documents"
}

// PresignedURLRequest contains parameters for getting a presigned upload URL.
type PresignedURLRequest struct {
	Filename string `json:"filename" validate:"required"`
}

// ConfirmUploadRequest confirms a direct upload.
type ConfirmUploadRequest struct {
	Path           string `json:"path" validate:"required"`
	AttachableType string `json:"attachable_type" validate:"required"`
	AttachableID   uint   `json:"attachable_id" validate:"required"`
	Collection     string `json:"collection"`
}

// GetMediaRequest contains parameters for getting media.
type GetMediaRequest struct {
	AttachableType string `json:"attachable_type" validate:"required"`
	AttachableID   uint   `json:"attachable_id" validate:"required"`
	Collection     string `json:"collection"`
}
