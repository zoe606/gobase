// Package restapi implements routing paths. Each services in own file.
package restapi

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
	"go-boilerplate/internal/controller/restapi/middleware"
	v1 "go-boilerplate/internal/controller/restapi/v1"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/internal/usecase/auth"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

// NewRouter sets up all routes and middleware.
// Swagger spec:
// @title       Go Clean Template API
// @description Using a translation service as an example
// @version     1.0
// @host        localhost:8080
// @BasePath    /v1
func NewRouter(app *fiber.App, cfg *config.Config, t usecase.Translation, authUC *auth.UseCase, jwtService jwt.Service, l logger.Interface) {
	// ============================================
	// Global Middleware Chain (order matters!)
	// ============================================

	// 1. Panic Recovery (must be first to catch all panics)
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	// 2. Request ID (for tracing/correlation)
	app.Use(requestid.New())

	// Note: Request timeout is handled per-handler using context.WithTimeout
	// The deprecated timeout.New middleware was removed due to data race issues.
	// Use ctx.UserContext() with context.WithTimeout in handlers that need it.

	// 3. CORS (Cross-Origin Resource Sharing)
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}))

	// 5. Security Headers (Helmet)
	app.Use(helmet.New())

	// 6. Response Compression
	app.Use(compress.New(compress.Config{
		Level: compress.LevelDefault,
	}))

	// 7. Rate Limiting (per IP)
	app.Use(limiter.New(limiter.Config{
		Max:        cfg.RateLimit.Max,
		Expiration: cfg.RateLimit.Expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "RATE_LIMITED",
					"message": "Too many requests. Please try again later.",
				},
			})
		},
	}))

	// 8. Custom Logger (with Zap - keeps existing format)
	app.Use(middleware.Logger(l))

	// ============================================
	// Prometheus Metrics
	// ============================================
	if cfg.Metrics.Enabled {
		prometheus := fiberprometheus.New(cfg.App.Name)
		prometheus.RegisterAt(app, "/metrics")
		app.Use(prometheus.Middleware)
	}

	// ============================================
	// Swagger Documentation
	// ============================================
	if cfg.Swagger.Enabled {
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// ============================================
	// Health Check Endpoints (K8s probes)
	// ============================================

	// Liveness probe - is the app alive?
	app.Get("/healthz", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(http.StatusOK)
	})

	// Readiness probe - is the app ready to receive traffic?
	// TODO: Add database and Redis health checks here
	app.Get("/readyz", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(http.StatusOK)
	})

	// ============================================
	// API Routes
	// ============================================
	apiV1Group := app.Group("/v1")
	{
		v1.NewTranslationRoutes(apiV1Group, t, l)
		v1.NewAuthRoutes(apiV1Group, authUC, jwtService, l)
	}
}
