package integration_test

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/worker/tasks"
)

func assertQueuedTaskOptions(t *testing.T, kind, identity string) {
	t.Helper()
	addr := os.Getenv("TEST_ASYNQ_REDIS_ADDR")
	if addr == "" || os.Getenv("TEST_WORKER_ENABLED") == "true" {
		return
	}
	maxRetry, err := strconv.Atoi(os.Getenv("TEST_ASYNQ_MAX_RETRY"))
	require.NoError(t, err)
	timeout, err := time.ParseDuration(os.Getenv("TEST_ASYNQ_JOB_TIMEOUT"))
	require.NoError(t, err)
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: addr})
	defer func() { _ = inspector.Close() }()
	pending, err := inspector.ListPendingTasks("default", asynq.PageSize(100))
	require.NoError(t, err)
	for _, task := range pending {
		if task.Type != kind {
			continue
		}
		var email tasks.EmailPayload
		var image tasks.ImageProcessingPayload
		matched := false
		if kind == tasks.TypeEmailNotification {
			require.NoError(t, json.Unmarshal(task.Payload, &email))
			matched = email.To == identity
		} else {
			require.NoError(t, json.Unmarshal(task.Payload, &image))
			matched = strconv.FormatUint(uint64(image.MediaID), 10) == identity
		}
		if matched {
			require.Equal(t, maxRetry, task.MaxRetry, "configured retry limit must reach the queue")
			require.Equal(t, timeout, task.Timeout, "configured timeout must reach the queue")
			return
		}
	}
	t.Fatalf("task %s for %s was not queued", kind, identity)
}
