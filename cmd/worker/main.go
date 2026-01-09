package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go-boilerplate/config"
	"go-boilerplate/internal/repo/persistent"
	"go-boilerplate/internal/repo/storage"
	"go-boilerplate/internal/repo/webapi/email"
	"go-boilerplate/internal/worker"
	"go-boilerplate/pkg/asynq"
	"go-boilerplate/pkg/logger"
	"go-boilerplate/pkg/postgres"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	l := initLogger(cfg)
	defer func() { _ = l.Sync() }() //nolint:errcheck // best effort sync

	l.Info("Starting worker (env: %s)", cfg.App.Env)

	// Initialize database
	pg, err := postgres.New(
		cfg.Postgres.DSN(),
		postgres.MaxPoolSize(cfg.Postgres.MaxPoolSize),
		postgres.MaxIdleConns(cfg.Postgres.MaxIdleConns),
		postgres.ConnMaxLifetime(cfg.Postgres.ConnMaxLifetime),
		postgres.ConnMaxIdleTime(cfg.Postgres.ConnMaxIdleTime),
	)
	if err != nil {
		l.Fatal(fmt.Errorf("postgres connection error: %w", err))
	}
	defer pg.Close()

	// Initialize email sender
	mailer := initEmailSender(cfg, l)

	// Initialize storage provider
	storageProvider := initStorage(cfg, l)

	// Initialize media repo
	mediaRepo := persistent.NewMediaRepo(pg.DB)

	// Initialize Asynq server
	srv := asynq.NewServer(asynq.ServerConfig{
		RedisAddr:     cfg.Redis.Addr(),
		RedisPassword: cfg.Redis.Password,
		RedisDB:       cfg.Redis.DB,
		Concurrency:   cfg.Asynq.Concurrency,
	}, l)

	// Create worker with dependencies
	w := worker.New(srv, l, mailer, mediaRepo, storageProvider)

	// Start worker in goroutine
	go func() {
		l.Info("Starting Asynq worker...")

		if err := w.Start(); err != nil {
			l.Fatal(fmt.Errorf("worker failed: %w", err))
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	l.Info("Shutting down worker...")
	w.Shutdown()
}

// initLogger creates a logger based on environment.
func initLogger(cfg *config.Config) *logger.Logger {
	if cfg.App.IsProduction() {
		return logger.New(cfg.Log.Level)
	}

	return logger.NewDevelopment()
}

// initEmailSender creates an email sender based on config.
func initEmailSender(cfg *config.Config, l *logger.Logger) email.Sender {
	if cfg.Email.APIKey == "" {
		l.Info("Email API key not configured, using noop mailer")

		return email.NewNoopSender(l)
	}

	l.Info("Using Resend email provider")

	return email.NewResendSender(
		cfg.Email.APIKey,
		cfg.Email.FromEmail,
	)
}

// initStorage creates storage provider based on configuration.
func initStorage(cfg *config.Config, l *logger.Logger) storage.Provider {
	switch cfg.Storage.Driver {
	case "local":
		return storage.NewLocalStorage(cfg.Storage.LocalPath, cfg.Storage.LocalURL)
	case "s3":
		s3Storage, err := storage.NewS3Storage(
			cfg.Storage.S3Endpoint,
			cfg.Storage.S3AccessKey,
			cfg.Storage.S3SecretKey,
			cfg.Storage.S3Bucket,
			cfg.Storage.S3UseSSL,
		)
		if err != nil {
			l.Fatal(fmt.Errorf("storage.NewS3Storage: %w", err))
		}
		return s3Storage
	default:
		l.Fatal(fmt.Errorf("unknown storage driver: %s", cfg.Storage.Driver))
		return nil
	}
}
