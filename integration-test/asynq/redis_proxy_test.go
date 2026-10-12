package asynq_test

import (
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type redisProxy struct {
	listener    net.Listener
	upstream    string
	mu          sync.Mutex
	enabled     bool
	connections map[net.Conn]struct{}
	wg          sync.WaitGroup
}

func newRedisProxy(t *testing.T, upstream string) *redisProxy {
	t.Helper()
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	p := &redisProxy{listener: listener, upstream: upstream, enabled: true, connections: make(map[net.Conn]struct{})}
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for {
			client, err := listener.Accept()
			if err != nil {
				return
			}
			p.wg.Add(1)
			go func() { defer p.wg.Done(); p.forward(t, client) }()
		}
	}()
	t.Cleanup(func() { p.available(false); _ = listener.Close(); p.wg.Wait() })
	return p
}

func (p *redisProxy) addr() string { return p.listener.Addr().String() }

func (p *redisProxy) available(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = enabled
	if !enabled {
		for conn := range p.connections {
			_ = conn.Close()
		}
	}
}

func (p *redisProxy) forward(t *testing.T, client net.Conn) {
	defer func() { _ = client.Close() }()
	upstream, err := (&net.Dialer{Timeout: time.Second}).DialContext(t.Context(), "tcp", p.upstream)
	if err != nil {
		return
	}
	defer func() { _ = upstream.Close() }()
	p.mu.Lock()
	if !p.enabled {
		p.mu.Unlock()
		return
	}
	p.connections[client] = struct{}{}
	p.connections[upstream] = struct{}{}
	p.mu.Unlock()
	defer func() { p.mu.Lock(); delete(p.connections, client); delete(p.connections, upstream); p.mu.Unlock() }()
	done := make(chan struct{})
	go func() { _, _ = io.Copy(upstream, client); _ = upstream.Close(); _ = client.Close(); close(done) }()
	_, _ = io.Copy(client, upstream)
	_ = upstream.Close()
	_ = client.Close()
	<-done
}
