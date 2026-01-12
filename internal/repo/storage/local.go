package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// ErrNotSupported is returned when an operation is not supported by the storage backend.
var ErrNotSupported = errors.New("operation not supported")

// LocalStorage implements Provider for local filesystem storage.
type LocalStorage struct {
	basePath string
	baseURL  string
}

// NewLocalStorage creates a new local storage provider.
func NewLocalStorage(basePath, baseURL string) *LocalStorage {
	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}
}

// Put stores a file on the local filesystem.
func (s *LocalStorage) Put(ctx context.Context, path string, reader io.Reader, size int64, mimeType string) (*FileInfo, error) {
	fullPath := filepath.Join(s.basePath, path)

	// Create directory if not exists.
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	// Create file.
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	// Hash while writing.
	hasher := sha256.New()
	writer := io.MultiWriter(file, hasher)

	written, err := io.Copy(writer, reader)
	if err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	return &FileInfo{
		Path:     path,
		Size:     written,
		MimeType: mimeType,
		Hash:     hex.EncodeToString(hasher.Sum(nil)),
	}, nil
}

// Get retrieves a file from the local filesystem.
func (s *LocalStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, path)
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("open file: %w", err)
	}
	return file, nil
}

// Delete removes a file from the local filesystem.
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.basePath, path)
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted.
		}
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

// Exists checks if a file exists on the local filesystem.
func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := filepath.Join(s.basePath, path)
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("stat file: %w", err)
	}
	return true, nil
}

// URL returns a public URL for the file.
func (s *LocalStorage) URL(ctx context.Context, path string) (string, error) {
	u, err := url.JoinPath(s.baseURL, path)
	if err != nil {
		return "", fmt.Errorf("join path: %w", err)
	}
	return u, nil
}

// TemporaryURL returns the same URL as URL for local storage.
// Local storage doesn't support signed URLs natively.
func (s *LocalStorage) TemporaryURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	return s.URL(ctx, path)
}

// PresignedUploadURL is not supported for local storage.
func (s *LocalStorage) PresignedUploadURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	return "", ErrNotSupported
}
