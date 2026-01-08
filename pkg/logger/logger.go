package logger

import (
	"io"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Option is a functional option for configuring the logger.
type Option func(*options)

type options struct {
	fileOutput io.Writer
}

// WithFileOutput adds file output alongside stdout.
// The caller is responsible for closing the file when done.
func WithFileOutput(w io.Writer) Option {
	return func(o *options) {
		o.fileOutput = w
	}
}

// Interface -.
type Interface interface {
	Debug(message interface{}, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message interface{}, args ...interface{})
	Fatal(message interface{}, args ...interface{})
	GetZapLogger() *zap.Logger
}

// Logger -.
type Logger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

var _ Interface = (*Logger)(nil)

// New creates a new logger with the specified level and options.
func New(level string, opts ...Option) *Logger {
	// Apply options
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	var l zapcore.Level

	switch strings.ToLower(level) {
	case "error":
		l = zapcore.ErrorLevel
	case "warn":
		l = zapcore.WarnLevel
	case "info":
		l = zapcore.InfoLevel
	case "debug":
		l = zapcore.DebugLevel
	default:
		l = zapcore.InfoLevel
	}

	// Encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// Build write syncer(s)
	var ws zapcore.WriteSyncer
	if o.fileOutput != nil {
		// Multi-write to both stdout and file
		ws = zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout),
			zapcore.AddSync(o.fileOutput),
		)
	} else {
		ws = zapcore.AddSync(os.Stdout)
	}

	// Core
	core := zapcore.NewCore(encoder, ws, l)

	// Logger with caller info
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &Logger{
		logger: logger,
		sugar:  logger.Sugar(),
	}
}

// NewDevelopment creates a development logger with console output.
func NewDevelopment() *Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	logger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic("failed to create development logger: " + err.Error())
	}

	return &Logger{
		logger: logger,
		sugar:  logger.Sugar(),
	}
}

// GetZapLogger returns the underlying zap logger for integrations.
func (l *Logger) GetZapLogger() *zap.Logger {
	return l.logger
}

// Debug -.
func (l *Logger) Debug(message interface{}, args ...interface{}) {
	l.log(zapcore.DebugLevel, message, args...)
}

// Info -.
func (l *Logger) Info(message string, args ...interface{}) {
	if len(args) == 0 {
		l.sugar.Info(message)
	} else {
		l.sugar.Infof(message, args...)
	}
}

// Warn -.
func (l *Logger) Warn(message string, args ...interface{}) {
	if len(args) == 0 {
		l.sugar.Warn(message)
	} else {
		l.sugar.Warnf(message, args...)
	}
}

// Error -.
func (l *Logger) Error(message interface{}, args ...interface{}) {
	l.log(zapcore.ErrorLevel, message, args...)
}

// Fatal -.
func (l *Logger) Fatal(message interface{}, args ...interface{}) {
	l.log(zapcore.FatalLevel, message, args...)
}

func (l *Logger) log(level zapcore.Level, message interface{}, args ...interface{}) {
	msg := formatMessage(message)
	logAtLevel(l.sugar, level, msg, args...)
}

// formatMessage converts various message types to string.
func formatMessage(message interface{}) string {
	switch m := message.(type) {
	case error:
		return m.Error()
	case string:
		return m
	default:
		return "unknown message type"
	}
}

// logAtLevel logs a message at the specified level.
//
//nolint:exhaustive // Only these levels are used in log method.
func logAtLevel(sugar *zap.SugaredLogger, level zapcore.Level, msg string, args ...interface{}) {
	switch level {
	case zapcore.DebugLevel:
		if len(args) == 0 {
			sugar.Debug(msg)
		} else {
			sugar.Debugf(msg, args...)
		}
	case zapcore.ErrorLevel:
		if len(args) == 0 {
			sugar.Error(msg)
		} else {
			sugar.Errorf(msg, args...)
		}
	case zapcore.FatalLevel:
		if len(args) == 0 {
			sugar.Fatal(msg)
		} else {
			sugar.Fatalf(msg, args...)
		}
	}
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.logger.Sync()
}
