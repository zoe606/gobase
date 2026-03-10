// Package asynqtracing provides OpenTelemetry tracing middleware for Asynq workers.
package asynqtracing

import (
	"context"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "asynq"

// Middleware returns an Asynq middleware that creates a span for each task.
func Middleware() asynq.MiddlewareFunc {
	tracer := otel.Tracer(tracerName)

	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
			queue, _ := asynq.GetQueueName(ctx)

			ctx, span := tracer.Start(ctx, "asynq.process "+task.Type(),
				trace.WithSpanKind(trace.SpanKindConsumer),
				trace.WithAttributes(
					attribute.String("messaging.system", "asynq"),
					attribute.String("messaging.operation", "process"),
					attribute.String("messaging.destination", queue),
					attribute.String("messaging.message.type", task.Type()),
				),
			)
			defer span.End()

			err := next.ProcessTask(ctx, task)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "")
			}

			return err
		})
	}
}
