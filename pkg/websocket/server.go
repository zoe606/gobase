package websocket

import (
	"context"
	"encoding/binary"
	"sync"
	"time"

	"go-boilerplate/pkg/logger"
)

const (
	TextMessage  = 1
	CloseMessage = 8
	PingMessage  = 9
)

type Connection interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(int, []byte) error
	WriteControl(int, []byte, time.Time) error
	SetReadLimit(int64)
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
	SetPongHandler(func(string) error)
	Close() error
}

type Options struct {
	HandshakeTimeout time.Duration
	MaxMessageSize   int64
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	PingInterval     time.Duration
	MaxConnections   int
}

type Server struct {
	options          Options
	l                logger.Interface
	mu               sync.Mutex
	sessions         map[*Session]struct{}
	closing          bool
	shutdownDeadline time.Time
	wg               sync.WaitGroup
}

type Session struct {
	server  *Server
	conn    Connection
	once    sync.Once
	timer   *time.Timer
	expired bool
}

func New(options Options, l logger.Interface) *Server {
	return &Server{options: options, l: l, sessions: make(map[*Session]struct{})}
}

// Reserve tracks an upgrade before a transport can start its connection callback.
func (s *Server) Reserve() (*Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closing || len(s.sessions) >= s.options.MaxConnections {
		return nil, false
	}
	session := &Session{server: s}
	s.sessions[session] = struct{}{}
	s.wg.Add(1)
	if s.options.HandshakeTimeout > 0 {
		session.timer = time.AfterFunc(s.options.HandshakeTimeout, session.expirePending)
	}
	return session, true
}

func (s *Session) expirePending() {
	s.server.mu.Lock()
	if s.conn != nil {
		s.server.mu.Unlock()
		return
	}
	s.expired = true
	s.server.mu.Unlock()
	s.Release()
}

func (s *Session) Release() {
	s.once.Do(func() {
		if s.timer != nil {
			s.timer.Stop()
		}
		s.server.mu.Lock()
		delete(s.server.sessions, s)
		s.server.mu.Unlock()
		s.server.wg.Done()
	})
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	if !s.closing {
		s.closing = true
		s.shutdownDeadline, _ = ctx.Deadline()
	}
	shutdownDeadline := s.shutdownDeadline
	connections := make([]Connection, 0, len(s.sessions))
	for session := range s.sessions {
		if session.conn != nil {
			connections = append(connections, session.conn)
		}
	}
	s.mu.Unlock()
	for _, conn := range connections {
		deadline := time.Now().Add(s.options.WriteTimeout)
		if !shutdownDeadline.IsZero() && shutdownDeadline.Before(deadline) {
			deadline = shutdownDeadline
		}
		closeConnection(conn, 1001, "server shutdown", deadline)
	}
	done := make(chan struct{})
	go func() { s.wg.Wait(); close(done) }()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.mu.Lock()
		var pending []*Session
		for session := range s.sessions {
			if session.conn == nil {
				pending = append(pending, session)
			}
		}
		s.mu.Unlock()
		for _, session := range pending {
			session.Release()
		}
		return ctx.Err()
	}
}

func (s *Server) Active() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.sessions)
}

func closeConnection(conn Connection, code uint16, reason string, deadline time.Time) {
	payload := append(binary.BigEndian.AppendUint16(nil, code), []byte(reason)...)
	_ = conn.WriteControl(CloseMessage, payload, deadline)
	_ = conn.Close()
}
