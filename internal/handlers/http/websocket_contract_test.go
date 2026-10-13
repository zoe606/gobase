package httphandler

import (
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	"go-boilerplate/config"
	"go-boilerplate/pkg/jwt"
	ws "go-boilerplate/pkg/websocket"
)

func websocketConfig() *config.Config {
	cfg := contractConfig()
	cfg.HTTP.RequestTimeout = 10 * time.Millisecond
	cfg.WebSocket = config.WebSocket{Enabled: true, AllowedOrigins: "https://example.com", HandshakeTimeout: time.Second, ReadTimeout: time.Second, WriteTimeout: 200 * time.Millisecond, PingInterval: 40 * time.Millisecond, MaxMessageSize: 4096, MaxConnections: 8}
	return cfg
}

func websocketToken(t *testing.T, service jwt.Service) string {
	t.Helper()
	token, _, err := service.GenerateAccessToken(7, "socket@example.com", "user", nil)
	require.NoError(t, err)
	return token
}

func websocketDial(t *testing.T, address, token string) (*websocket.Conn, http.Header) {
	t.Helper()
	conn, response, err := (&websocket.Dialer{HandshakeTimeout: time.Second}).DialContext(t.Context(), address, http.Header{"Authorization": {"Bearer " + token}, "Origin": {"https://example.com"}, "X-Request-ID": {"ws-contract"}, "Accept-Encoding": {"gzip"}})
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	t.Cleanup(func() { _ = conn.Close() })
	require.NoError(t, conn.SetReadDeadline(time.Now().Add(3*time.Second)))
	return conn, response.Header
}

func TestEngineWebSocketHandshake(t *testing.T) {
	cfg := websocketConfig()
	service := jwt.New("socket-test-key", time.Hour, time.Hour)
	address, hub, _ := websocketFixture(t, cfg, service)
	token := websocketToken(t, service)
	refresh, _, err := service.GenerateRefreshToken()
	require.NoError(t, err)
	expired := websocketToken(t, jwt.New("socket-test-key", -time.Hour, time.Hour))
	for _, test := range []struct {
		name, token, origin string
		status              int
	}{
		{"missing", "", "https://example.com", 401},
		{"invalid", "bad-token", "https://example.com", 401},
		{"refresh", "Bearer " + refresh, "https://example.com", 401},
		{"expired", "Bearer " + expired, "https://example.com", 401},
		{"origin", "Bearer " + token, "https://bad.example.com", 403},
		{"missing origin", "Bearer " + token, "", 403},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, response, err := (&websocket.Dialer{HandshakeTimeout: time.Second}).DialContext(t.Context(), address, http.Header{"Authorization": {test.token}, "Origin": {test.origin}, "X-Request-ID": {"ws-rejected"}})
			require.Error(t, err)
			require.NotNil(t, response)
			defer response.Body.Close()
			require.Equal(t, test.status, response.StatusCode)
			require.Equal(t, "ws-rejected", response.Header.Get("X-Request-ID"))
		})
	}
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, strings.Replace(address, "ws:", "http:", 1), http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Origin", "https://example.com")
	response, err := (&http.Client{Timeout: time.Second}).Do(req)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.Equal(t, 426, response.StatusCode)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	for _, headers := range []struct{ version, key string }{
		{"12", "dGhlIHNhbXBsZSBub25jZQ=="},
		{"13", "bad-key"},
	} {
		req.Header.Set("Sec-WebSocket-Version", headers.version)
		req.Header.Set("Sec-WebSocket-Key", headers.key)
		response, err = (&http.Client{Timeout: time.Second}).Do(req)
		require.NoError(t, err)
		require.NoError(t, response.Body.Close())
		require.Equal(t, 426, response.StatusCode)
		require.Equal(t, "13", response.Header.Get("Sec-WebSocket-Version"))
	}
	require.Zero(t, hub.Active())
	conn, headers := websocketDial(t, address, token)
	require.Equal(t, "ws-contract", headers.Get("X-Request-ID"))
	require.Empty(t, headers.Get("Content-Encoding"))
	require.Empty(t, headers.Get("Sec-WebSocket-Extensions"))
	var message ws.Message
	require.NoError(t, conn.ReadJSON(&message))
	require.Equal(t, "notification", message.Type)
	require.Equal(t, "connected", message.Data)
	require.Equal(t, "ws-contract", message.RequestID)
	for range 4 {
		time.Sleep(50 * time.Millisecond)
		require.NoError(t, conn.WriteJSON(ws.Message{Type: "echo", Data: "hello", RequestID: "spoofed"}))
		require.NoError(t, conn.ReadJSON(&message))
		require.Equal(t, ws.Message{Type: "echo", Data: "hello", RequestID: "ws-contract"}, message)
	}
	require.NoError(t, conn.Close())
	require.Eventually(t, func() bool { return hub.Active() == 0 }, time.Second, 10*time.Millisecond)
	for range 3 {
		conn, _ := websocketDial(t, address, token)
		require.NoError(t, conn.ReadJSON(&message))
		require.NoError(t, conn.Close())
		require.Eventually(t, func() bool { return hub.Active() == 0 }, time.Second, 10*time.Millisecond)
	}
}

func TestEngineWebSocketLimits(t *testing.T) {
	for _, test := range []struct {
		name string
		kind int
		data string
		code int
	}{
		{"binary", websocket.BinaryMessage, "binary", 1003},
		{"JSON", websocket.TextMessage, "{bad", 1007},
		{"type", websocket.TextMessage, `{"type":"unsupported"}`, 1008},
		{"size", websocket.TextMessage, strings.Repeat("x", 5000), 1009},
	} {
		t.Run(test.name, func(t *testing.T) {
			service := jwt.New("socket-test-key", time.Hour, time.Hour)
			address, hub, _ := websocketFixture(t, websocketConfig(), service)
			conn, _ := websocketDial(t, address, websocketToken(t, service))
			var message ws.Message
			require.NoError(t, conn.ReadJSON(&message))
			require.NoError(t, conn.WriteMessage(test.kind, []byte(test.data)))
			_, _, err := conn.ReadMessage()
			require.True(t, websocket.IsCloseError(err, test.code), "%v", err)
			require.Eventually(t, func() bool { return hub.Active() == 0 }, time.Second, 10*time.Millisecond)
		})
	}
}

func TestEngineWebSocketTimeoutAndShutdown(t *testing.T) {
	service := jwt.New("socket-test-key", time.Hour, time.Hour)
	cfg := websocketConfig()
	address, hub, shutdown := websocketFixture(t, cfg, service)
	conn, _ := websocketDial(t, address, websocketToken(t, service))
	conn.SetPingHandler(func(_ string) error { return nil })
	var message ws.Message
	require.NoError(t, conn.ReadJSON(&message))
	start := time.Now()
	_, _, err := conn.ReadMessage()
	require.Error(t, err)
	require.Less(t, time.Since(start), 2*time.Second)
	require.Eventually(t, func() bool { return hub.Active() == 0 }, time.Second, 10*time.Millisecond)
	conn, _ = websocketDial(t, address, websocketToken(t, service))
	require.NoError(t, conn.ReadJSON(&message))
	shutdown()
	_, _, err = conn.ReadMessage()
	require.True(t, websocket.IsCloseError(err, 1001), "%v", err)
	require.Zero(t, hub.Active())
}

func TestEngineWebSocketSlowClient(t *testing.T) {
	cfg := websocketConfig()
	cfg.WebSocket.MaxMessageSize = 2 << 20
	cfg.WebSocket.ReadTimeout = 10 * time.Second
	service := jwt.New("socket-test-key", time.Hour, time.Hour)
	address, hub, _ := websocketFixture(t, cfg, service)
	conn, _ := websocketDial(t, address, websocketToken(t, service))
	require.NoError(t, conn.UnderlyingConn().(*net.TCPConn).SetReadBuffer(1024))
	var message ws.Message
	require.NoError(t, conn.ReadJSON(&message))
	require.NoError(t, conn.SetWriteDeadline(time.Now().Add(2*time.Second)))
	for range 8 {
		if conn.WriteJSON(ws.Message{Type: "echo", Data: strings.Repeat("a", 1<<20)}) != nil {
			break
		}
	}
	require.Eventually(t, func() bool { return hub.Active() == 0 }, 3*time.Second, 20*time.Millisecond)
}

func TestEngineWebSocketExpiryAndCapacity(t *testing.T) {
	service := jwt.New("socket-test-key", 2*time.Second, time.Hour)
	cfg := websocketConfig()
	cfg.WebSocket.MaxConnections = 1
	address, hub, _ := websocketFixture(t, cfg, service)
	token := websocketToken(t, service)
	conn, _ := websocketDial(t, address, token)
	var message ws.Message
	require.NoError(t, conn.ReadJSON(&message))
	_, response, err := (&websocket.Dialer{HandshakeTimeout: time.Second}).DialContext(t.Context(), address, http.Header{"Authorization": {"Bearer " + token}, "Origin": {"https://example.com"}})
	require.Error(t, err)
	require.Equal(t, 503, response.StatusCode)
	require.NoError(t, response.Body.Close())
	start := time.Now()
	_, _, err = conn.ReadMessage()
	require.Error(t, err)
	require.Less(t, time.Since(start), 3*time.Second)
	require.Eventually(t, func() bool { return hub.Active() == 0 }, time.Second, 10*time.Millisecond)
}

func TestEngineWebSocketDisabled(t *testing.T) {
	cfg := websocketConfig()
	cfg.WebSocket.Enabled = false
	service := jwt.New("socket-test-key", time.Hour, time.Hour)
	address, hub, _ := websocketFixture(t, cfg, service)
	_, response, err := (&websocket.Dialer{HandshakeTimeout: time.Second}).DialContext(t.Context(), address, http.Header{"Authorization": {"Bearer " + websocketToken(t, service)}, "Origin": {"https://example.com"}})
	require.Error(t, err)
	require.Equal(t, 404, response.StatusCode)
	require.NoError(t, response.Body.Close())
	require.Zero(t, hub.Active())
}
