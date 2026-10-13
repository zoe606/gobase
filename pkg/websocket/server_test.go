package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	transport "github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/logger"
)

func TestShutdownDuringPendingUpgrade(t *testing.T) {
	server := New(Options{MaxConnections: 1, WriteTimeout: time.Millisecond}, logger.New("error"))
	session, ok := server.Reserve()
	require.True(t, ok)
	_, ok = server.Reserve()
	require.False(t, ok)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()
	require.ErrorIs(t, server.Shutdown(ctx), context.DeadlineExceeded)
	require.Zero(t, server.Active())
	_, ok = server.Reserve()
	require.False(t, ok)
	session.Release()
	require.NoError(t, server.Shutdown(t.Context()))
}

func TestPendingUpgradeExpires(t *testing.T) {
	server := New(Options{MaxConnections: 1, HandshakeTimeout: 20 * time.Millisecond}, logger.New("error"))
	session, ok := server.Reserve()
	require.True(t, ok)
	require.Eventually(t, func() bool { return server.Active() == 0 }, time.Second, time.Millisecond)
	session.Release()
	session, ok = server.Reserve()
	require.True(t, ok)
	session.Release()
	require.NoError(t, server.Shutdown(t.Context()))
}

type panicConnection struct {
	*transport.Conn
	writer bool
}

func (c *panicConnection) ReadMessage() (kind int, data []byte, err error) {
	kind, data, err = c.Conn.ReadMessage()
	if !c.writer && err == nil {
		panic("private panic value")
	}
	return kind, data, err
}

func (c *panicConnection) WriteMessage(kind int, data []byte) error {
	if c.writer {
		panic("private panic value")
	}
	return c.Conn.WriteMessage(kind, data)
}

func TestSessionRecoveryAndCleanup(t *testing.T) {
	for _, writer := range []bool{false, true} {
		name := "reader"
		if writer {
			name = "writer"
		}
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "session.log")
			file, err := os.Create(path)
			require.NoError(t, err)
			t.Cleanup(func() { _ = file.Close() })
			hub := New(Options{MaxConnections: 1, MaxMessageSize: 4096, ReadTimeout: time.Second, WriteTimeout: time.Second, PingInterval: 50 * time.Millisecond}, logger.New("info", logger.WithFileOutput(file)))
			done := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer close(done)
				session, ok := hub.Reserve()
				if !ok {
					http.Error(w, "unavailable", http.StatusServiceUnavailable)
					return
				}
				defer session.Release()
				conn, upgradeErr := (&transport.Upgrader{}).Upgrade(w, r, nil)
				if upgradeErr != nil {
					return
				}
				session.Serve(&panicConnection{Conn: conn, writer: writer}, "recovery-test", 7, time.Now().Add(time.Minute))
			}))
			defer server.Close()
			conn, response, err := transport.DefaultDialer.DialContext(t.Context(), strings.Replace(server.URL, "http", "ws", 1), nil)
			require.NoError(t, err)
			defer conn.Close()
			require.NoError(t, response.Body.Close())
			require.NoError(t, conn.SetReadDeadline(time.Now().Add(2*time.Second)))
			if !writer {
				_, _, err = conn.ReadMessage()
				require.NoError(t, err)
				require.NoError(t, conn.WriteMessage(transport.TextMessage, []byte("trigger")))
			}
			_, _, err = conn.ReadMessage()
			var closeErr *transport.CloseError
			require.ErrorAs(t, err, &closeErr)
			require.Equal(t, 1011, closeErr.Code)
			require.Equal(t, "internal error", closeErr.Text)
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("session did not stop")
			}
			require.Zero(t, hub.Active())
			logs, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Contains(t, string(logs), "request_id=recovery-test")
			require.Contains(t, string(logs), "private panic value")
			require.Contains(t, string(logs), "WebSocket disconnected")
		})
	}
	require.Eventually(t, func() bool {
		stacks := make([]byte, 1<<20)
		n := runtime.Stack(stacks, true)
		return !strings.Contains(string(stacks[:n]), "websocket.(*Session).Serve")
	}, time.Second, 10*time.Millisecond)
}
