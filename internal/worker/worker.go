// Package worker provides background job processing using Asynq.
package worker

import (
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/repo/storage"
	"go-boilerplate/internal/repo/webapi/email"
	"go-boilerplate/internal/worker/tasks"
	"go-boilerplate/pkg/asynq"
	"go-boilerplate/pkg/logger"
)

// Worker manages background task processing.
type Worker struct {
	server    *asynq.Server
	l         logger.Interface
	mailer    email.Sender
	mediaRepo repo.MediaRepo
	storage   storage.Provider
}

// New creates a new Worker instance.
func New(server *asynq.Server, l logger.Interface, mailer email.Sender, mediaRepo repo.MediaRepo, storageProvider storage.Provider) *Worker {
	return &Worker{
		server:    server,
		l:         l,
		mailer:    mailer,
		mediaRepo: mediaRepo,
		storage:   storageProvider,
	}
}

// RegisterHandlers registers all task handlers.
func (w *Worker) RegisterHandlers() {
	// Email notification handler.
	emailHandler := tasks.NewEmailHandler(w.l, w.mailer)
	w.server.HandleFunc(tasks.TypeEmailNotification, emailHandler.ProcessTask)

	// Image processing handler.
	if w.mediaRepo != nil && w.storage != nil {
		imageHandler := tasks.NewImageProcessingHandler(w.l, w.mediaRepo, w.storage)
		w.server.HandleFunc(tasks.TypeImageProcessing, imageHandler.ProcessTask)
		w.l.Info("Registered image processing handler")
	}

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
