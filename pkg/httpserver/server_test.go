package httpserver_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/httpserver"
	"go-boilerplate/pkg/logger"
)

func TestNew_Defaults(t *testing.T) {
	t.Parallel()

	l := logger.NewDevelopment()
	s := httpserver.New(l)

	require.NotNil(t, s.App)
	require.NotNil(t, s.Notify())
}

func TestNew_WithOptions(t *testing.T) {
	t.Parallel()

	l := logger.NewDevelopment()
	s := httpserver.New(l,
		httpserver.Port("9999"),
		httpserver.ReadTimeout(10*time.Second),
		httpserver.WriteTimeout(10*time.Second),
		httpserver.ShutdownTimeout(5*time.Second),
		httpserver.Prefork(false),
	)

	require.NotNil(t, s.App)
}
