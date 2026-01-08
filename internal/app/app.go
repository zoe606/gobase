// Package app configures and runs application.
package app

import (
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
	"go-boilerplate/internal/repo/webapi"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/internal/usecase/auth"
	"go-boilerplate/internal/usecase/translation"
	"go-boilerplate/pkg/httpserver"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
	"go-boilerplate/pkg/postgres"
)

// repositories holds all repository instances.
type repositories struct {
	translation    repo.TranslationRepo
	translationAPI repo.TranslationWebAPI
	user           repo.UserRepo
	role           repo.RoleRepo
	refreshToken   repo.RefreshTokenRepo
}

// usecases holds all usecase instances.
type usecases struct {
	translation usecase.Translation
	auth        usecase.Auth
}

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := initLogger(cfg)
	defer func() { _ = l.Sync() }()

	l.Info("Starting %s v%s (env: %s)", cfg.App.Name, cfg.App.Version, cfg.App.Env)

	pg := initDatabase(cfg, l)
	defer pg.Close()

	jwtService := initJWT(cfg)
	repos := initRepositories(pg.DB)
	uc := initUseCases(repos, jwtService)
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
	); err != nil {
		l.Fatal(fmt.Errorf("app - Run - AutoMigrate: %w", err))
	}

	l.Info("Database migration completed")
}

// initJWT creates JWT service.
func initJWT(cfg *config.Config) jwt.Service {
	return jwt.New(
		cfg.JWT.SecretKey,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
	)
}

// initRepositories creates all repository instances.
func initRepositories(db *gorm.DB) *repositories {
	return &repositories{
		translation:    persistent.New(db),
		translationAPI: webapi.New(),
		user:           persistent.NewUserRepo(db),
		role:           persistent.NewRoleRepo(db),
		refreshToken:   persistent.NewRefreshTokenRepo(db),
	}
}

// initUseCases creates all usecase instances.
func initUseCases(repos *repositories, jwtService jwt.Service) *usecases {
	return &usecases{
		translation: translation.New(repos.translation, repos.translationAPI),
		auth:        auth.New(repos.user, repos.role, repos.refreshToken, jwtService),
	}
}

// initHTTPServer creates and starts HTTP server with routes.
func initHTTPServer(cfg *config.Config, l *logger.Logger, uc *usecases, jwtService jwt.Service, pg *postgres.Postgres) *httpserver.Server {
	httpServer := httpserver.New(
		l,
		httpserver.Port(cfg.HTTP.Port),
		httpserver.ReadTimeout(cfg.HTTP.Timeout),
		httpserver.WriteTimeout(cfg.HTTP.Timeout),
	)

	httphandler.SetupRoutes(httpServer.App, cfg, uc.translation, uc.auth, jwtService, l, pg)
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
