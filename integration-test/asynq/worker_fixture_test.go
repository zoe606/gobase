package asynq_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/repo/webapi/email"
	"go-boilerplate/internal/worker"
	"go-boilerplate/internal/worker/tasks"
	queue "go-boilerplate/pkg/asynq"
	"go-boilerplate/pkg/logger"
)

type redisFixture struct {
	addr      string
	name      string
	client    *queue.Client
	inspector *asynq.Inspector
	data      *redis.Client
	keys      []string
}

func newRedisFixture(t *testing.T) *redisFixture {
	t.Helper()
	addr := os.Getenv("ASYNQ_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("set ASYNQ_TEST_REDIS_ADDR to run real Redis worker checks")
	}
	options := asynq.RedisClientOpt{Addr: addr, DB: 15}
	f := &redisFixture{
		addr: addr, name: "foundation-" + uuid.NewString(),
		client:    queue.NewClient(queue.Config{RedisAddr: addr, RedisDB: 15}),
		inspector: asynq.NewInspector(options),
		data:      redis.NewClient(&redis.Options{Addr: addr, DB: 15}),
	}
	require.NoError(t, f.data.Ping(t.Context()).Err())
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = f.inspector.DeleteQueue(f.name, true)
		if len(f.keys) > 0 {
			_ = f.data.Del(ctx, f.keys...).Err()
		}
		_ = f.client.Close()
		_ = f.inspector.Close()
		_ = f.data.Close()
	})
	return f
}

func (f *redisFixture) identity() string {
	id := f.name + ":" + uuid.NewString()
	f.keys = append(f.keys, id+":attempts", id+":result", id+":effects")
	return id
}

func (f *redisFixture) enqueue(t *testing.T, id, mode string, options ...asynq.Option) *asynq.TaskInfo {
	t.Helper()
	task, err := tasks.NewEmailTask(id, mode, "", "fixture")
	require.NoError(t, err)
	options = append([]asynq.Option{asynq.Queue(f.name), asynq.Retention(time.Minute)}, options...)
	info, err := f.client.EnqueueContext(t.Context(), task, options...)
	require.NoError(t, err)
	return info
}

func (f *redisFixture) number(t *testing.T, key string) int {
	t.Helper()
	value, err := f.data.Get(t.Context(), key).Int()
	if errors.Is(err, redis.Nil) {
		return 0
	}
	require.NoError(t, err)
	return value
}

func (f *redisFixture) waitState(t *testing.T, id string, state asynq.TaskState) *asynq.TaskInfo {
	t.Helper()
	var result *asynq.TaskInfo
	require.Eventually(t, func() bool {
		info, err := f.inspector.GetTaskInfo(f.name, id)
		if err == nil && info.State == state {
			result = info
			return true
		}
		return false
	}, 15*time.Second, 50*time.Millisecond)
	return result
}

type workerProcess struct {
	cmd     *exec.Cmd
	done    chan struct{}
	err     error
	logPath string
}

func (f *redisFixture) startWorker(t *testing.T, addr string) *workerProcess {
	t.Helper()
	path := t.TempDir() + "/worker.log"
	log, err := os.Create(path)
	require.NoError(t, err)
	// The child runs only the fixture entry point in this test binary.
	cmd := exec.CommandContext(context.Background(), os.Args[0], "-test.run=^TestAsynqWorkerProcess$", "-test.timeout=5m") //nolint:gosec // The executable is the current test binary.
	cmd.Env = append(os.Environ(), "FOUNDATION_WORKER_PROCESS=1", "FOUNDATION_WORKER_ADDR="+addr, "FOUNDATION_WORKER_QUEUE="+f.name)
	cmd.Stdout, cmd.Stderr = log, log
	require.NoError(t, cmd.Start())
	p := &workerProcess{cmd: cmd, done: make(chan struct{}), logPath: path}
	go func() {
		p.err = cmd.Wait()
		_ = log.Close()
		close(p.done)
	}()
	t.Cleanup(func() {
		select {
		case <-p.done:
		default:
			_ = p.stop(syscall.SIGTERM)
		}
		if t.Failed() {
			contents, err := os.ReadFile(p.logPath)
			if err == nil {
				t.Logf("worker output:\n%s", contents)
			}
		}
	})
	return p
}

func (p *workerProcess) stop(sig os.Signal) error {
	select {
	case <-p.done:
		return p.err
	default:
	}
	if err := p.cmd.Process.Signal(sig); err != nil {
		return err
	}
	select {
	case <-p.done:
		return p.err
	case <-time.After(10 * time.Second):
		_ = p.cmd.Process.Kill()
		<-p.done
		return fmt.Errorf("worker did not stop; log: %s", p.logPath)
	}
}

type fixtureMailer struct{ data *redis.Client }

func (m *fixtureMailer) Send(ctx context.Context, message email.Email) error {
	id := message.To[0]
	attempt, err := m.data.Incr(ctx, id+":attempts").Result()
	if err != nil {
		return err
	}
	mode := message.Subject
	if strings.HasPrefix(mode, "fail:") {
		limit, parseErr := strconv.ParseInt(strings.TrimPrefix(mode, "fail:"), 10, 64)
		if parseErr != nil {
			return parseErr
		}
		if attempt <= limit {
			return errors.New("temporary fixture failure")
		}
	}
	if mode == "permanent" {
		return errors.New("permanent fixture failure")
	}
	if mode == "block" {
		<-ctx.Done()
		return ctx.Err()
	}
	if strings.HasPrefix(mode, "sleep:") {
		duration, parseErr := time.ParseDuration(strings.TrimPrefix(mode, "sleep:"))
		if parseErr != nil {
			return parseErr
		}
		select {
		case <-time.After(duration):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	// Only this controlled sender provides the idempotent effect used by the test.
	err = m.data.Eval(ctx, `if redis.call('SETNX', KEYS[1], 'sent') == 1 then redis.call('INCR', KEYS[2]) end; return 1`, []string{id + ":result", id + ":effects"}).Err()
	if err != nil {
		return err
	}
	if mode == "after-effect" && attempt == 1 {
		<-ctx.Done()
		return ctx.Err()
	}
	return nil
}

func TestAsynqWorkerProcess(t *testing.T) {
	if os.Getenv("FOUNDATION_WORKER_PROCESS") != "1" {
		t.Skip("worker subprocess entry point")
	}
	addr := os.Getenv("FOUNDATION_WORKER_ADDR")
	data := redis.NewClient(&redis.Options{Addr: addr, DB: 15})
	defer func() { _ = data.Close() }()
	l := logger.New("error")
	srv := queue.NewServer(queue.ServerConfig{
		RedisAddr: addr, RedisDB: 15, Concurrency: 1,
		Queues:          map[string]int{os.Getenv("FOUNDATION_WORKER_QUEUE"): 1},
		ShutdownTimeout: 500 * time.Millisecond, TaskCheckInterval: 100 * time.Millisecond,
		DelayedTaskCheckInterval: 200 * time.Millisecond,
		RetryDelayFunc:           func(_ int, _ error, _ *asynq.Task) time.Duration { return 100 * time.Millisecond },
	}, l)
	w := worker.New(srv, l, &fixtureMailer{data: data}, nil, nil)
	quit, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- w.Start() }()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-quit.Done():
		w.Shutdown()
		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(2 * time.Second):
			t.Fatal("worker shutdown did not return")
		}
	}
}
