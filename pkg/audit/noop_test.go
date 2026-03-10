package audit_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/audit"
)

func TestNoopLogger_Log(t *testing.T) {
	t.Parallel()

	l := audit.NewNoop()
	err := l.Log(context.Background(), audit.Entry{})
	require.NoError(t, err)
}

func TestNoopLogger_LogCreate(t *testing.T) {
	t.Parallel()

	l := audit.NewNoop()
	userID := uint(1)
	err := l.LogCreate(context.Background(), "user", 1, &userID, map[string]any{"name": "test"})
	require.NoError(t, err)
}

func TestNoopLogger_LogUpdate(t *testing.T) {
	t.Parallel()

	l := audit.NewNoop()
	err := l.LogUpdate(context.Background(), "user", 1, nil, map[string]any{"name": "old"}, map[string]any{"name": "new"})
	require.NoError(t, err)
}

func TestNoopLogger_LogDelete(t *testing.T) {
	t.Parallel()

	l := audit.NewNoop()
	err := l.LogDelete(context.Background(), "user", 1, nil, map[string]any{"name": "test"})
	require.NoError(t, err)
}
