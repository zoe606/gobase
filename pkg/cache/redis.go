package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"go-boilerplate/pkg/json"
)

// RedisCache implements Cache using Redis.
type RedisCache struct {
	client *redis.Client
	prefix string
}

// NewRedis creates a new Redis cache.
func NewRedis(client *redis.Client, prefix string) *RedisCache {
	return &RedisCache{
		client: client,
		prefix: prefix,
	}
}

// prefixKey adds the configured prefix to a key.
func (c *RedisCache) prefixKey(key string) string {
	return c.prefix + key
}

// Get retrieves a value from Redis.
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, c.prefixKey(key)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrNotFound
		}
		return err
	}
	return json.Unmarshal(data, dest)
}

// Set stores a value in Redis with expiration.
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.prefixKey(key), data, ttl).Err()
}

// Delete removes a key from Redis.
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.prefixKey(key)).Err()
}

// Exists checks if a key exists in Redis.
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, c.prefixKey(key)).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Remember gets from cache or calls fn and caches the result.
func (c *RedisCache) Remember(ctx context.Context, key string, ttl time.Duration, dest interface{}, fn func() (interface{}, error)) error {
	// Try to get from cache first
	err := c.Get(ctx, key, dest)
	if err == nil {
		return nil // Cache hit
	}
	if !errors.Is(err, ErrNotFound) {
		return err // Real error
	}

	// Cache miss - call the function
	value, err := fn()
	if err != nil {
		return err
	}

	// Cache the result (best effort - don't fail if caching fails)
	_ = c.Set(ctx, key, value, ttl)

	// Copy value to dest
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}
