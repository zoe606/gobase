package resilience_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sony/gobreaker/v2"
	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/resilience"
)

func TestCircuitBreaker_Execute_Success(t *testing.T) {
	t.Parallel()

	cb := resilience.New(resilience.DefaultConfig("test"))

	result, err := cb.Execute(func() (any, error) {
		return "success", nil
	})

	require.NoError(t, err)
	require.Equal(t, "success", result)
}

func TestCircuitBreaker_Execute_Failure(t *testing.T) {
	t.Parallel()

	cb := resilience.New(resilience.DefaultConfig("test"))
	expectedErr := errors.New("service error")

	result, err := cb.Execute(func() (any, error) {
		return nil, expectedErr
	})

	require.Error(t, err)
	require.Nil(t, result)
}

func TestCircuitBreaker_TripsOnFailureRatio(t *testing.T) {
	t.Parallel()

	// Configure circuit to trip after 5 requests with 50% failure ratio
	cfg := resilience.Config{
		Name:         "test",
		MaxRequests:  1,
		Interval:     time.Minute,
		Timeout:      time.Second,
		FailureRatio: 0.5,
		MinRequests:  4,
	}
	cb := resilience.New(cfg)

	// Fail all requests to trip the circuit
	for i := 0; i < 5; i++ {
		_, _ = cb.Execute(func() (any, error) {
			return nil, errors.New("failure")
		})
	}

	// Circuit should be open now
	require.Equal(t, resilience.StateOpen, cb.State())

	// Requests should fail immediately
	_, err := cb.Execute(func() (any, error) {
		return "should not run", nil
	})
	require.Error(t, err)
}

func TestCircuitBreaker_ExecuteWithContext_Success(t *testing.T) {
	t.Parallel()

	cb := resilience.New(resilience.DefaultConfig("test"))
	ctx := context.Background()

	result, err := cb.ExecuteWithContext(ctx, func(ctx context.Context) (any, error) {
		return "success", nil
	})

	require.NoError(t, err)
	require.Equal(t, "success", result)
}

func TestCircuitBreaker_ExecuteWithContext_CancelledContext(t *testing.T) {
	t.Parallel()

	cb := resilience.New(resilience.DefaultConfig("test"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := cb.ExecuteWithContext(ctx, func(ctx context.Context) (any, error) {
		return "should not run", nil
	})

	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
}

func TestCircuitBreaker_State(t *testing.T) {
	t.Parallel()

	cb := resilience.New(resilience.DefaultConfig("test"))

	// Initial state should be closed
	require.Equal(t, resilience.StateClosed, cb.State())
}

func TestCircuitBreaker_Name(t *testing.T) {
	t.Parallel()

	cb := resilience.New(resilience.DefaultConfig("my-service"))

	require.Equal(t, "my-service", cb.Name())
}

func TestCircuitBreaker_Counts(t *testing.T) {
	t.Parallel()

	cb := resilience.New(resilience.DefaultConfig("test"))

	// Execute some successful requests
	for i := 0; i < 3; i++ {
		_, _ = cb.Execute(func() (any, error) {
			return "success", nil
		})
	}

	// Execute some failed requests
	for i := 0; i < 2; i++ {
		_, _ = cb.Execute(func() (any, error) {
			return nil, errors.New("failure")
		})
	}

	counts := cb.Counts()
	require.Equal(t, uint32(5), counts.Requests)
	require.Equal(t, uint32(3), counts.TotalSuccesses)
	require.Equal(t, uint32(2), counts.TotalFailures)
}

func TestState_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		state    resilience.State
		expected string
	}{
		{resilience.StateClosed, "closed"},
		{resilience.StateHalfOpen, "half-open"},
		{resilience.StateOpen, "open"},
		{resilience.State(99), "unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.expected, tt.state.String())
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := resilience.DefaultConfig("test-service")

	require.Equal(t, "test-service", cfg.Name)
	require.Equal(t, uint32(3), cfg.MaxRequests)
	require.Equal(t, 60*time.Second, cfg.Interval)
	require.Equal(t, 30*time.Second, cfg.Timeout)
	require.Equal(t, 0.5, cfg.FailureRatio)
	require.Equal(t, uint32(10), cfg.MinRequests)
}

func TestConfig_WithMethods(t *testing.T) {
	t.Parallel()

	cfg := resilience.DefaultConfig("test").
		WithMaxRequests(5).
		WithInterval(30 * time.Second).
		WithTimeout(15 * time.Second).
		WithFailureRatio(0.6).
		WithMinRequests(20).
		WithOnStateChange(func(_ string, _, _ gobreaker.State) {})

	require.Equal(t, uint32(5), cfg.MaxRequests)
	require.Equal(t, 30*time.Second, cfg.Interval)
	require.Equal(t, 15*time.Second, cfg.Timeout)
	require.Equal(t, 0.6, cfg.FailureRatio)
	require.Equal(t, uint32(20), cfg.MinRequests)
	require.NotNil(t, cfg.OnStateChange)
}
