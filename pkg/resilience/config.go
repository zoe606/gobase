package resilience

import (
	"time"

	"github.com/sony/gobreaker/v2"
)

// Config holds the configuration for a CircuitBreaker.
type Config struct {
	// Name is the name of the circuit breaker (used for logging/metrics).
	Name string

	// MaxRequests is the maximum number of requests allowed to pass through
	// when the circuit breaker is half-open. If MaxRequests is 0, only 1
	// request is allowed.
	MaxRequests uint32

	// Interval is the cyclic period of the closed state for the circuit breaker
	// to clear the internal counts. If Interval is 0, the counts are never cleared.
	Interval time.Duration

	// Timeout is the period of the open state, after which the state becomes half-open.
	// If Timeout is 0, the default value of 60 seconds is used.
	Timeout time.Duration

	// FailureRatio is the failure ratio (0.0-1.0) at which the circuit will trip.
	// For example, 0.5 means the circuit trips when 50% of requests fail.
	FailureRatio float64

	// MinRequests is the minimum number of requests needed before the failure
	// ratio is evaluated. This prevents the circuit from tripping on the first
	// few failures.
	MinRequests uint32

	// OnStateChange is called when the circuit breaker changes state.
	// This can be used for logging or metrics.
	OnStateChange func(name string, from gobreaker.State, to gobreaker.State)
}

// DefaultConfig returns a Config with sensible defaults for the given name.
// Defaults:
//   - MaxRequests: 3 (requests allowed in half-open state)
//   - Interval: 60s (reset interval in closed state)
//   - Timeout: 30s (time in open state before half-open)
//   - FailureRatio: 0.5 (50% failures to trip)
//   - MinRequests: 10 (minimum requests before evaluating)
func DefaultConfig(name string) Config {
	return Config{
		Name:         name,
		MaxRequests:  3,
		Interval:     60 * time.Second,
		Timeout:      30 * time.Second,
		FailureRatio: 0.5,
		MinRequests:  10,
	}
}

// WithMaxRequests returns a new Config with the MaxRequests field set.
func (c Config) WithMaxRequests(maxRequests uint32) Config {
	c.MaxRequests = maxRequests
	return c
}

// WithInterval returns a new Config with the Interval field set.
func (c Config) WithInterval(interval time.Duration) Config {
	c.Interval = interval
	return c
}

// WithTimeout returns a new Config with the Timeout field set.
func (c Config) WithTimeout(timeout time.Duration) Config {
	c.Timeout = timeout
	return c
}

// WithFailureRatio returns a new Config with the FailureRatio field set.
func (c Config) WithFailureRatio(ratio float64) Config {
	c.FailureRatio = ratio
	return c
}

// WithMinRequests returns a new Config with the MinRequests field set.
func (c Config) WithMinRequests(minRequests uint32) Config {
	c.MinRequests = minRequests
	return c
}

// WithOnStateChange returns a new Config with the OnStateChange callback set.
func (c Config) WithOnStateChange(fn func(name string, from gobreaker.State, to gobreaker.State)) Config {
	c.OnStateChange = fn
	return c
}
