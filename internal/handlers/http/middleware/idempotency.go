package middleware

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"go-boilerplate/config"
	"go-boilerplate/pkg/cache"
)

const idempotencyHeader = "Idempotency-Key"

// idempotencyResponse stores a cached response for replay.
type idempotencyResponse struct {
	Status      int    `json:"status"`
	ContentType string `json:"content_type"`
	Body        []byte `json:"body"`
}

// Idempotency returns middleware that deduplicates mutating requests using an Idempotency-Key header.
// GET, DELETE, OPTIONS, and HEAD requests bypass the middleware entirely.
func Idempotency(appCache cache.Cache, cfg config.Idempotency) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		method := ctx.Method()
		if method == fiber.MethodGet || method == fiber.MethodDelete ||
			method == fiber.MethodOptions || method == fiber.MethodHead {
			return ctx.Next()
		}

		key := strings.TrimSpace(ctx.Get(idempotencyHeader))

		if key == "" {
			if cfg.RequiredForPost && method == fiber.MethodPost {
				return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error": fiber.Map{
						"code":    "MISSING_IDEMPOTENCY_KEY",
						"message": "Idempotency-Key header is required for POST requests",
					},
				})
			}

			return ctx.Next()
		}

		// Best-effort user scoping (only works if JWTAuth ran before this middleware)
		userPart := "shared"
		if id, ok := ctx.Locals(UserIDKey).(uint); ok {
			userPart = strconv.FormatUint(uint64(id), 10)
		}
		cacheKey := "idempotency:" + userPart + ":" + ctx.Method() + ":" + ctx.Path() + ":" + key

		// Check for cached response.
		var cached idempotencyResponse

		err := appCache.Get(ctx.Context(), cacheKey, &cached)
		if err == nil {
			ctx.Set("Content-Type", cached.ContentType)
			ctx.Set("X-Idempotent-Replay", "true")

			return ctx.Status(cached.Status).Send(cached.Body)
		}

		if !errors.Is(err, cache.ErrNotFound) {
			// Real cache error — proceed without idempotency.
			return ctx.Next()
		}

		// Cache miss — execute the handler.
		if err := ctx.Next(); err != nil {
			return err
		}

		// Cache the response (best effort).
		resp := idempotencyResponse{
			Status:      ctx.Response().StatusCode(),
			ContentType: string(ctx.Response().Header.ContentType()),
			Body:        ctx.Response().Body(),
		}
		_ = appCache.Set(ctx.Context(), cacheKey, resp, cfg.TTL)

		return nil
	}
}
