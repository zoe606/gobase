package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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

// New -.
func New(level string) *Logger {
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

	// Core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		l,
	)

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
	var msg string
	switch m := message.(type) {
	case error:
		msg = m.Error()
	case string:
		msg = m
	default:
		msg = "unknown message type"
	}

	switch level { //nolint:exhaustive // Only these levels are used in log method.
	case zapcore.DebugLevel:
		if len(args) == 0 {
			l.sugar.Debug(msg)
		} else {
			l.sugar.Debugf(msg, args...)
		}
	case zapcore.ErrorLevel:
		if len(args) == 0 {
			l.sugar.Error(msg)
		} else {
			l.sugar.Errorf(msg, args...)
		}
	case zapcore.FatalLevel:
		if len(args) == 0 {
			l.sugar.Fatal(msg)
		} else {
			l.sugar.Fatalf(msg, args...)
		}
	default:
		// Other levels use their dedicated methods (Info, Warn, etc.)
	}
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.logger.Sync()
}
