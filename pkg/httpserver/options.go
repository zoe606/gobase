package httpserver

import (
	"context"
	"net"
	"time"
)

func OnShutdown(hook func(context.Context) error) Option {
	return func(s *Server) { s.shutdownHooks = append(s.shutdownHooks, hook) }
}

// Option -.
type Option func(*Server)

// Port -.
func Port(port string) Option {
	return func(s *Server) {
		s.address = net.JoinHostPort("", port)
	}
}

// Prefork -.
func Prefork(prefork bool) Option {
	return func(s *Server) {
		s.prefork = prefork
	}
}

// ReadTimeout -.
func ReadTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.readTimeout = timeout
	}
}

// WriteTimeout -.
func WriteTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.writeTimeout = timeout
	}
}

// ShutdownTimeout -.
func ShutdownTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = timeout
	}
}

// BodyLimit sets the maximum allowed request body size in bytes.
func BodyLimit(limit int) Option {
	return func(s *Server) {
		s.bodyLimit = limit
	}
}
