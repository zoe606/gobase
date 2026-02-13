// Package sqlite implements SQLite connection using GORM with pure Go driver.
package sqlite

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	_defaultConnMaxLifetime = time.Hour
	_defaultConnMaxIdleTime = 30 * time.Minute
)

// SQLite holds the GORM database connection with SQLite driver.
// SQLite holds the GORM database connection with SQLite driver.
type SQLite struct {
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
	logLevel        logger.LogLevel

	DB *gorm.DB
}

// New creates a new GORM database connection using pure Go SQLite driver.
func New(path string, opts ...Option) (*SQLite, error) {
	s := &SQLite{
		connMaxLifetime: _defaultConnMaxLifetime,
		connMaxIdleTime: _defaultConnMaxIdleTime,
		logLevel:        logger.Warn,
	}

	// Apply custom options
	for _, opt := range opts {
		opt(s)
	}

	// GORM config
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(s.logLevel),
	}

	// Open SQLite connection
	db, err := gorm.Open(sqlite.Open(path), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("sqlite - New - connection failed: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("sqlite - New - get sql.DB: %w", err)
	}

	// Configure connection pool
	// SQLite is single-writer, so we keep pool small
	sqlDB.SetMaxOpenConns(1) // SQLite only supports one writer
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(s.connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(s.connMaxIdleTime)

	// Enable WAL mode for better concurrent read performance
	if err := db.Exec("PRAGMA journal_mode=WAL").Error; err != nil {
		return nil, fmt.Errorf("sqlite - New - enable WAL: %w", err)
	}

	// Enable foreign keys
	if err := db.Exec("PRAGMA foreign_keys=ON").Error; err != nil {
		return nil, fmt.Errorf("sqlite - New - enable foreign keys: %w", err)
	}

	s.DB = db

	return s, nil
}

// Close closes the database connection.
func (s *SQLite) Close() error {
	if s.DB != nil {
		sqlDB, err := s.DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// Ping checks if the database is reachable.
func (s *SQLite) Ping() error {
	sqlDB, err := s.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Stats returns database statistics.
func (s *SQLite) Stats() (map[string]interface{}, error) {
	sqlDB, err := s.DB.DB()
	if err != nil {
		return nil, err
	}
	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
	}, nil
}
