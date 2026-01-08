// Package worker provides background job processing using Asynq.
package worker

import (
	"go-boilerplate/internal/worker/tasks"
	"go-boilerplate/pkg/asynq"
	"go-boilerplate/pkg/logger"
)

// Worker manages background task processing.
type Worker struct {
	server *asynq.Server
	l      logger.Interface
}

// New creates a new Worker instance.
func New(server *asynq.Server, l logger.Interface) *Worker {
	return &Worker{
		server: server,
		l:      l,
	}
}

// RegisterHandlers registers all task handlers.
func (w *Worker) RegisterHandlers() {
	// Email notification handler
	emailHandler := tasks.NewEmailHandler(w.l)
	w.server.HandleFunc(tasks.TypeEmailNotification, emailHandler.ProcessTask)

	// Add more handlers here as needed
	// w.server.HandleFunc(tasks.TypeUserCleanup, cleanupHandler.ProcessTask)

	w.l.Info("Registered all task handlers")
}

// Start starts the worker server.
func (w *Worker) Start() error {
	w.RegisterHandlers()
	return w.server.Start()
}

// Shutdown gracefully shuts down the worker.
func (w *Worker) Shutdown() {
	w.server.Shutdown()
}
