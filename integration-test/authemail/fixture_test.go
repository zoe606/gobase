package authemail_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"go-boilerplate/config"
	"go-boilerplate/internal/repo/persistent"
	"go-boilerplate/internal/repo/webapi/email"
	"go-boilerplate/internal/usecase/auth"
	"go-boilerplate/internal/worker"
	"go-boilerplate/internal/worker/tasks"
	queue "go-boilerplate/pkg/asynq"
	"go-boilerplate/pkg/audit"
	"go-boilerplate/pkg/logger"
	"go-boilerplate/pkg/postgres"
	"go-boilerplate/pkg/tx"
)

type authFixture struct {
	root, binary, prefix string
	redis                *redis.Client
	inspector            *asynq.Inspector
	client               *queue.Client
	db                   *gorm.DB
	env                  map[string]string
	appURL               string
	secrets              []string
}

func newAuthFixture(t *testing.T) *authFixture {
	t.Helper()
	if os.Getenv("AUTH_EMAIL_TEST_ENABLED") != "true" {
		t.Skip("set AUTH_EMAIL_TEST_ENABLED=true with dedicated PostgreSQL and Redis fixtures")
	}
	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	cfg, err := config.NewConfig("")
	require.NoError(t, err)
	pg, err := postgres.New(cfg.Postgres.DSN())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pg.Close()) })
	const redisDB = 14
	data := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr(), Password: cfg.Redis.Password, DB: redisDB})
	require.NoError(t, data.Ping(t.Context()).Err())
	count, err := data.DBSize(t.Context()).Result()
	require.NoError(t, err)
	require.Zero(t, count, "Redis database 14 must be empty and dedicated to this fixture")
	f := &authFixture{
		root: root, binary: filepath.Join(t.TempDir(), "app"), prefix: "auth-email-" + uuid.NewString(),
		redis: data, db: pg.DB,
		secrets:   []string{"replacement-password", "auth-email-malformed-sensitive"},
		inspector: asynq.NewInspector(asynq.RedisClientOpt{Addr: cfg.Redis.Addr(), Password: cfg.Redis.Password, DB: redisDB}),
		client:    queue.NewClient(queue.Config{RedisAddr: cfg.Redis.Addr(), RedisPassword: cfg.Redis.Password, RedisDB: redisDB}),
		env: map[string]string{
			"APP_ENV": "development", "REDIS_DB": strconv.Itoa(redisDB),
			"EMAIL_PROVIDER": "resend", "EMAIL_API_KEY": "controlled-fixture-key", "EMAIL_FROM_EMAIL": "sender@example.com",
			"EMAIL_VERIFICATION_ENABLED": "true", "EMAIL_VERIFICATION_AUTO_VERIFY": "false",
			"EMAIL_VERIFICATION_BASE_URL": "https://app.example/base", "PASSWORD_RESET_BASE_URL": "https://app.example/base",
			"JWT_SECRET_KEY": "auth-email-integration-secret", "STORAGE_DRIVER": "local",
			"STORAGE_LOCAL_PATH": t.TempDir(), "WEBSOCKET_ENABLED": "false", "GRPC_ENABLED": "false",
			"ASYNQ_MAX_RETRY": "2", "ASYNQ_JOB_TIMEOUT": "2s", "RATE_LIMIT_MAX": "1000",
			"LOG_LOG_REQUEST_BODY": "true", "LOG_LOG_RESPONSE_BODY": "false", "TELEMETRY_ENABLED": "false",
			"AUTH_EMAIL_CAPTURE_PREFIX": "",
		},
	}
	f.env["AUTH_EMAIL_CAPTURE_PREFIX"] = f.prefix
	t.Cleanup(func() {
		_ = f.inspector.DeleteQueue("default", true)
		// Database 14 was empty before startup and is reserved for these processes.
		keys, _ := data.Keys(context.Background(), "*").Result()
		if len(keys) > 0 {
			_ = data.Del(context.Background(), keys...).Err()
		}
		_ = f.client.Close()
		_ = f.inspector.Close()
		_ = data.Close()
	})
	build := exec.CommandContext(t.Context(), "go", "build", "-o", f.binary, "./cmd/app") //nolint:gosec // Build the fixture's own checkout.
	build.Dir = root
	output, err := build.CombinedOutput()
	require.NoError(t, err, string(output))
	f.appURL = f.startApp(t, nil)
	f.startWorker(t)
	return f
}

func authEnvironment(overrides map[string]string) []string {
	env := make([]string, 0, len(os.Environ())+len(overrides))
	for _, value := range os.Environ() {
		key, _, _ := strings.Cut(value, "=")
		if _, replaced := overrides[key]; !replaced {
			env = append(env, value)
		}
	}
	for key, value := range overrides {
		env = append(env, key+"="+value)
	}
	return env
}

func (f *authFixture) startApp(t *testing.T, overrides map[string]string) string {
	t.Helper()
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	env := make(map[string]string, len(f.env)+len(overrides)+1)
	for key, value := range f.env {
		env[key] = value
	}
	for key, value := range overrides {
		env[key] = value
	}
	env["HTTP_PORT"] = strconv.Itoa(port)
	f.startProcess(t, f.binary, nil, env)
	base := fmt.Sprintf("http://127.0.0.1:%d", port)
	require.Eventually(t, func() bool {
		request, requestErr := http.NewRequestWithContext(t.Context(), http.MethodGet, base+"/readyz", http.NoBody)
		if requestErr != nil {
			return false
		}
		response, requestErr := (&http.Client{Timeout: time.Second}).Do(request)
		if requestErr != nil {
			return false
		}
		_ = response.Body.Close()
		return response.StatusCode == http.StatusOK
	}, 30*time.Second, 100*time.Millisecond)
	return base
}

func (f *authFixture) startWorker(t *testing.T) {
	t.Helper()
	env := make(map[string]string, len(f.env)+1)
	for key, value := range f.env {
		env[key] = value
	}
	env["AUTH_EMAIL_WORKER_PROCESS"] = "true"
	f.startProcess(t, os.Args[0], []string{"-test.run=^TestAuthEmailWorkerProcess$", "-test.timeout=4m"}, env)
}

func (f *authFixture) startProcess(t *testing.T, binary string, args []string, env map[string]string) {
	t.Helper()
	log, err := os.CreateTemp(t.TempDir(), "process-*.log")
	require.NoError(t, err)
	cmd := exec.CommandContext(context.Background(), binary, args...) //nolint:gosec // Only the built app or current test binary is used.
	cmd.Dir, cmd.Env, cmd.Stdout, cmd.Stderr = f.root, authEnvironment(env), log, log
	require.NoError(t, cmd.Start())
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	t.Cleanup(func() {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		select {
		case processErr := <-done:
			require.NoError(t, processErr)
		case <-time.After(10 * time.Second):
			_ = cmd.Process.Kill()
			<-done
			t.Error("fixture process required forced shutdown")
		}
		require.NoError(t, log.Close())
		contents, readErr := os.ReadFile(log.Name())
		require.NoError(t, readErr)
		for _, secret := range f.secrets {
			require.False(t, strings.Contains(string(contents), secret), "auth token or password found in process logs")
		}
	})
}

type controlledAuthSender struct {
	data   *redis.Client
	prefix string
}

func (s *controlledAuthSender) Send(ctx context.Context, message email.Email) error {
	address := message.To[0]
	key := s.prefix + ":" + message.Subject + ":" + address
	attempt, err := s.data.Incr(ctx, key+":attempts").Result()
	if err != nil {
		return err
	}
	if message.Subject == "Reset your password" {
		if strings.HasPrefix(address, "fail-") || (strings.HasPrefix(address, "retry-") && attempt == 1) {
			return errors.New("controlled delivery failure")
		}
	}
	encoded, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return s.data.RPush(ctx, key+":messages", encoded).Err()
}

func TestAuthEmailWorkerProcess(t *testing.T) {
	if os.Getenv("AUTH_EMAIL_WORKER_PROCESS") != "true" {
		t.Skip("worker subprocess entry point")
	}
	cfg, err := config.NewConfig("")
	require.NoError(t, err)
	pg, err := postgres.New(cfg.Postgres.DSN())
	require.NoError(t, err)
	defer pg.Close()
	data := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr(), Password: cfg.Redis.Password, DB: cfg.Redis.DB})
	defer func() { _ = data.Close() }()
	sender := &controlledAuthSender{data: data, prefix: os.Getenv("AUTH_EMAIL_CAPTURE_PREFIX")}
	uc := auth.New(persistent.NewUserRepo(pg.DB), nil, persistent.NewRefreshTokenRepo(pg.DB), nil, audit.NewNoop()).
		WithAsynq(nil, cfg.App.Name).
		WithEmailVerification(persistent.NewEmailVerificationRepo(pg.DB), auth.VerificationConfig{
			Enabled: cfg.EmailVerification.Enabled, TokenTTL: cfg.EmailVerification.TokenTTL, BaseURL: cfg.EmailVerification.BaseURL,
		}).
		WithPasswordReset(persistent.NewPasswordResetRepo(pg.DB), auth.ResetConfig{TokenTTL: cfg.PasswordReset.TokenTTL, BaseURL: cfg.PasswordReset.BaseURL}).
		WithEmailDelivery(cfg.Email.CanSend(), cfg.App.IsProduction()).WithEmailSender(sender).WithTransactions(tx.New(pg.DB))
	l := logger.New("error")
	srv := queue.NewServer(queue.ServerConfig{
		RedisAddr: cfg.Redis.Addr(), RedisPassword: cfg.Redis.Password, RedisDB: cfg.Redis.DB, Concurrency: 2,
		Queues: map[string]int{"default": 1}, ShutdownTimeout: time.Second, TaskCheckInterval: 50 * time.Millisecond,
		DelayedTaskCheckInterval: 100 * time.Millisecond,
		RetryDelayFunc:           func(int, error, *asynq.Task) time.Duration { return 100 * time.Millisecond },
	}, l)
	w := worker.New(srv, l, sender, nil, nil).WithAuthEmail(uc)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- w.Start() }()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-ctx.Done():
		w.Shutdown()
		require.NoError(t, <-done)
	}
}

func (f *authFixture) mail(t *testing.T, subject, address string) email.Email {
	t.Helper()
	result, err := f.redis.BLPop(t.Context(), 15*time.Second, f.prefix+":"+subject+":"+address+":messages").Result()
	require.NoError(t, err)
	require.Len(t, result, 2)
	var message email.Email
	require.NoError(t, json.Unmarshal([]byte(result[1]), &message))
	for _, line := range strings.Split(message.Text, "\n") {
		link, parseErr := url.Parse(line)
		if parseErr != nil {
			continue
		}
		fragment, _ := url.ParseQuery(link.Fragment)
		if token := fragment.Get("token"); token != "" {
			f.secrets = append(f.secrets, token)
		}
	}
	return message
}

func (f *authFixture) task(t *testing.T, kind, address string, userID uint, state asynq.TaskState) *asynq.TaskInfo {
	t.Helper()
	var matched *asynq.TaskInfo
	require.Eventually(t, func() bool {
		var jobs []*asynq.TaskInfo
		var err error
		switch state {
		case asynq.TaskStateCompleted:
			jobs, err = f.inspector.ListCompletedTasks("default", asynq.PageSize(100))
		case asynq.TaskStateArchived:
			jobs, err = f.inspector.ListArchivedTasks("default", asynq.PageSize(100))
		default:
			return false
		}
		if err != nil {
			return false
		}
		for _, job := range jobs {
			if job.Type != kind {
				continue
			}
			var payload tasks.AuthEmailPayload
			if json.Unmarshal(job.Payload, &payload) == nil && payload.Email == address && payload.UserID == userID {
				matched = job
				return true
			}
		}
		return false
	}, 20*time.Second, 50*time.Millisecond)
	return matched
}
