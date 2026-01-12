// Package resilience provides patterns for building resilient applications.
package resilience

import (
	"context"
	"fmt"

	"github.com/sony/gobreaker/v2"
)

// CircuitBreaker wraps sony/gobreaker to provide circuit breaker functionality.
// It protects external service calls by failing fast when the service is unhealthy.
type CircuitBreaker struct {
	cb *gobreaker.CircuitBreaker[any]
}

// New creates a new CircuitBreaker with the given configuration.
func New(cfg Config) *CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        cfg.Name,
		MaxRequests: cfg.MaxRequests,
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < cfg.MinRequests {
				return false
			}
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRatio >= cfg.FailureRatio
		},
		OnStateChange: cfg.OnStateChange,
	}

	return &CircuitBreaker{
		cb: gobreaker.NewCircuitBreaker[any](settings),
	}
}

// Execute runs the given function within the circuit breaker.
// If the circuit is open, it returns ErrCircuitOpen immediately.
// If the function fails, the failure is recorded and may trip the circuit.
func (c *CircuitBreaker) Execute(fn func() (any, error)) (any, error) {
	result, err := c.cb.Execute(fn)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ExecuteWithContext runs the given function within the circuit breaker with context support.
// It respects context cancellation and deadlines.
func (c *CircuitBreaker) ExecuteWithContext(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
	// Check context before executing
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return c.cb.Execute(func() (any, error) {
		// Check context again inside the circuit breaker
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return fn(ctx)
	})
}

// State returns the current state of the circuit breaker.
func (c *CircuitBreaker) State() State {
	// gobreaker.State is an int, our State is uint32
	// Both represent the same 3 states (0, 1, 2), so this is safe
	state := c.cb.State()
	switch state {
	case 0:
		return StateClosed
	case 1:
		return StateHalfOpen
	case 2:
		return StateOpen
	default:
		return StateClosed // Fallback to closed for unknown states
	}
}

// Name returns the name of the circuit breaker.
func (c *CircuitBreaker) Name() string {
	return c.cb.Name()
}

// Counts returns the current counts of the circuit breaker.
func (c *CircuitBreaker) Counts() Counts {
	counts := c.cb.Counts()
	return Counts{
		Requests:             counts.Requests,
		TotalSuccesses:       counts.TotalSuccesses,
		TotalFailures:        counts.TotalFailures,
		ConsecutiveSuccesses: counts.ConsecutiveSuccesses,
		ConsecutiveFailures:  counts.ConsecutiveFailures,
	}
}

// State represents the state of a circuit breaker.
type State uint32

const (
	// StateClosed means the circuit is closed and requests flow normally.
	StateClosed State = iota
	// StateHalfOpen means the circuit is testing if the service has recovered.
	StateHalfOpen
	// StateOpen means the circuit is open and requests fail fast.
	StateOpen
)

// String returns the string representation of the state.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// Counts holds the counts of requests and their outcomes.
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}
