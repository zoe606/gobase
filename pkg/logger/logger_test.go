package logger

import (
	"errors"
	"os"
	"os/exec"
	"testing"
)

var errTest = errors.New("test error")

func TestNew(t *testing.T) {
	t.Parallel()

	cases := []struct {
		level string
	}{
		{"debug"},
		{"info"},
		{"warn"},
		{"error"},
		{"unknown"}, // default path
	}

	for _, tc := range cases {
		l := New(tc.level)
		if l == nil || l.logger == nil {
			t.Fatalf("New(%q) returned nil logger", tc.level)
		}
	}
}

func TestNewDevelopment(t *testing.T) {
	t.Parallel()

	l := NewDevelopment()
	if l == nil || l.logger == nil {
		t.Fatal("NewDevelopment() returned nil logger")
	}
}

func TestLoggerMethods(t *testing.T) {
	t.Parallel()

	l := New("debug")

	// These should not panic
	l.Debug("debug message")
	l.Debug("debug message %s", "with arg")
	l.Info("info message")
	l.Info("info message %s", "with arg")
	l.Warn("warn message")
	l.Warn("warn message %s", "with arg")
	l.Error("error message")
	l.Error(errTest)
	l.Error("error message %s", "with arg")
}

func TestGetZapLogger(t *testing.T) {
	t.Parallel()

	l := New("info")
	zapLogger := l.GetZapLogger()

	if zapLogger == nil {
		t.Fatal("GetZapLogger() returned nil")
	}
}

func TestFatal_ExitsAndLogs(t *testing.T) {
	t.Parallel()

	if os.Getenv("LOGGER_FATAL_SUBPROC") == "1" {
		l := New("debug")
		l.Fatal("fatal now")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run", "TestFatal_ExitsAndLogs")
	cmd.Env = append(os.Environ(), "LOGGER_FATAL_SUBPROC=1")

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-nil error due to os.Exit in Fatal")
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if status := exitErr.ExitCode(); status != 1 {
			t.Fatalf("expected exit code 1, got %d", status)
		}
	}
}
