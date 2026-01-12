// Package audit provides audit logging functionality for tracking entity changes.
package audit

import (
	"context"
)

// Action represents the type of audit action.
type Action string

const (
	// ActionCreate represents a create operation.
	ActionCreate Action = "create"
	// ActionUpdate represents an update operation.
	ActionUpdate Action = "update"
	// ActionDelete represents a delete operation.
	ActionDelete Action = "delete"
)

// Entry represents an audit log entry.
type Entry struct {
	// EntityType is the type of entity being audited (e.g., "user", "translation").
	EntityType string

	// EntityID is the ID of the entity being audited.
	EntityID uint

	// Action is the type of action (create, update, delete).
	Action Action

	// UserID is the ID of the user who performed the action (optional).
	UserID *uint

	// OldValues contains the previous values before the change (for updates).
	OldValues map[string]any

	// NewValues contains the new values after the change.
	NewValues map[string]any

	// Metadata contains additional context about the action.
	Metadata map[string]any
}

// Logger defines the interface for audit logging.
type Logger interface {
	// Log records an audit entry.
	Log(ctx context.Context, entry Entry) error

	// LogCreate is a convenience method for logging create actions.
	LogCreate(ctx context.Context, entityType string, entityID uint, userID *uint, values map[string]any) error

	// LogUpdate is a convenience method for logging update actions.
	LogUpdate(ctx context.Context, entityType string, entityID uint, userID *uint, oldValues, newValues map[string]any) error

	// LogDelete is a convenience method for logging delete actions.
	LogDelete(ctx context.Context, entityType string, entityID uint, userID *uint, values map[string]any) error
}

// NewEntry creates a new audit entry with the given parameters.
func NewEntry(entityType string, entityID uint, action Action) Entry {
	return Entry{
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
	}
}

// WithUserID sets the user ID on the entry.
func (e Entry) WithUserID(userID uint) Entry {
	e.UserID = &userID
	return e
}

// WithOldValues sets the old values on the entry.
func (e Entry) WithOldValues(values map[string]any) Entry {
	e.OldValues = values
	return e
}

// WithNewValues sets the new values on the entry.
func (e Entry) WithNewValues(values map[string]any) Entry {
	e.NewValues = values
	return e
}

// WithMetadata sets the metadata on the entry.
func (e Entry) WithMetadata(metadata map[string]any) Entry {
	e.Metadata = metadata
	return e
}
