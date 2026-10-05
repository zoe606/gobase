package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "http"

// headerCarrier adapts Fiber request headers for OTel propagation.
type headerCarrier struct {
	ctx *fiber.Ctx
}

func (c *headerCarrier) Get(key string) string {
	return c.ctx.Get(key)
}

func (c *headerCarrier) Set(key, value string) {
	c.ctx.Set(key, value)
}

func (c *headerCarrier) Keys() []string {
	keys := make([]string, 0)
	for k := range c.ctx.Request().Header.All() {
		keys = append(keys, string(k))
	}
	return keys
}

// Tracing creates an OpenTelemetry tracing middleware for Fiber.
func Tracing() fiber.Handler {
	tracer := otel.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()

	return func(c *fiber.Ctx) error {
		carrier := &headerCarrier{ctx: c}
		ctx := propagator.Extract(c.UserContext(), carrier)

		route := c.Route().Path
		if route == "" {
			route = c.Path()
		}

		spanName := fmt.Sprintf("HTTP %s %s", c.Method(), route)
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", c.Method()),
				attribute.String("http.route", route),
				attribute.String("http.target", c.OriginalURL()),
			),
		)
		defer span.End()

		c.SetUserContext(ctx)

		err := c.Next()

		statusCode := c.Response().StatusCode()
		span.SetAttributes(attribute.Int("http.status_code", statusCode))

		if statusCode >= 500 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", statusCode))
		} else {
			span.SetStatus(codes.Ok, "")
		}

		if err != nil {
			span.RecordError(err)
		}

		return err
	}
}
