package audit

import (
	"context"

	"gorm.io/gorm"

	"go-boilerplate/internal/entity"
	"go-boilerplate/pkg/tx"
)

// PostgresLogger implements Logger using PostgreSQL via GORM.
type PostgresLogger struct {
	db *gorm.DB
}

// NewPostgres creates a new PostgreSQL audit logger.
func NewPostgres(db *gorm.DB) *PostgresLogger {
	return &PostgresLogger{db: db}
}

// Log records an audit entry to the database.
func (l *PostgresLogger) Log(ctx context.Context, entry Entry) error {
	auditLog := &entity.AuditLog{
		EntityType: entry.EntityType,
		EntityID:   entry.EntityID,
		Action:     string(entry.Action),
		UserID:     entry.UserID,
		OldValues:  entry.OldValues,
		NewValues:  entry.NewValues,
		Metadata:   entry.Metadata,
	}

	db := tx.DBFromContext(ctx, l.db)
	return db.Create(auditLog).Error
}

// LogCreate records a create action.
func (l *PostgresLogger) LogCreate(ctx context.Context, entityType string, entityID uint, userID *uint, values map[string]any) error {
	return l.Log(ctx, Entry{
		EntityType: entityType,
		EntityID:   entityID,
		Action:     ActionCreate,
		UserID:     userID,
		NewValues:  values,
	})
}

// LogUpdate records an update action.
func (l *PostgresLogger) LogUpdate(ctx context.Context, entityType string, entityID uint, userID *uint, oldValues, newValues map[string]any) error {
	return l.Log(ctx, Entry{
		EntityType: entityType,
		EntityID:   entityID,
		Action:     ActionUpdate,
		UserID:     userID,
		OldValues:  oldValues,
		NewValues:  newValues,
	})
}

// LogDelete records a delete action.
func (l *PostgresLogger) LogDelete(ctx context.Context, entityType string, entityID uint, userID *uint, values map[string]any) error {
	return l.Log(ctx, Entry{
		EntityType: entityType,
		EntityID:   entityID,
		Action:     ActionDelete,
		UserID:     userID,
		OldValues:  values,
	})
}
