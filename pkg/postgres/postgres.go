// Package postgres implements PostgreSQL connection using GORM with pgx driver.
package postgres

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	_defaultMaxPoolSize     = 10
	_defaultMaxIdleConns    = 5
	_defaultConnMaxLifetime = time.Hour
	_defaultConnMaxIdleTime = 30 * time.Minute
	_defaultConnAttempts    = 10
	_defaultConnTimeout     = time.Second
)

// Postgres holds the GORM database connection with pgx driver.
type Postgres struct {
	maxPoolSize     int
	maxIdleConns    int
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
	connAttempts    int
	connTimeout     time.Duration
	logLevel        logger.LogLevel

	DB *gorm.DB
}

// New creates a new GORM database connection using pgx driver.
func New(dsn string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:     _defaultMaxPoolSize,
		maxIdleConns:    _defaultMaxIdleConns,
		connMaxLifetime: _defaultConnMaxLifetime,
		connMaxIdleTime: _defaultConnMaxIdleTime,
		connAttempts:    _defaultConnAttempts,
		connTimeout:     _defaultConnTimeout,
		logLevel:        logger.Warn,
	}

	// Apply custom options
	for _, opt := range opts {
		opt(pg)
	}

	// GORM config
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(pg.logLevel),
	}

	// PostgreSQL driver config using pgx
	// gorm.io/driver/postgres uses pgx v5 as the underlying driver
	pgConfig := postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: false, // Use extended protocol for better performance
	}

	var db *gorm.DB
	var err error

	// Retry connection
	for pg.connAttempts > 0 {
		db, err = gorm.Open(postgres.New(pgConfig), gormConfig)
		if err == nil {
			break
		}

		pg.connAttempts--
		if pg.connAttempts == 0 {
			return nil, fmt.Errorf("postgres - New - connection failed: %w", err)
		}

		time.Sleep(pg.connTimeout)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("postgres - New - get sql.DB: %w", err)
	}

	// Configure connection pool (pgx pool settings via sql.DB)
	sqlDB.SetMaxOpenConns(pg.maxPoolSize)
	sqlDB.SetMaxIdleConns(pg.maxIdleConns)
	sqlDB.SetConnMaxLifetime(pg.connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(pg.connMaxIdleTime)

	// Test connection
	if err = sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("postgres - New - ping failed: %w", err)
	}

	pg.DB = db

	return pg, nil
}

// Close closes the database connection.
func (p *Postgres) Close() error {
	if p.DB != nil {
		sqlDB, err := p.DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// Ping checks if the database is reachable.
func (p *Postgres) Ping() error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Stats returns database statistics.
func (p *Postgres) Stats() (map[string]interface{}, error) {
	sqlDB, err := p.DB.DB()
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
