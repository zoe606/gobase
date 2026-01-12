package cache

import (
	"context"
	"time"

	"go-boilerplate/pkg/json"
)

// NoopCache implements Cache with no-op operations.
// Useful for development or when caching is disabled.
type NoopCache struct{}

// NewNoop creates a new no-op cache.
func NewNoop() *NoopCache {
	return &NoopCache{}
}

// Get always returns ErrNotFound.
func (c *NoopCache) Get(_ context.Context, _ string, _ interface{}) error {
	return ErrNotFound
}

// Set does nothing and returns nil.
func (c *NoopCache) Set(_ context.Context, _ string, _ interface{}, _ time.Duration) error {
	return nil
}

// Delete does nothing and returns nil.
func (c *NoopCache) Delete(_ context.Context, _ string) error {
	return nil
}

// Exists always returns false.
func (c *NoopCache) Exists(_ context.Context, _ string) (bool, error) {
	return false, nil
}

// Remember always calls the function since there's no cache.
func (c *NoopCache) Remember(_ context.Context, _ string, _ time.Duration, dest interface{}, fn func() (interface{}, error)) error {
	value, err := fn()
	if err != nil {
		return err
	}

	// Copy value to dest
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}
