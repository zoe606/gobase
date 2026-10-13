package websocket

import (
	"time"
	"unicode/utf8"

	"go-boilerplate/pkg/json"
)

type Message struct {
	Type      string `json:"type"`
	Data      string `json:"data"`
	RequestID string `json:"request_id,omitempty"`
}

func (s *Session) Serve(conn Connection, requestID string, userID uint, expiresAt time.Time) {
	defer s.Release()
	defer func() { _ = conn.Close() }()
	defer s.recoverConnection(conn, requestID)
	s.server.mu.Lock()
	s.conn = conn
	if s.timer != nil {
		s.timer.Stop()
	}
	expired := s.expired
	closing, deadline := s.server.closing, s.server.shutdownDeadline
	s.server.mu.Unlock()
	if expired {
		closeConnection(conn, 1013, "handshake timeout", time.Now().Add(s.server.options.WriteTimeout))
		return
	}
	if closing {
		if deadline.IsZero() {
			deadline = time.Now()
		}
		closeConnection(conn, 1001, "server shutdown", deadline)
		return
	}
	options := s.server.options
	conn.SetReadLimit(options.MaxMessageSize)
	readDeadline := func() time.Time {
		deadline := time.Now().Add(options.ReadTimeout)
		if expiresAt.Before(deadline) {
			deadline = expiresAt
		}
		return deadline
	}
	if err := conn.SetReadDeadline(readDeadline()); err != nil {
		return
	}
	conn.SetPongHandler(func(_ string) error { return conn.SetReadDeadline(readDeadline()) })
	s.server.l.Info("WebSocket connected request_id=%s user_id=%d", requestID, userID)
	defer s.server.l.Info("WebSocket disconnected request_id=%s user_id=%d", requestID, userID)
	output := make(chan Message, 1)
	done := make(chan struct{})
	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		defer func() { _ = conn.Close() }()
		defer s.recoverConnection(conn, requestID)
		s.write(conn, requestID, output, done)
	}()
	defer func() { close(done); _ = conn.Close(); <-writerDone }()
	s.read(conn, requestID, output, writerDone)
}

func (s *Session) read(conn Connection, requestID string, output chan<- Message, writerDone <-chan struct{}) {
	defer s.recoverConnection(conn, requestID)
	for {
		kind, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if kind != TextMessage {
			closeConnection(conn, 1003, "text messages required", time.Now().Add(s.server.options.WriteTimeout))
			return
		}
		var message Message
		if !utf8.Valid(data) || json.Unmarshal(data, &message) != nil {
			closeConnection(conn, 1007, "invalid JSON", time.Now().Add(s.server.options.WriteTimeout))
			return
		}
		if message.Type != "echo" {
			closeConnection(conn, 1008, "unsupported message type", time.Now().Add(s.server.options.WriteTimeout))
			return
		}
		message.RequestID = requestID
		select {
		case output <- message:
		case <-writerDone:
			return
		}
	}
}

func (s *Session) write(conn Connection, requestID string, output <-chan Message, done <-chan struct{}) {
	if !s.writeMessage(conn, Message{Type: "notification", Data: "connected", RequestID: requestID}) {
		return
	}
	ticker := time.NewTicker(s.server.options.PingInterval)
	defer ticker.Stop()
	for {
		select {
		case message := <-output:
			if !s.writeMessage(conn, message) {
				return
			}
		case <-ticker.C:
			if err := conn.WriteControl(PingMessage, nil, time.Now().Add(s.server.options.WriteTimeout)); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}

func (s *Session) writeMessage(conn Connection, message Message) bool {
	data, err := json.Marshal(message)
	if err != nil {
		return false
	}
	if err := conn.SetWriteDeadline(time.Now().Add(s.server.options.WriteTimeout)); err != nil {
		return false
	}
	return conn.WriteMessage(TextMessage, data) == nil
}

func (s *Session) recoverConnection(conn Connection, requestID string) {
	if recovered := recover(); recovered != nil {
		s.server.l.Error("WebSocket session panic request_id=%s: %v", requestID, recovered)
		closeConnection(conn, 1011, "internal error", time.Now().Add(s.server.options.WriteTimeout))
	}
}
