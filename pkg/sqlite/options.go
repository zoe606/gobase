package sqlite

import (
	"time"

	"gorm.io/gorm/logger"
)

// Option configures SQLite.
type Option func(*SQLite)

// ConnMaxLifetime sets the maximum lifetime of a connection.
func ConnMaxLifetime(lifetime time.Duration) Option {
	return func(s *SQLite) {
		s.connMaxLifetime = lifetime
	}
}

// ConnMaxIdleTime sets the maximum idle time of a connection.
func ConnMaxIdleTime(idleTime time.Duration) Option {
	return func(s *SQLite) {
		s.connMaxIdleTime = idleTime
	}
}

// LogLevel sets the GORM log level.
func LogLevel(level logger.LogLevel) Option {
	return func(s *SQLite) {
		s.logLevel = level
	}
}
