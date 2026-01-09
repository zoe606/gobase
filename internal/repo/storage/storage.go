// Package storage provides file storage abstraction for multiple backends.
package storage

import (
	"context"
	"io"
	"time"
)

// FileInfo contains metadata about a stored file.
type FileInfo struct {
	Path     string
	Size     int64
	MimeType string
	Hash     string
}

// Provider defines the interface for file storage backends.
type Provider interface {
	// Put stores a file and returns its metadata.
	Put(ctx context.Context, path string, reader io.Reader, size int64, mimeType string) (*FileInfo, error)

	// Get retrieves a file as a reader.
	Get(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes a file.
	Delete(ctx context.Context, path string) error

	// Exists checks if a file exists.
	Exists(ctx context.Context, path string) (bool, error)

	// URL returns a public URL for the file.
	URL(ctx context.Context, path string) (string, error)

	// TemporaryURL returns a signed URL with expiration for private files.
	TemporaryURL(ctx context.Context, path string, expiry time.Duration) (string, error)

	// PresignedUploadURL returns a URL for direct client uploads.
	PresignedUploadURL(ctx context.Context, path string, expiry time.Duration) (string, error)
}
