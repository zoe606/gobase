// Package redistracing provides an OpenTelemetry tracing hook for go-redis.
package redistracing

import (
	"context"
	"net"
	"strings"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "redis"

// Hook implements redis.Hook for OpenTelemetry tracing.
type Hook struct {
	tracer trace.Tracer
}

// NewHook creates a new Redis tracing hook.
func NewHook() *Hook {
	return &Hook{
		tracer: otel.Tracer(tracerName),
	}
}

// DialHook returns the dial hook (passthrough).
func (h *Hook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

// ProcessHook wraps each Redis command in a span.
func (h *Hook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		ctx, span := h.tracer.Start(ctx, "redis."+cmd.FullName(),
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(
				attribute.String("db.system", "redis"),
				attribute.String("db.operation", cmd.FullName()),
			),
		)
		defer span.End()

		err := next(ctx, cmd)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		return err
	}
}

// ProcessPipelineHook wraps pipeline operations in a span.
func (h *Hook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		names := make([]string, 0, len(cmds))
		for _, cmd := range cmds {
			names = append(names, cmd.FullName())
		}

		ctx, span := h.tracer.Start(ctx, "redis.pipeline",
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(
				attribute.String("db.system", "redis"),
				attribute.String("db.operation", "pipeline"),
				attribute.Int("db.redis.pipeline_length", len(cmds)),
				attribute.String("db.statement", strings.Join(names, ", ")),
			),
		)
		defer span.End()

		err := next(ctx, cmds)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		return err
	}
}
