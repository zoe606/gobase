package entity

import (
	"time"
)

// AuditLog represents an audit log entry for tracking entity changes.
type AuditLog struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	EntityType string         `json:"entity_type" gorm:"size:100;not null;index:idx_audit_logs_entity"`
	EntityID   uint           `json:"entity_id" gorm:"not null;index:idx_audit_logs_entity"`
	Action     string         `json:"action" gorm:"size:20;not null"`
	UserID     *uint          `json:"user_id" gorm:"index"`
	User       *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	OldValues  map[string]any `json:"old_values,omitempty" gorm:"type:jsonb"`
	NewValues  map[string]any `json:"new_values,omitempty" gorm:"type:jsonb"`
	Metadata   map[string]any `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime"`
}

// TableName overrides the table name used by GORM.
func (AuditLog) TableName() string {
	return "audit_logs"
}
