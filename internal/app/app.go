// Package app configures and runs application.
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gorm.io/gorm"

	"go-boilerplate/config"
	"go-boilerplate/internal/entity"
	httphandler "go-boilerplate/internal/handlers/http"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/repo/persistent"
	"go-boilerplate/internal/repo/storage"
	"go-boilerplate/internal/repo/webapi"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/internal/usecase/article"
	"go-boilerplate/internal/usecase/auth"
	"go-boilerplate/internal/usecase/media"
	"go-boilerplate/internal/usecase/profile"
	"go-boilerplate/internal/usecase/translation"
	"go-boilerplate/pkg/asynq"
	"go-boilerplate/pkg/httpserver"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
	"go-boilerplate/pkg/postgres"
	"go-boilerplate/pkg/telemetry"
	"go-boilerplate/pkg/telemetry/gormtracing"
)

// repositories holds all repository instances.
type repositories struct {
	translation    repo.TranslationRepo
	translationAPI repo.TranslationWebAPI
	user           repo.UserRepo
	role           repo.RoleRepo
	refreshToken   repo.RefreshTokenRepo
	media          repo.MediaRepo
	profile        repo.ProfileRepo
	article        repo.ArticleRepo
}

// usecases holds all usecase instances.
type usecases struct {
	translation usecase.Translation
	auth        usecase.Auth
	media       usecase.Media
	profile     usecase.Profile
	article     usecase.Article
}

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := initLogger(cfg)

	defer func() { _ = l.Sync() }() //nolint:errcheck // best effort sync

	l.Info("Starting %s v%s (env: %s)", cfg.App.Name, cfg.App.Version, cfg.App.Env)

	for _, w := range cfg.Warnings() {
		l.Warn(w)
	}

	// Initialize telemetry (no-op when disabled)
	shutdownTelemetry, err := telemetry.Init(telemetry.Config{
		Enabled:        cfg.Telemetry.Enabled,
		ServiceName:    cfg.App.Name,
		ServiceVersion: cfg.App.Version,
		OTLPEndpoint:   cfg.Telemetry.OTLPEndpoint,
		OTLPInsecure:   cfg.Telemetry.OTLPInsecure,
		Environment:    cfg.App.Env,
	})
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - telemetry.Init: %w", err))
	}
	defer func() { _ = shutdownTelemetry(context.Background()) }()

	pg := initDatabase(cfg, l)
	defer pg.Close()

	// Register GORM tracing plugin when telemetry is enabled
	if cfg.Telemetry.Enabled {
		if err := pg.DB.Use(gormtracing.New()); err != nil {
			l.Error(fmt.Errorf("app - Run - gormtracing plugin: %w", err))
		}

		// Register DB pool metrics
		if sqlDB, dbErr := pg.DB.DB(); dbErr == nil {
			telemetry.RegisterDBMetrics(sqlDB)
		}
	}

	asynqClient := initAsynqClient(cfg)
	defer asynqClient.Close()

	storageProvider := initStorage(cfg, l)

	jwtService := initJWT(cfg)
	repos := initRepositories(pg.DB)
	uc := initUseCases(cfg, repos, jwtService, asynqClient, storageProvider, l)
	httpServer := initHTTPServer(cfg, l, uc, jwtService, pg)

	l.Info("Server started on port %s", cfg.HTTP.Port)

	waitForShutdown(httpServer, l)
}

// initLogger creates a logger based on environment.
func initLogger(cfg *config.Config) *logger.Logger {
	if cfg.App.IsProduction() {
		return logger.New(cfg.Log.Level) // JSON format for production
	}

	return logger.NewDevelopment() // Console format for development
}

// initDatabase creates database connection and runs migrations if needed.
func initDatabase(cfg *config.Config, l *logger.Logger) *postgres.Postgres {
	pg, err := postgres.New(
		cfg.Postgres.DSN(),
		postgres.MaxPoolSize(cfg.Postgres.MaxPoolSize),
		postgres.MaxIdleConns(cfg.Postgres.MaxIdleConns),
		postgres.ConnMaxLifetime(cfg.Postgres.ConnMaxLifetime),
		postgres.ConnMaxIdleTime(cfg.Postgres.ConnMaxIdleTime),
	)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}

	runAutoMigrate(cfg, pg.DB, l)

	return pg
}

// runAutoMigrate runs database migrations in development mode.
func runAutoMigrate(cfg *config.Config, db *gorm.DB, l *logger.Logger) {
	if !cfg.App.ShouldAutoMigrate() {
		l.Info("Skipping AutoMigrate (production mode) - use CLI migrations")

		return
	}

	l.Info("Running AutoMigrate (development mode)")

	if err := db.AutoMigrate(
		&entity.Translation{},
		&entity.Permission{},
		&entity.Role{},
		&entity.User{},
		&entity.RefreshToken{},
		&entity.Media{},
		&entity.Profile{},
		&entity.Article{},
	); err != nil {
		l.Fatal(fmt.Errorf("app - Run - AutoMigrate: %w", err))
	}

	l.Info("Database migration completed")

	// Seed default data in development mode
	runSeeder(db, l)
}

// initJWT creates JWT service based on configured algorithm.
func initJWT(cfg *config.Config) jwt.Service {
	switch cfg.JWT.Algorithm {
	case "rs256":
		svc, err := jwt.NewRS256(
			cfg.JWT.PrivateKeyPath,
			cfg.JWT.PublicKeyPath,
			cfg.JWT.AccessExpiry,
			cfg.JWT.RefreshExpiry,
		)
		if err != nil {
			panic(fmt.Sprintf("app - initJWT - jwt.NewRS256: %v", err))
		}
		return svc
	case "es256":
		svc, err := jwt.NewES256(
			cfg.JWT.PrivateKeyPath,
			cfg.JWT.PublicKeyPath,
			cfg.JWT.AccessExpiry,
			cfg.JWT.RefreshExpiry,
		)
		if err != nil {
			panic(fmt.Sprintf("app - initJWT - jwt.NewES256: %v", err))
		}
		return svc
	default: // "hs256" or empty
		return jwt.New(
			cfg.JWT.SecretKey,
			cfg.JWT.AccessExpiry,
			cfg.JWT.RefreshExpiry,
		)
	}
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
			l.Fatal(fmt.Errorf("app - Run - storage.NewS3Storage: %w", err))
		}
		return s3Storage
	default:
		l.Fatal(fmt.Errorf("app - Run - unknown storage driver: %s", cfg.Storage.Driver))
		return nil
	}
}

// initRepositories creates all repository instances.
func initRepositories(db *gorm.DB) *repositories {
	return &repositories{
		translation:    persistent.New(db),
		translationAPI: webapi.New(),
		user:           persistent.NewUserRepo(db),
		role:           persistent.NewRoleRepo(db),
		refreshToken:   persistent.NewRefreshTokenRepo(db),
		media:          persistent.NewMediaRepo(db),
		profile:        persistent.NewProfileRepo(db),
		article:        persistent.NewArticleRepo(db),
	}
}

// initAsynqClient creates an Asynq client for background job queuing.
func initAsynqClient(cfg *config.Config) *asynq.Client {
	return asynq.NewClient(asynq.Config{
		RedisAddr:     cfg.Redis.Addr(),
		RedisPassword: cfg.Redis.Password,
		RedisDB:       cfg.Redis.DB,
	})
}

// initUseCases creates all usecase instances.
func initUseCases(cfg *config.Config, repos *repositories, jwtService jwt.Service, asynqClient *asynq.Client, storageProvider storage.Provider, l logger.Interface) *usecases {
	authUC := auth.New(repos.user, repos.role, repos.refreshToken, jwtService).
		WithAsynq(asynqClient, cfg.App.Name)

	mediaUC := media.New(
		repos.media,
		storageProvider,
		asynqClient.Client,
		l,
		cfg.Storage.Driver,
		cfg.Storage.MaxSize,
	)

	profileUC := profile.New(
		repos.profile,
		repos.media,
		storageProvider,
		l,
	)

	articleUC := article.New(repos.article)

	return &usecases{
		translation: translation.New(repos.translation, repos.translationAPI),
		auth:        authUC,
		media:       mediaUC,
		profile:     profileUC,
		article:     articleUC,
	}
}

// initHTTPServer creates and starts HTTP server with routes.
func initHTTPServer(cfg *config.Config, l *logger.Logger, uc *usecases, jwtService jwt.Service, pg *postgres.Postgres) *httpserver.Server {
	httpServer := httpserver.New(
		l,
		httpserver.Port(cfg.HTTP.Port),
		httpserver.ReadTimeout(cfg.HTTP.Timeout),
		httpserver.WriteTimeout(cfg.HTTP.Timeout),
		httpserver.BodyLimit(cfg.HTTP.BodyLimit),
		httpserver.ShutdownTimeout(cfg.HTTP.ShutdownTimeout),
	)

	httphandler.SetupRoutes(httpServer.App, cfg, uc.translation, uc.auth, uc.media, uc.profile, uc.article, jwtService, l, pg)
	httpServer.Start()

	return httpServer
}

// waitForShutdown blocks until interrupt signal and performs graceful shutdown.
func waitForShutdown(httpServer *httpserver.Server, l *logger.Logger) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: %s", s.String())
	case err := <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	l.Info("Shutting down...")

	if err := httpServer.Shutdown(); err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}
