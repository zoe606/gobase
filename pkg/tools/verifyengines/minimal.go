package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func checkMinimal(ctx context.Context, path string) error {
	if err := run(ctx, path, "go", "test", "-race", "./..."); err != nil {
		return err
	}
	if err := checkMinimalModules(ctx, path); err != nil {
		return err
	}
	if err := checkUnsupportedCommands(ctx, path); err != nil {
		return err
	}
	directory, err := os.MkdirTemp("", "gobase-minimal-runtime-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(directory) }()
	binary := filepath.Join(directory, "app")
	if err := run(ctx, path, "go", "build", "-o", binary, "./cmd/app"); err != nil {
		return err
	}
	if err := checkMinimalProcess(ctx, path, binary, directory); err != nil {
		return err
	}
	return checkBindFailure(ctx, path, binary)
}

func checkMinimalModules(ctx context.Context, path string) error {
	cmd := exec.CommandContext(ctx, "go", "list", "-m", "all")
	cmd.Dir = path
	data, err := cmd.Output()
	if err != nil {
		return err
	}
	for _, module := range []string{"gorm.io/", "github.com/redis/", "github.com/hibiken/asynq", "github.com/minio/", "github.com/resend/", "github.com/golang-jwt/", "github.com/golang-migrate/", "github.com/Conight/", "go.opentelemetry.io/", "github.com/air-verse/", "github.com/swaggo/"} {
		if strings.Contains(string(data), module) {
			return fmt.Errorf("unexpected minimal module %s", module)
		}
	}
	return nil
}

func checkUnsupportedCommands(ctx context.Context, path string) error {
	commands := [][]string{
		{"make", "gen"}, {"make", "gen-entity"}, {"make", "gen-full"}, {"make", "wire"},
		{"go", "run", "./pkg/codegen/cmd/codegen"}, {"go", "run", "./pkg/codegen/cmd/wire"},
	}
	for _, command := range commands {
		cmd := exec.CommandContext(ctx, command[0], command[1:]...) // #nosec G204 -- All command names and arguments are fixed verification literals.
		cmd.Dir = path
		data, err := cmd.CombinedOutput()
		if err == nil {
			return fmt.Errorf("unsupported command succeeded: %v", command)
		}
		if !strings.Contains(string(data), "full project profile") {
			return fmt.Errorf("expected clear unsupported command error for %v: %s: %w", command, data, err)
		}
	}
	for _, name := range []string{"migrations", "internal/entity", "internal/repo", "internal/usecase"} {
		if _, err := os.Stat(filepath.Join(path, name)); !os.IsNotExist(err) {
			return fmt.Errorf("unsupported commands created persistence output %s", name)
		}
	}
	return nil
}

type minimalProcess struct {
	cmd    *exec.Cmd
	exited chan struct{}
	err    error
}

func (p *minimalProcess) stop() {
	_ = p.cmd.Process.Kill()
	<-p.exited
}

func listenLocal(ctx context.Context) (net.Listener, error) {
	var config net.ListenConfig
	return config.Listen(ctx, "tcp", "127.0.0.1:0")
}

func listenerPort(listener net.Listener) (string, error) {
	_, port, err := net.SplitHostPort(listener.Addr().String())
	return port, err
}

func checkMinimalProcess(ctx context.Context, path, binary, directory string) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	listener, err := listenLocal(ctx)
	if err != nil {
		return err
	}
	port, err := listenerPort(listener)
	if err != nil {
		_ = listener.Close()
		return err
	}
	if err := listener.Close(); err != nil {
		return err
	}
	logPath := filepath.Join(directory, "app.log")
	log, err := os.Create(logPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = log.Close()
		if err != nil {
			data, _ := os.ReadFile(logPath)
			fmt.Fprintf(os.Stderr, "Minimal app log:\n%s\n", data)
		}
	}()
	p := &minimalProcess{cmd: exec.CommandContext(ctx, binary), exited: make(chan struct{})}
	p.cmd.Dir = path
	p.cmd.Env = append(os.Environ(), "APP_ENV=production", "HTTP_PORT="+port, "HTTP_SHUTDOWN_TIMEOUT=5s", "POSTGRES_HOST=invalid.invalid", "REDIS_HOST=invalid.invalid")
	p.cmd.Stdout, p.cmd.Stderr = log, log
	if err := p.cmd.Start(); err != nil {
		return err
	}
	go func() { p.err = p.cmd.Wait(); close(p.exited) }()
	defer p.stop()
	if err := checkMinimalHTTP(ctx, "http://127.0.0.1:"+port, p); err != nil {
		return err
	}
	if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}
	select {
	case <-p.exited:
		if p.err != nil {
			return fmt.Errorf("minimal shutdown: %w", p.err)
		}
		fmt.Println("Minimal startup, health, request ID, 404, and SIGTERM shutdown verified without infrastructure")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("minimal shutdown: %w", ctx.Err())
	}
}

func checkMinimalHTTP(ctx context.Context, base string, p *minimalProcess) error {
	client := &http.Client{Timeout: time.Second}
	for {
		if err := checkProbe(ctx, client, base+"/healthz", http.StatusOK, "OK"); err == nil {
			break
		}
		select {
		case <-p.exited:
			return fmt.Errorf("minimal app exited before readiness: %s", fmt.Sprint(p.err))
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	if err := checkProbe(ctx, client, base+"/readyz", http.StatusOK, "OK"); err != nil {
		return err
	}
	return checkProbe(ctx, client, base+"/missing", http.StatusNotFound, `"code":"NOT_FOUND"`)
}

func checkProbe(ctx context.Context, client *http.Client, url string, status int, body string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return err
	}
	request.Header.Set("X-Request-ID", "minimal-runtime")
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode != status || !strings.Contains(string(data), body) || response.Header.Get("X-Request-ID") != "minimal-runtime" {
		return fmt.Errorf("unexpected %s response: status %d, request ID %q, body %s", url, response.StatusCode, response.Header.Get("X-Request-ID"), data)
	}
	return nil
}

func checkBindFailure(ctx context.Context, path, binary string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var config net.ListenConfig
	listener, err := config.Listen(ctx, "tcp", ":0")
	if err != nil {
		return err
	}
	defer func() { _ = listener.Close() }()
	port, err := listenerPort(listener)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, binary)
	cmd.Dir = path
	cmd.Env = append(os.Environ(), "APP_ENV=production", "HTTP_PORT="+port)
	data, err := cmd.CombinedOutput()
	if err == nil {
		return fmt.Errorf("minimal app started on an occupied port")
	}
	if ctx.Err() != nil {
		return fmt.Errorf("minimal startup failure was not reported before timeout: %w", ctx.Err())
	}
	if !strings.Contains(string(data), "address already in use") {
		return fmt.Errorf("expected startup failure on occupied port: %s: %w", data, err)
	}
	fmt.Println("Minimal occupied-port startup failure verified")
	return nil
}
