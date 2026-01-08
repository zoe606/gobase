// Package app configures and runs application.
package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go-boilerplate/config"
	"go-boilerplate/internal/controller/restapi"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo/persistent"
	"go-boilerplate/internal/repo/webapi"
	"go-boilerplate/internal/usecase/auth"
	"go-boilerplate/internal/usecase/translation"
	"go-boilerplate/pkg/httpserver"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
	"go-boilerplate/pkg/postgres"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	// ============================================
	// Logger
	// ============================================
	var l *logger.Logger

	if cfg.App.IsProduction() {
		l = logger.New(cfg.Log.Level) // JSON format for production
	} else {
		l = logger.NewDevelopment() // Console format for development
	}

	defer func() { _ = l.Sync() }()

	l.Info("Starting %s v%s (env: %s)", cfg.App.Name, cfg.App.Version, cfg.App.Env)

	// ============================================
	// Database
	// ============================================
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
	defer pg.Close()

	// Auto-migrate database schema (development only)
	if cfg.App.ShouldAutoMigrate() {
		l.Info("Running AutoMigrate (development mode)")

		if err = pg.DB.AutoMigrate(
			&entity.Translation{},
			&entity.Permission{},
			&entity.Role{},
			&entity.User{},
			&entity.RefreshToken{},
		); err != nil {
			l.Fatal(fmt.Errorf("app - Run - AutoMigrate: %w", err))
		}

		l.Info("Database migration completed")
	} else {
		l.Info("Skipping AutoMigrate (production mode) - use CLI migrations")
	}

	// ============================================
	// JWT Service
	// ============================================
	jwtService := jwt.New(
		cfg.JWT.SecretKey,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
	)

	// ============================================
	// Repositories
	// ============================================
	translationRepo := persistent.New(pg)
	translationWebAPI := webapi.New()
	userRepo := persistent.NewUserRepo(pg)
	roleRepo := persistent.NewRoleRepo(pg)
	refreshTokenRepo := persistent.NewRefreshTokenRepo(pg)

	// ============================================
	// Use Cases
	// ============================================
	translationUseCase := translation.New(translationRepo, translationWebAPI)
	authUseCase := auth.New(userRepo, roleRepo, refreshTokenRepo, jwtService)

	// ============================================
	// HTTP Server
	// ============================================
	httpServer := httpserver.New(
		l,
		httpserver.Port(cfg.HTTP.Port),
		httpserver.ReadTimeout(cfg.HTTP.Timeout),
		httpserver.WriteTimeout(cfg.HTTP.Timeout),
	)

	// Setup routes with dependency injection
	restapi.NewRouter(httpServer.App, cfg, translationUseCase, authUseCase, jwtService, l)

	// Start server
	httpServer.Start()

	l.Info("Server started on port %s", cfg.HTTP.Port)

	// ============================================
	// Graceful Shutdown
	// ============================================
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: %s", s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	l.Info("Shutting down...")
	if err = httpServer.Shutdown(); err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}
