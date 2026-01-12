// Package tx provides database transaction utilities with context-based detection.
package tx

import (
	"context"

	"gorm.io/gorm"
)

// Helper provides transaction management utilities.
type Helper interface {
	// RunInTx executes fn within a transaction.
	// If fn returns an error, the transaction is rolled back automatically.
	// If fn returns nil, the transaction is committed automatically.
	RunInTx(ctx context.Context, fn func(txCtx context.Context) error) error

	// RunInTxWithOptions executes fn within a transaction with custom options.
	RunInTxWithOptions(ctx context.Context, opts *TxOptions, fn func(txCtx context.Context) error) error
}

// TxOptions holds transaction options.
type TxOptions struct {
	// Propagation defines how transactions should propagate.
	// If true, nested RunInTx calls will use the same transaction.
	// If false, nested calls create savepoints (default GORM behavior).
	Propagation bool
}

// GormHelper implements Helper using GORM.
type GormHelper struct {
	db *gorm.DB
}

// New creates a new transaction helper.
func New(db *gorm.DB) *GormHelper {
	return &GormHelper{db: db}
}

// RunInTx executes the function within a transaction.
// GORM automatically commits on nil return and rollbacks on error return.
func (h *GormHelper) RunInTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	return h.RunInTxWithOptions(ctx, nil, fn)
}

// RunInTxWithOptions executes the function within a transaction with custom options.
func (h *GormHelper) RunInTxWithOptions(ctx context.Context, opts *TxOptions, fn func(txCtx context.Context) error) error {
	// Check if already in a transaction and propagation is enabled
	if opts != nil && opts.Propagation {
		if existingTx := FromContext(ctx); existingTx != nil {
			// Reuse existing transaction
			return fn(ctx)
		}
	}

	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := WithTx(ctx, tx)
		return fn(txCtx)
	})
}
