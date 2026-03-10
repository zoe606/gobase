package telemetry

import (
	"context"
	"database/sql"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const meterName = "go-boilerplate"

// CacheMetrics tracks cache hit/miss counters.
type CacheMetrics struct {
	hits   atomic.Int64
	misses atomic.Int64
}

// NewCacheMetrics creates a new CacheMetrics and registers OTel instruments.
func NewCacheMetrics() *CacheMetrics {
	m := &CacheMetrics{}

	meter := otel.Meter(meterName)

	_, _ = meter.Int64ObservableCounter("cache_hits_total",
		metric.WithDescription("Total cache hits"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(m.hits.Load())
			return nil
		}),
	)

	_, _ = meter.Int64ObservableCounter("cache_misses_total",
		metric.WithDescription("Total cache misses"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(m.misses.Load())
			return nil
		}),
	)

	return m
}

// RecordHit increments the cache hit counter.
func (m *CacheMetrics) RecordHit() {
	m.hits.Add(1)
}

// RecordMiss increments the cache miss counter.
func (m *CacheMetrics) RecordMiss() {
	m.misses.Add(1)
}

// RegisterDBMetrics registers OTel instruments that read from sql.DB.Stats().
func RegisterDBMetrics(sqlDB *sql.DB) {
	meter := otel.Meter(meterName)

	_, _ = meter.Int64ObservableGauge("db_pool_open_connections",
		metric.WithDescription("Number of open database connections"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			stats := sqlDB.Stats()
			o.Observe(int64(stats.OpenConnections))
			return nil
		}),
	)

	_, _ = meter.Int64ObservableGauge("db_pool_in_use",
		metric.WithDescription("Number of in-use database connections"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			stats := sqlDB.Stats()
			o.Observe(int64(stats.InUse))
			return nil
		}),
	)

	_, _ = meter.Int64ObservableGauge("db_pool_idle",
		metric.WithDescription("Number of idle database connections"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			stats := sqlDB.Stats()
			o.Observe(int64(stats.Idle))
			return nil
		}),
	)
}

// AsynqMetrics tracks asynq task counters.
type AsynqMetrics struct {
	processed atomic.Int64
	failed    atomic.Int64
}

// NewAsynqMetrics creates a new AsynqMetrics and registers OTel instruments.
func NewAsynqMetrics() *AsynqMetrics {
	m := &AsynqMetrics{}

	meter := otel.Meter(meterName)

	_, _ = meter.Int64ObservableCounter("asynq_tasks_processed_total",
		metric.WithDescription("Total asynq tasks processed"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(m.processed.Load())
			return nil
		}),
	)

	_, _ = meter.Int64ObservableCounter("asynq_tasks_failed_total",
		metric.WithDescription("Total asynq tasks failed"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(m.failed.Load())
			return nil
		}),
	)

	return m
}

// RecordProcessed increments the processed counter.
func (m *AsynqMetrics) RecordProcessed() {
	m.processed.Add(1)
}

// RecordFailed increments the failed counter.
func (m *AsynqMetrics) RecordFailed() {
	m.failed.Add(1)
}
