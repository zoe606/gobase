package httpserver_test

import (
	"context"
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

func TestShutdownHooksShareBudget(t *testing.T) {
	var deadlines []time.Time
	s := httpserver.New(logger.New("error"), httpserver.Port("0"), httpserver.ShutdownTimeout(30*time.Millisecond),
		httpserver.OnShutdown(func(ctx context.Context) error {
			deadline, ok := ctx.Deadline()
			require.True(t, ok)
			deadlines = append(deadlines, deadline)
			<-ctx.Done()
			return ctx.Err()
		}),
		httpserver.OnShutdown(func(ctx context.Context) error {
			deadline, _ := ctx.Deadline()
			deadlines = append(deadlines, deadline)
			require.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
			return nil
		}),
	)
	s.Start()
	start := time.Now()
	require.ErrorIs(t, s.Shutdown(), context.DeadlineExceeded)
	require.Len(t, deadlines, 2)
	require.Equal(t, deadlines[0], deadlines[1])
	require.Less(t, time.Since(start), time.Second)
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
		httpserver.BodyLimit(2*1024*1024),
	)

	require.NotNil(t, s.App)
}
