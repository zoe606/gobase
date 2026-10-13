package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWebSocketValidation(t *testing.T) {
	valid := WebSocket{Enabled: true, AllowedOrigins: "https://example.com, http://localhost:3000", HandshakeTimeout: time.Second, ReadTimeout: time.Minute, WriteTimeout: time.Second, PingInterval: 20 * time.Second, MaxMessageSize: 1024, MaxConnections: 8}
	require.Empty(t, valid.validationErrors())
	require.True(t, valid.AllowsOrigin("https://example.com"))
	require.False(t, valid.AllowsOrigin("https://bad.example.com"))
	require.False(t, valid.AllowsOrigin(""))
	require.Empty(t, (WebSocket{}).validationErrors())
	for _, origin := range []string{"", "*", "null", "https://*.example.com", "https://example.com/", "https://user:pass@example.com", "https://example.com?q=x", "https://example.com#x"} {
		t.Run(origin, func(t *testing.T) {
			cfg := valid
			cfg.AllowedOrigins = origin
			require.NotEmpty(t, cfg.validationErrors())
		})
	}
	invalid := valid
	invalid.PingInterval = valid.ReadTimeout
	require.NotEmpty(t, invalid.validationErrors())
	invalid = valid
	invalid.MaxConnections = 0
	require.NotEmpty(t, invalid.validationErrors())
}

func TestWebSocketEnvironment(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "localhost")
	t.Setenv("POSTGRES_DBNAME", "testdb")
	t.Setenv("WEBSOCKET_ENABLED", "false")
	cfg, err := NewConfig("")
	require.NoError(t, err)
	require.False(t, cfg.WebSocket.Enabled)
	require.Equal(t, int64(65536), cfg.WebSocket.MaxMessageSize)
	t.Setenv("WEBSOCKET_ENABLED", "true")
	t.Setenv("WEBSOCKET_ALLOWED_ORIGINS", "https://example.com")
	t.Setenv("WEBSOCKET_MAX_CONNECTIONS", "2")
	cfg, err = NewConfig("")
	require.NoError(t, err)
	require.True(t, cfg.WebSocket.Enabled)
	require.Equal(t, 2, cfg.WebSocket.MaxConnections)
	require.Equal(t, 25*time.Second, cfg.WebSocket.PingInterval)
	t.Setenv("WEBSOCKET_ALLOWED_ORIGINS", "*")
	_, err = NewConfig("")
	require.ErrorContains(t, err, "WEBSOCKET_ALLOWED_ORIGINS")
}
