// Package asynq provides a wrapper around the Asynq task queue.
package asynq

import (
	"fmt"

	"github.com/hibiken/asynq"
)

// Client wraps the Asynq client for enqueueing tasks.
type Client struct {
	*asynq.Client
}

// Config holds Asynq configuration.
type Config struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

// NewClient creates a new Asynq client.
func NewClient(cfg Config) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	return &Client{Client: client}
}

// EnqueueTask enqueues a task with default options.
func (c *Client) EnqueueTask(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.Enqueue(task, opts...)
}

// Close closes the Asynq client.
func (c *Client) Close() error {
	return c.Client.Close()
}

// NewTask creates a new Asynq task.
func NewTask(typename string, payload []byte, opts ...asynq.Option) *asynq.Task {
	return asynq.NewTask(typename, payload, opts...)
}

// RedisAddr returns the Redis address from config.
func RedisAddr(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}
