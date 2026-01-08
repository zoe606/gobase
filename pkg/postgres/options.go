package postgres

import (
	"time"

	"gorm.io/gorm/logger"
)

// Option configures Postgres.
type Option func(*Postgres)

// MaxPoolSize sets the maximum number of open connections.
func MaxPoolSize(size int) Option {
	return func(p *Postgres) {
		p.maxPoolSize = size
	}
}

// MaxIdleConns sets the maximum number of idle connections.
func MaxIdleConns(size int) Option {
	return func(p *Postgres) {
		p.maxIdleConns = size
	}
}

// ConnMaxLifetime sets the maximum lifetime of a connection.
func ConnMaxLifetime(lifetime time.Duration) Option {
	return func(p *Postgres) {
		p.connMaxLifetime = lifetime
	}
}

// ConnMaxIdleTime sets the maximum idle time of a connection.
func ConnMaxIdleTime(idleTime time.Duration) Option {
	return func(p *Postgres) {
		p.connMaxIdleTime = idleTime
	}
}

// ConnAttempts sets the number of connection attempts.
func ConnAttempts(attempts int) Option {
	return func(p *Postgres) {
		p.connAttempts = attempts
	}
}

// ConnTimeout sets the timeout between connection attempts.
func ConnTimeout(timeout time.Duration) Option {
	return func(p *Postgres) {
		p.connTimeout = timeout
	}
}

// LogLevel sets the GORM log level.
func LogLevel(level logger.LogLevel) Option {
	return func(p *Postgres) {
		p.logLevel = level
	}
}
