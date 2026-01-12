package tx

import (
	"context"

	"gorm.io/gorm"
)

type contextKey string

const txKey contextKey = "gorm_tx"

// WithTx stores the transaction in context.
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

// FromContext retrieves the transaction from context, or returns nil if not present.
func FromContext(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx
	}
	return nil
}

// DBFromContext returns the transaction if present in context, otherwise returns the default db.
// This is the primary function repositories should use to get their database handle.
func DBFromContext(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx := FromContext(ctx); tx != nil {
		return tx
	}
	return defaultDB.WithContext(ctx)
}

// IsInTransaction returns true if the context contains a transaction.
func IsInTransaction(ctx context.Context) bool {
	return FromContext(ctx) != nil
}
