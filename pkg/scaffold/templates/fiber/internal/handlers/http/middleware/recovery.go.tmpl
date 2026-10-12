package middleware

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/gofiber/fiber/v2"

	"go-boilerplate/pkg/logger"
	"go-boilerplate/pkg/response"
)

func buildPanicMessage(ctx *fiber.Ctx, err interface{}) string {
	var result strings.Builder

	result.WriteString(ctx.IP())
	result.WriteString(" - ")
	result.WriteString(ctx.Method())
	result.WriteString(" ")
	result.WriteString(ctx.OriginalURL())
	result.WriteString(" PANIC DETECTED: ")
	fmt.Fprintf(&result, "%v\n%s\n", err, debug.Stack())

	return result.String()
}

func logPanic(l logger.Interface) func(c *fiber.Ctx, err interface{}) {
	return func(ctx *fiber.Ctx, err interface{}) {
		l.Error(buildPanicMessage(ctx, err))
	}
}

// Recovery returns a middleware that recovers from panics.
func Recovery(l logger.Interface) func(c *fiber.Ctx) error {
	logRecovery := logPanic(l)
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logRecovery(c, recovered)
				err = response.InternalError(c)
			}
		}()
		return c.Next()
	}
}
