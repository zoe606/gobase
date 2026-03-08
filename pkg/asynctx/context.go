// Package asynctx provides context utilities for async job execution.
// IMPORTANT: Never use HTTP request context for async jobs!
package asynctx

import (
	"context"
	"time"
)

// DefaultJobTimeout is the default timeout for async jobs.
const DefaultJobTimeout = 5 * time.Minute

// NewJobContext creates a new context for async job execution.
// This creates a fresh context from context.Background() with the specified timeout.
// Use this when you need a completely independent context for the job.
func NewJobContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = DefaultJobTimeout
	}

	return context.WithTimeout(context.Background(), timeout) //nolint:gosec // cancel function returned to caller
}

// FromAsynqContext wraps the asynq-provided context with an additional timeout.
// The asynq context already supports cancellation; this adds a deadline.
// Use this when you want to respect asynq's cancellation but add a timeout.
func FromAsynqContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = DefaultJobTimeout
	}

	return context.WithTimeout(ctx, timeout) //nolint:gosec // cancel function returned to caller
}

// NewBackgroundContext creates a cancelable context without timeout.
// Use this for long-running jobs that should only be canceled manually.
func NewBackgroundContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background()) //nolint:gosec // cancel function returned to caller
}
