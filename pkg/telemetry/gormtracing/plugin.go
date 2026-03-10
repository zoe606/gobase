// Package gormtracing provides an OpenTelemetry tracing plugin for GORM.
package gormtracing

import (
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

const (
	tracerName      = "gorm"
	maxStatementLen = 256
	callbackPrefix  = "otel:"
	spanPrefix      = "gorm."
)

// Plugin implements gorm.Plugin for OpenTelemetry tracing.
type Plugin struct{}

// New creates a new GORM tracing plugin.
func New() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "otel-tracing"
}

// Initialize registers tracing callbacks on all GORM operations.
func (p *Plugin) Initialize(db *gorm.DB) error {
	tracer := otel.Tracer(tracerName)

	cb := func(operation string) func(*gorm.DB) {
		return func(db *gorm.DB) {
			if db.Statement == nil || db.Statement.Context == nil {
				return
			}

			_, span := tracer.Start(db.Statement.Context, spanPrefix+operation,
				trace.WithSpanKind(trace.SpanKindClient),
			)

			attrs := []attribute.KeyValue{
				attribute.String("db.system", "postgresql"),
			}

			if db.Statement.Table != "" {
				attrs = append(attrs, attribute.String("db.sql.table", db.Statement.Table))
			}

			sql := db.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
			if len(sql) > maxStatementLen {
				sql = sql[:maxStatementLen]
			}
			if sql != "" {
				attrs = append(attrs, attribute.String("db.statement", sql))
			}

			span.SetAttributes(attrs...)

			if db.Error != nil {
				span.RecordError(db.Error)
				span.SetStatus(codes.Error, db.Error.Error())
			}

			span.End()
		}
	}

	type callbackReg struct {
		name     string
		register func(string, func(*gorm.DB)) error
	}

	registrations := []callbackReg{
		{"create", func(n string, f func(*gorm.DB)) error {
			return db.Callback().Create().After("gorm:create").Register(n, f)
		}},
		{"query", func(n string, f func(*gorm.DB)) error {
			return db.Callback().Query().After("gorm:query").Register(n, f)
		}},
		{"update", func(n string, f func(*gorm.DB)) error {
			return db.Callback().Update().After("gorm:update").Register(n, f)
		}},
		{"delete", func(n string, f func(*gorm.DB)) error {
			return db.Callback().Delete().After("gorm:delete").Register(n, f)
		}},
		{"row", func(n string, f func(*gorm.DB)) error { return db.Callback().Row().After("gorm:row").Register(n, f) }},
		{"raw", func(n string, f func(*gorm.DB)) error { return db.Callback().Raw().After("gorm:raw").Register(n, f) }},
	}

	for _, reg := range registrations {
		name := fmt.Sprintf("%s%s", callbackPrefix, reg.name)
		if err := reg.register(name, cb(reg.name)); err != nil {
			return fmt.Errorf("register %s callback: %w", reg.name, err)
		}
	}

	return nil
}
