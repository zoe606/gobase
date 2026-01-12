package audit

import (
	"context"
)

// NoopLogger implements Logger with no-op operations.
// Useful when audit logging is disabled.
type NoopLogger struct{}

// NewNoop creates a new no-op audit logger.
func NewNoop() *NoopLogger {
	return &NoopLogger{}
}

// Log does nothing and returns nil.
func (l *NoopLogger) Log(_ context.Context, _ Entry) error {
	return nil
}

// LogCreate does nothing and returns nil.
func (l *NoopLogger) LogCreate(_ context.Context, _ string, _ uint, _ *uint, _ map[string]any) error {
	return nil
}

// LogUpdate does nothing and returns nil.
func (l *NoopLogger) LogUpdate(_ context.Context, _ string, _ uint, _ *uint, _, _ map[string]any) error {
	return nil
}

// LogDelete does nothing and returns nil.
func (l *NoopLogger) LogDelete(_ context.Context, _ string, _ uint, _ *uint, _ map[string]any) error {
	return nil
}
