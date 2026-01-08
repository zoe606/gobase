// Package httphandler implements HTTP routing and handlers.
package httphandler

import (
	"net/http"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"

	"go-boilerplate/config"
	_ "go-boilerplate/docs" // Swagger docs.
	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/internal/handlers/http/v1/auth"
	"go-boilerplate/internal/handlers/http/v1/translation"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

// HealthChecker provides health check capabilities for readiness probes.
type HealthChecker interface {
	Ping() error
}

// SetupRoutes sets up all routes and middleware.
// Swagger spec:
//
//	@title       Go Clean Template API
//	@description Using a translation service as an example
//	@version     1.0
//	@host        localhost:8080
//	@BasePath    /v1
func SetupRoutes(app *fiber.App, cfg *config.Config, translationUC usecase.Translation, authUC usecase.Auth, jwtService jwt.Service, l logger.Interface, healthChecker HealthChecker) {
	setupMiddleware(app, cfg, l)
	setupOptionalFeatures(app, cfg)
	setupHealthEndpoints(app, healthChecker)
	setupAPIRoutes(app, translationUC, authUC, jwtService, l)
}

// setupMiddleware configures global middleware chain.
func setupMiddleware(app *fiber.App, cfg *config.Config, l logger.Interface) {
	app.Use(recover.New(recover.Config{EnableStackTrace: true}))
	app.Use(requestid.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}))
	app.Use(helmet.New())
	app.Use(compress.New(compress.Config{Level: compress.LevelDefault}))
	app.Use(limiter.New(limiter.Config{
		Max:          cfg.RateLimit.Max,
		Expiration:   cfg.RateLimit.Expiration,
		KeyGenerator: func(c *fiber.Ctx) string { return c.IP() },
		LimitReached: rateLimitReached,
	}))
	app.Use(middleware.Logger(l))
}

// rateLimitReached handles rate limit exceeded responses.
func rateLimitReached(c *fiber.Ctx) error {
	return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    "RATE_LIMITED",
			"message": "Too many requests. Please try again later.",
		},
	})
}

// setupOptionalFeatures configures metrics and swagger based on config.
func setupOptionalFeatures(app *fiber.App, cfg *config.Config) {
	if cfg.Metrics.Enabled {
		prometheus := fiberprometheus.New(cfg.App.Name)
		prometheus.RegisterAt(app, "/metrics")
		app.Use(prometheus.Middleware)
	}

	if cfg.Swagger.Enabled {
		app.Get("/swagger/*", swagger.HandlerDefault)
	}
}

// setupHealthEndpoints configures K8s health check endpoints.
func setupHealthEndpoints(app *fiber.App, checker HealthChecker) {
	// Liveness probe - always returns OK if the server is running
	app.Get("/healthz", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(http.StatusOK)
	})

	// Readiness probe - checks if dependencies (DB) are available
	app.Get("/readyz", func(ctx *fiber.Ctx) error {
		if err := checker.Ping(); err != nil {
			return ctx.Status(http.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "unhealthy",
				"error":  err.Error(),
			})
		}
		return ctx.SendStatus(http.StatusOK)
	})
}

// setupAPIRoutes configures API v1 routes.
func setupAPIRoutes(app *fiber.App, translationUC usecase.Translation, authUC usecase.Auth, jwtService jwt.Service, l logger.Interface) {
	apiV1Group := app.Group("/v1")

	translationHandler := translation.New(translationUC, l)
	translationHandler.RegisterRoutes(apiV1Group)

	authHandler := auth.New(authUC, jwtService, l)
	authHandler.RegisterRoutes(apiV1Group)
}
