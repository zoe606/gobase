package entity_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/entity"
)

func TestAuditLog_TableName(t *testing.T) {
	t.Parallel()
	require.Equal(t, "audit_logs", entity.AuditLog{}.TableName())
}
