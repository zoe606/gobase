// Package ratelimiter provides rate limiter storage backends.
package ratelimiter

import "time"

// Storage defines the interface for rate limiter backends.
// This matches Fiber's fiber.Storage interface for testability.
type Storage interface {
	Get(key string) ([]byte, error)
	Set(key string, val []byte, exp time.Duration) error
	Delete(key string) error
	Reset() error
	Close() error
}

// RedisStore implements Fiber's Storage interface backed by a Redis-compatible backend.
type RedisStore struct {
	backend Storage
}

// NewRedisStore creates a new RedisStore wrapping the given storage backend.
func NewRedisStore(backend Storage) *RedisStore {
	return &RedisStore{backend: backend}
}

// Get retrieves a value by key.
func (s *RedisStore) Get(key string) ([]byte, error) {
	return s.backend.Get(key)
}

// Set stores a value with expiration.
func (s *RedisStore) Set(key string, val []byte, exp time.Duration) error {
	return s.backend.Set(key, val, exp)
}

// Delete removes a key.
func (s *RedisStore) Delete(key string) error {
	return s.backend.Delete(key)
}

// Reset clears all keys.
func (s *RedisStore) Reset() error {
	return s.backend.Reset()
}

// Close closes the storage connection.
func (s *RedisStore) Close() error {
	return s.backend.Close()
}
