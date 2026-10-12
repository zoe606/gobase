package asynq_test

import (
	"context"
	"syscall"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/worker/tasks"
	queue "go-boilerplate/pkg/asynq"
)

func TestAsynqRetryAndArchive(t *testing.T) {
	f := newRedisFixture(t)
	f.startWorker(t, f.addr)
	id := f.identity()
	task := f.enqueue(t, id, "fail:2", asynq.MaxRetry(2))
	result := f.waitState(t, task.ID, asynq.TaskStateCompleted)
	require.Equal(t, 2, result.Retried)
	require.Equal(t, 3, f.number(t, id+":attempts"))
	require.Equal(t, 1, f.number(t, id+":effects"))
	id = f.identity()
	task = f.enqueue(t, id, "permanent", asynq.MaxRetry(1))
	result = f.waitState(t, task.ID, asynq.TaskStateArchived)
	require.Equal(t, 1, result.Retried)
	require.Equal(t, 2, f.number(t, id+":attempts"))
	require.Zero(t, f.number(t, id+":effects"))
	bad, err := f.client.EnqueueContext(t.Context(), queue.NewTask(tasks.TypeEmailNotification, []byte("{invalid")), asynq.Queue(f.name), asynq.MaxRetry(3))
	require.NoError(t, err)
	result = f.waitState(t, bad.ID, asynq.TaskStateArchived)
	require.Zero(t, result.Retried, "malformed JSON cannot succeed after retry")
}

func TestAsynqPendingAndRepeatedExecution(t *testing.T) {
	f := newRedisFixture(t)
	id := f.identity()
	first := f.enqueue(t, id, "success", asynq.MaxRetry(1))
	info, err := f.inspector.GetTaskInfo(f.name, first.ID)
	require.NoError(t, err)
	require.Equal(t, asynq.TaskStatePending, info.State)
	p := f.startWorker(t, f.addr)
	f.waitState(t, first.ID, asynq.TaskStateCompleted)
	second := f.enqueue(t, id, "success", asynq.MaxRetry(1))
	require.NotEqual(t, first.ID, second.ID)
	f.waitState(t, second.ID, asynq.TaskStateCompleted)
	require.Equal(t, 2, f.number(t, id+":attempts"))
	require.Equal(t, 1, f.number(t, id+":effects"))
	require.NoError(t, p.stop(syscall.SIGTERM))
	pending := f.enqueue(t, f.identity(), "success", asynq.MaxRetry(1))
	f.startWorker(t, f.addr)
	f.waitState(t, pending.ID, asynq.TaskStateCompleted)
}

func TestAsynqActiveCrashRecovery(t *testing.T) {
	f := newRedisFixture(t)
	p := f.startWorker(t, f.addr)
	id := f.identity()
	task := f.enqueue(t, id, "after-effect", asynq.MaxRetry(2), asynq.Timeout(5*time.Minute))
	require.Eventually(t, func() bool { return f.number(t, id+":effects") == 1 }, 10*time.Second, 50*time.Millisecond)
	require.Error(t, p.stop(syscall.SIGKILL))
	f.startWorker(t, f.addr)
	require.Eventually(t, func() bool {
		info, err := f.inspector.GetTaskInfo(f.name, task.ID)
		return err == nil && info.State == asynq.TaskStateCompleted
	}, 130*time.Second, 200*time.Millisecond, "expired lease must recover after the killed worker")
	require.Equal(t, 2, f.number(t, id+":attempts"))
	require.Equal(t, 1, f.number(t, id+":effects"))
}

func TestAsynqTaskTimeoutAndActiveShutdown(t *testing.T) {
	f := newRedisFixture(t)
	p := f.startWorker(t, f.addr)
	timed := f.enqueue(t, f.identity(), "block", asynq.MaxRetry(0), asynq.Timeout(time.Second))
	info := f.waitState(t, timed.ID, asynq.TaskStateArchived)
	require.Zero(t, info.Retried)
	id := f.identity()
	active := f.enqueue(t, id, "sleep:200ms", asynq.MaxRetry(1))
	require.Eventually(t, func() bool { return f.number(t, id+":attempts") == 1 }, 10*time.Second, 20*time.Millisecond)
	require.NoError(t, p.stop(syscall.SIGTERM))
	f.waitState(t, active.ID, asynq.TaskStateCompleted)
	p = f.startWorker(t, f.addr)
	id = f.identity()
	unfinished := f.enqueue(t, id, "after-effect", asynq.MaxRetry(1))
	require.Eventually(t, func() bool { return f.number(t, id+":effects") == 1 }, 10*time.Second, 20*time.Millisecond)
	require.NoError(t, p.stop(syscall.SIGTERM))
	f.startWorker(t, f.addr)
	f.waitState(t, unfinished.ID, asynq.TaskStateCompleted)
	require.Equal(t, 1, f.number(t, id+":effects"))
}

func TestAsynqRedisFailureAndRecovery(t *testing.T) {
	f := newRedisFixture(t)
	proxy := newRedisProxy(t, f.addr)
	client := queue.NewClient(queue.Config{RedisAddr: proxy.addr(), RedisDB: 15})
	defer func() { _ = client.Close() }()
	f.startWorker(t, proxy.addr())
	task := f.enqueue(t, f.identity(), "success", asynq.MaxRetry(1))
	f.waitState(t, task.ID, asynq.TaskStateCompleted)
	proxy.available(false)
	ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
	start := time.Now()
	_, err := client.EnqueueContext(ctx, queue.NewTask("fixture:unavailable", nil), asynq.Queue(f.name))
	cancel()
	require.Error(t, err)
	require.Less(t, time.Since(start), 2*time.Second)
	pending := f.enqueue(t, f.identity(), "success", asynq.MaxRetry(1))
	proxy.available(true)
	f.waitState(t, pending.ID, asynq.TaskStateCompleted)
}
