package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Storage implements Provider for S3/MinIO storage.
type S3Storage struct {
	client   *minio.Client
	bucket   string
	endpoint string
	useSSL   bool
}

// NewS3Storage creates a new S3/MinIO storage provider.
func NewS3Storage(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*S3Storage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	return &S3Storage{
		client:   client,
		bucket:   bucket,
		endpoint: endpoint,
		useSSL:   useSSL,
	}, nil
}

// EnsureBucket creates the bucket if it doesn't exist.
func (s *S3Storage) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket exists: %w", err)
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
	}
	return nil
}

// Put stores a file in S3/MinIO.
func (s *S3Storage) Put(ctx context.Context, path string, reader io.Reader, size int64, mimeType string) (*FileInfo, error) {
	opts := minio.PutObjectOptions{
		ContentType: mimeType,
	}

	info, err := s.client.PutObject(ctx, s.bucket, path, reader, size, opts)
	if err != nil {
		return nil, fmt.Errorf("put object: %w", err)
	}

	return &FileInfo{
		Path:     path,
		Size:     info.Size,
		MimeType: mimeType,
		Hash:     info.ETag,
	}, nil
}

// Get retrieves a file from S3/MinIO.
func (s *S3Storage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}

	// Verify object exists by checking stat.
	_, err = obj.Stat()
	if err != nil {
		obj.Close()
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("stat object: %w", err)
	}

	return obj, nil
}

// Delete removes a file from S3/MinIO.
func (s *S3Storage) Delete(ctx context.Context, path string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, path, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("remove object: %w", err)
	}
	return nil
}

// Exists checks if a file exists in S3/MinIO.
func (s *S3Storage) Exists(ctx context.Context, path string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("stat object: %w", err)
	}
	return true, nil
}

// URL returns a public URL for the file.
func (s *S3Storage) URL(ctx context.Context, path string) (string, error) {
	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}
	u := &url.URL{
		Scheme: scheme,
		Host:   s.endpoint,
		Path:   fmt.Sprintf("/%s/%s", s.bucket, path),
	}
	return u.String(), nil
}

// TemporaryURL returns a presigned URL for reading the file.
func (s *S3Storage) TemporaryURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucket, path, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("presigned get: %w", err)
	}
	return presignedURL.String(), nil
}

// PresignedUploadURL returns a presigned URL for direct client uploads.
func (s *S3Storage) PresignedUploadURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	presignedURL, err := s.client.PresignedPutObject(ctx, s.bucket, path, expiry)
	if err != nil {
		return "", fmt.Errorf("presigned put: %w", err)
	}
	return presignedURL.String(), nil
}
