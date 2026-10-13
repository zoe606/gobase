package echo_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/usecase/echo"
)

func TestEcho(t *testing.T) {
	u := echo.New()
	for _, message := range []string{"hello", " halo 世界 ", strings.Repeat("a", 4096)} {
		result, err := u.Echo(t.Context(), message)
		require.NoError(t, err)
		require.Equal(t, message, result)
	}
	for _, message := range []string{"", " \t\n", strings.Repeat("a", 4097), string([]byte{0xff})} {
		_, err := u.Echo(t.Context(), message)
		require.ErrorIs(t, err, echo.ErrInvalidMessage)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := u.Echo(ctx, "hello")
	require.ErrorIs(t, err, context.Canceled)
}
