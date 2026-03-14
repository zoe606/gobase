// Package cache provides a caching abstraction with multiple backend implementations.
package cache

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a key is not found in the cache.
var ErrNotFound = errors.New("cache: key not found")

// Cache defines the caching interface.
type Cache interface {
	// Get retrieves a value by key. Returns ErrNotFound if key doesn't exist.
	Get(ctx context.Context, key string, dest interface{}) error

	// Set stores a value with expiration.
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a key.
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists.
	Exists(ctx context.Context, key string) (bool, error)

	// Remember gets from cache or calls fn and caches the result.
	// If the key exists, it returns the cached value.
	// If not, it calls fn, caches the result with ttl, and returns it.
	Remember(ctx context.Context, key string, ttl time.Duration, dest interface{}, fn func() (interface{}, error)) error

	// DeleteByPrefix removes all keys matching the given prefix.
	DeleteByPrefix(ctx context.Context, prefix string) error
}
