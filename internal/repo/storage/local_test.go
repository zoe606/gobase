package storage_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/repo/storage"
)

func setupTestStorage(t *testing.T) (*storage.LocalStorage, string) {
	t.Helper()

	// Create temp directory for tests
	tempDir, err := os.MkdirTemp("", "storage-test-*")
	require.NoError(t, err)

	s := storage.NewLocalStorage(tempDir, "http://localhost:8080/uploads")

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return s, tempDir
}

func TestLocalStorage_Put(t *testing.T) {
	t.Parallel()

	s, tempDir := setupTestStorage(t)
	ctx := context.Background()

	tests := []struct {
		name     string
		path     string
		content  string
		size     int64
		mimeType string
		wantErr  bool
	}{
		{
			name:     "success - simple file",
			path:     "test.txt",
			content:  "hello world",
			size:     11,
			mimeType: "text/plain",
			wantErr:  false,
		},
		{
			name:     "success - nested path",
			path:     "users/avatar/1/2024/01/image.jpg",
			content:  "fake image data",
			size:     15,
			mimeType: "image/jpeg",
			wantErr:  false,
		},
		{
			name:     "success - empty file",
			path:     "empty.txt",
			content:  "",
			size:     0,
			mimeType: "text/plain",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader([]byte(tt.content))
			info, err := s.Put(ctx, tt.path, reader, tt.size, tt.mimeType)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, info)
			assert.Equal(t, tt.path, info.Path)
			assert.Equal(t, int64(len(tt.content)), info.Size)
			assert.Equal(t, tt.mimeType, info.MimeType)
			assert.NotEmpty(t, info.Hash)

			// Verify file exists on disk
			fullPath := filepath.Join(tempDir, tt.path)
			_, err = os.Stat(fullPath)
			assert.NoError(t, err)

			// Verify content
			content, err := os.ReadFile(fullPath)
			require.NoError(t, err)
			assert.Equal(t, tt.content, string(content))
		})
	}
}

func TestLocalStorage_Get(t *testing.T) {
	t.Parallel()

	s, tempDir := setupTestStorage(t)
	ctx := context.Background()

	// Create a test file
	testPath := "test-get.txt"
	testContent := "test content for get"
	fullPath := filepath.Join(tempDir, testPath)
	err := os.WriteFile(fullPath, []byte(testContent), 0o644)
	require.NoError(t, err)

	tests := []struct {
		name        string
		path        string
		wantContent string
		wantErr     bool
	}{
		{
			name:        "success",
			path:        testPath,
			wantContent: testContent,
			wantErr:     false,
		},
		{
			name:    "file not found",
			path:    "nonexistent.txt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := s.Get(ctx, tt.path)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, reader)
			defer reader.Close()

			content, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.Equal(t, tt.wantContent, string(content))
		})
	}
}

func TestLocalStorage_Delete(t *testing.T) {
	t.Parallel()

	s, tempDir := setupTestStorage(t)
	ctx := context.Background()

	// Create a test file
	testPath := "test-delete.txt"
	fullPath := filepath.Join(tempDir, testPath)
	err := os.WriteFile(fullPath, []byte("to be deleted"), 0o644)
	require.NoError(t, err)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "success",
			path:    testPath,
			wantErr: false,
		},
		{
			name:    "file not found - no error",
			path:    "nonexistent.txt",
			wantErr: false, // Most implementations don't error on delete of non-existent files
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.Delete(ctx, tt.path)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify file no longer exists
			if tt.path == testPath {
				_, err = os.Stat(fullPath)
				assert.True(t, os.IsNotExist(err))
			}
		})
	}
}

func TestLocalStorage_Exists(t *testing.T) {
	t.Parallel()

	s, tempDir := setupTestStorage(t)
	ctx := context.Background()

	// Create a test file
	testPath := "test-exists.txt"
	fullPath := filepath.Join(tempDir, testPath)
	err := os.WriteFile(fullPath, []byte("exists"), 0o644)
	require.NoError(t, err)

	tests := []struct {
		name       string
		path       string
		wantExists bool
		wantErr    bool
	}{
		{
			name:       "file exists",
			path:       testPath,
			wantExists: true,
			wantErr:    false,
		},
		{
			name:       "file does not exist",
			path:       "nonexistent.txt",
			wantExists: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := s.Exists(ctx, tt.path)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantExists, exists)
		})
	}
}

func TestLocalStorage_URL(t *testing.T) {
	t.Parallel()

	s, _ := setupTestStorage(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		path    string
		wantURL string
		wantErr bool
	}{
		{
			name:    "simple path",
			path:    "test.jpg",
			wantURL: "http://localhost:8080/uploads/test.jpg",
			wantErr: false,
		},
		{
			name:    "nested path",
			path:    "users/avatar/1/image.jpg",
			wantURL: "http://localhost:8080/uploads/users/avatar/1/image.jpg",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := s.URL(ctx, tt.path)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, url)
		})
	}
}

func TestLocalStorage_TemporaryURL(t *testing.T) {
	t.Parallel()

	s, _ := setupTestStorage(t)
	ctx := context.Background()

	// For local storage, temporary URL is the same as regular URL
	url, err := s.TemporaryURL(ctx, "test.jpg", 0)
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/uploads/test.jpg", url)
}

func TestLocalStorage_PutAndGet_Integration(t *testing.T) {
	t.Parallel()

	s, _ := setupTestStorage(t)
	ctx := context.Background()

	// Test put and get cycle
	testPath := "integration/test.bin"
	testContent := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD}

	// Put
	info, err := s.Put(ctx, testPath, bytes.NewReader(testContent), int64(len(testContent)), "application/octet-stream")
	require.NoError(t, err)
	assert.Equal(t, testPath, info.Path)
	assert.Equal(t, int64(len(testContent)), info.Size)

	// Get
	reader, err := s.Get(ctx, testPath)
	require.NoError(t, err)

	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, testContent, content)
	reader.Close() // Close before delete

	// Exists
	exists, err := s.Exists(ctx, testPath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Delete
	err = s.Delete(ctx, testPath)
	require.NoError(t, err)

	// Verify deleted
	exists, err = s.Exists(ctx, testPath)
	require.NoError(t, err)
	assert.False(t, exists)
}
