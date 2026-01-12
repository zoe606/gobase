package asynq

import (
	"context"

	"go-boilerplate/pkg/logger"

	"github.com/hibiken/asynq"
)

// Server wraps the Asynq server for processing tasks.
type Server struct {
	*asynq.Server
	mux *asynq.ServeMux
	l   logger.Interface
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	Concurrency   int
}

// NewServer creates a new Asynq server.
func NewServer(cfg ServerConfig, l logger.Interface) *Server {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		},
		asynq.Config{
			Concurrency: cfg.Concurrency,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				l.Error(err, "asynq task failed",
					"type", task.Type(),
				)
			}),
		},
	)

	return &Server{
		Server: srv,
		mux:    asynq.NewServeMux(),
		l:      l,
	}
}

// HandleFunc registers a handler function for a task type.
func (s *Server) HandleFunc(pattern string, handler func(context.Context, *asynq.Task) error) {
	s.mux.HandleFunc(pattern, handler)
}

// Handle registers a handler for a task type.
func (s *Server) Handle(pattern string, handler asynq.Handler) {
	s.mux.Handle(pattern, handler)
}

// Start starts the Asynq server.
func (s *Server) Start() error {
	s.l.Info("Starting Asynq worker server...")
	return s.Run(s.mux)
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() {
	s.l.Info("Shutting down Asynq worker server...")
	s.Server.Shutdown()
}
