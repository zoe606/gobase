package audit_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/audit"
)

func TestNewEntry(t *testing.T) {
	t.Parallel()

	entry := audit.NewEntry("user", 123, audit.ActionCreate)

	require.Equal(t, "user", entry.EntityType)
	require.Equal(t, uint(123), entry.EntityID)
	require.Equal(t, audit.ActionCreate, entry.Action)
	require.Nil(t, entry.UserID)
	require.Nil(t, entry.OldValues)
	require.Nil(t, entry.NewValues)
	require.Nil(t, entry.Metadata)
}

func TestEntry_WithUserID(t *testing.T) {
	t.Parallel()

	entry := audit.NewEntry("user", 123, audit.ActionUpdate).
		WithUserID(456)

	require.NotNil(t, entry.UserID)
	require.Equal(t, uint(456), *entry.UserID)
}

func TestEntry_WithOldValues(t *testing.T) {
	t.Parallel()

	oldValues := map[string]any{"name": "old"}
	entry := audit.NewEntry("user", 123, audit.ActionUpdate).
		WithOldValues(oldValues)

	require.Equal(t, oldValues, entry.OldValues)
}

func TestEntry_WithNewValues(t *testing.T) {
	t.Parallel()

	newValues := map[string]any{"name": "new"}
	entry := audit.NewEntry("user", 123, audit.ActionUpdate).
		WithNewValues(newValues)

	require.Equal(t, newValues, entry.NewValues)
}

func TestEntry_WithMetadata(t *testing.T) {
	t.Parallel()

	metadata := map[string]any{"ip": "127.0.0.1"}
	entry := audit.NewEntry("user", 123, audit.ActionCreate).
		WithMetadata(metadata)

	require.Equal(t, metadata, entry.Metadata)
}

func TestEntry_Chaining(t *testing.T) {
	t.Parallel()

	entry := audit.NewEntry("user", 123, audit.ActionUpdate).
		WithUserID(456).
		WithOldValues(map[string]any{"name": "old"}).
		WithNewValues(map[string]any{"name": "new"}).
		WithMetadata(map[string]any{"ip": "127.0.0.1"})

	require.Equal(t, "user", entry.EntityType)
	require.Equal(t, uint(123), entry.EntityID)
	require.Equal(t, audit.ActionUpdate, entry.Action)
	require.NotNil(t, entry.UserID)
	require.Equal(t, uint(456), *entry.UserID)
	require.Equal(t, "old", entry.OldValues["name"])
	require.Equal(t, "new", entry.NewValues["name"])
	require.Equal(t, "127.0.0.1", entry.Metadata["ip"])
}

func TestActions(t *testing.T) {
	t.Parallel()

	require.Equal(t, audit.Action("create"), audit.ActionCreate)
	require.Equal(t, audit.Action("update"), audit.ActionUpdate)
	require.Equal(t, audit.Action("delete"), audit.ActionDelete)
}
