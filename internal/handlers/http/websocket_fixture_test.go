package httphandler

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	"go-boilerplate/config"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
	ws "go-boilerplate/pkg/websocket"
)

func websocketFixture(t *testing.T, cfg *config.Config, service jwt.Service) (url string, websocketServer *ws.Server, shutdownHTTP func()) {
	t.Helper()
	l := logger.New("error")
	hub := ws.New(ws.Options{HandshakeTimeout: cfg.WebSocket.HandshakeTimeout, MaxMessageSize: cfg.WebSocket.MaxMessageSize, ReadTimeout: cfg.WebSocket.ReadTimeout, WriteTimeout: cfg.WebSocket.WriteTimeout, PingInterval: cfg.WebSocket.PingInterval, MaxConnections: cfg.WebSocket.MaxConnections}, l)
	app := fiber.New(fiber.Config{DisableStartupMessage: true, ReadTimeout: time.Second, IdleTimeout: time.Second})
	setupMiddleware(app, cfg, l, nil, nil, service)
	setupOptionalFeatures(app, cfg)
	SetupWebSocketRoutes(app, cfg.WebSocket, service, hub)
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	done := make(chan struct{})
	go func() { defer close(done); _ = app.Listener(listener) }()
	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		require.NoError(t, hub.Shutdown(ctx))
		require.NoError(t, app.ShutdownWithContext(ctx))
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("HTTP server did not stop")
		}
	}
	t.Cleanup(shutdown)
	return "ws://" + listener.Addr().String() + "/v1/ws", hub, shutdown
}
