// Package httpserver implements HTTP server.
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"go-boilerplate/pkg/logger"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/sync/errgroup"
)

const (
	_defaultAddr            = ":80"
	_defaultReadTimeout     = 5 * time.Second
	_defaultWriteTimeout    = 5 * time.Second
	_defaultShutdownTimeout = 3 * time.Second
	_maxPortAttempts        = 10 // Max ports to try when auto-finding
)

// Server -.
type Server struct {
	ctx context.Context
	eg  *errgroup.Group

	App    *fiber.App
	notify chan error

	address         string
	actualPort      string // The port actually bound to (may differ if autoPort is enabled)
	prefork         bool
	autoPort        bool // Enable automatic port finding if configured port is busy
	readTimeout     time.Duration
	writeTimeout    time.Duration
	shutdownTimeout time.Duration

	logger logger.Interface
}

// New -.
func New(l logger.Interface, opts ...Option) *Server {
	group, ctx := errgroup.WithContext(context.Background())
	group.SetLimit(1) // Run only one goroutine

	s := &Server{
		ctx:             ctx,
		eg:              group,
		App:             nil,
		notify:          make(chan error, 1),
		address:         _defaultAddr,
		readTimeout:     _defaultReadTimeout,
		writeTimeout:    _defaultWriteTimeout,
		shutdownTimeout: _defaultShutdownTimeout,
		logger:          l,
	}

	// Custom options
	for _, opt := range opts {
		opt(s)
	}

	app := fiber.New(fiber.Config{
		Prefork:      s.prefork,
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		JSONDecoder:  json.Unmarshal,
		JSONEncoder:  json.Marshal,
	})

	s.App = app

	return s
}

// Start -.
func (s *Server) Start() {
	s.eg.Go(func() error {
		address := s.address
		var err error

		if s.autoPort {
			address, err = s.findAvailablePort()
			if err != nil {
				s.notify <- err
				close(s.notify)

				return err
			}
		}

		s.actualPort = extractPort(address)

		s.logger.Info("HTTP server listening on %s", address)

		err = s.App.Listen(address)
		if err != nil {
			s.notify <- err

			close(s.notify)

			return err
		}

		return nil
	})
}

// findAvailablePort tries the configured port and increments if busy.
func (s *Server) findAvailablePort() (string, error) {
	basePort := extractPort(s.address)
	port, err := strconv.Atoi(basePort)
	if err != nil {
		return s.address, nil // If port is not numeric, just use as-is
	}

	lc := net.ListenConfig{}
	ctx := context.Background()

	for i := 0; i < _maxPortAttempts; i++ {
		testPort := port + i
		testAddr := net.JoinHostPort("", strconv.Itoa(testPort))

		// Try to listen briefly to check if port is available
		// Use tcp4 to match Fiber's default behavior on Windows
		listener, listenErr := lc.Listen(ctx, "tcp4", testAddr)
		if listenErr != nil {
			if i == 0 {
				s.logger.Info("Port %d is busy, trying next port...", testPort)
			}

			continue
		}

		listener.Close()

		if i > 0 {
			s.logger.Info("Found available port: %d (configured: %d)", testPort, port)
		}

		return testAddr, nil
	}

	return "", fmt.Errorf("could not find available port after %d attempts (tried %d-%d)", _maxPortAttempts, port, port+_maxPortAttempts-1)
}

// extractPort extracts the port from an address string.
func extractPort(address string) string {
	// Handle formats: ":8080", "localhost:8080", "0.0.0.0:8080"
	if idx := strings.LastIndex(address, ":"); idx != -1 {
		return address[idx+1:]
	}

	return address
}

// Port returns the actual port the server is listening on.
// This may differ from the configured port if autoPort is enabled.
func (s *Server) Port() string {
	if s.actualPort != "" {
		return s.actualPort
	}

	return extractPort(s.address)
}

// Notify -.
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Shutdown -.
func (s *Server) Shutdown() error {
	var shutdownErrors []error

	err := s.App.ShutdownWithTimeout(s.shutdownTimeout)
	if err != nil && !errors.Is(err, context.Canceled) {
		s.logger.Error(err, "restapi server - Server - Shutdown - s.App.ShutdownWithTimeout")

		shutdownErrors = append(shutdownErrors, err)
	}

	// Wait for all goroutines to finish and get any error
	err = s.eg.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		s.logger.Error(err, "restapi server - Server - Shutdown - s.eg.Wait")

		shutdownErrors = append(shutdownErrors, err)
	}

	s.logger.Info("restapi server - Server - Shutdown")

	return errors.Join(shutdownErrors...)
}
