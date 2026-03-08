// Package redis provides a Redis client wrapper with connection management.
package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps the Redis client with additional functionality.
type Client struct {
	*redis.Client
}

// Config holds Redis configuration.
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// New creates a new Redis client.
func New(cfg Config) (*Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &Client{Client: client}, nil
}

// NewWithRetry creates a new Redis client with retry logic.
func NewWithRetry(cfg Config, attempts int, delay time.Duration) (*Client, error) {
	var client *Client
	var err error

	for i := 0; i < attempts; i++ {
		client, err = New(cfg)
		if err == nil {
			return client, nil
		}

		time.Sleep(delay)
	}

	return nil, fmt.Errorf("redis connection failed after %d attempts: %w", attempts, err)
}

// Ping checks the connection to Redis.
func (c *Client) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}

// Close closes the Redis connection.
func (c *Client) Close() error {
	return c.Client.Close()
}

// SetJSON stores a value as JSON with expiration.
func (c *Client) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.Set(ctx, key, value, expiration).Err()
}

// GetJSON retrieves a value and unmarshals it.
func (c *Client) GetJSON(ctx context.Context, key string) (string, error) {
	return c.Get(ctx, key).Result()
}

// Delete removes a key.
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	return c.Del(ctx, keys...).Err()
}

// Exists checks if a key exists.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// SetNX sets a value only if the key does not exist.
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	result, err := c.Client.SetArgs(ctx, key, value, redis.SetArgs{
		Mode: "NX",
		TTL:  expiration,
	}).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return result == "OK", nil
}
