package webapi_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/repo/webapi"
	"go-boilerplate/pkg/resilience"
)

func TestNew(t *testing.T) {
	t.Parallel()

	api := webapi.New()
	require.NotNil(t, api)
	require.Equal(t, resilience.StateClosed, api.CircuitState())
}

func TestNewWithCircuitBreaker(t *testing.T) {
	t.Parallel()

	cfg := resilience.DefaultConfig("test-translate")
	api := webapi.NewWithCircuitBreaker(cfg)
	require.NotNil(t, api)
	require.Equal(t, resilience.StateClosed, api.CircuitState())
}
