package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Timeout adds a context deadline to each request, propagating through all layers via ctx.
func Timeout(timeout time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), timeout)
		defer cancel()
		c.SetUserContext(ctx)
		return c.Next()
	}
}
