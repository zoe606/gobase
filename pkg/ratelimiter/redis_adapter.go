package ratelimiter

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisAdapter adapts a go-redis client to the Storage interface.
type RedisAdapter struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisAdapter creates a new Redis adapter.
func NewRedisAdapter(client *redis.Client) *RedisAdapter {
	return &RedisAdapter{
		client: client,
		ctx:    context.Background(),
	}
}

// Get retrieves a value by key.
// Returns nil, nil when the key does not exist.
func (a *RedisAdapter) Get(key string) ([]byte, error) {
	val, err := a.client.Get(a.ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}

	return val, err
}

// Set stores a value with expiration.
func (a *RedisAdapter) Set(key string, val []byte, exp time.Duration) error {
	return a.client.Set(a.ctx, key, val, exp).Err()
}

// Delete removes a key.
func (a *RedisAdapter) Delete(key string) error {
	return a.client.Del(a.ctx, key).Err()
}

// Reset flushes the database.
func (a *RedisAdapter) Reset() error {
	return a.client.FlushDB(a.ctx).Err()
}

// Close closes the Redis connection.
func (a *RedisAdapter) Close() error {
	return a.client.Close()
}
