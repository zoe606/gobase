package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Default configuration values.
const (
	DefaultPostgresPort     = 5432
	DefaultPostgresPoolSize = 10
	DefaultPostgresIdleConn = 5
	DefaultRedisPort        = 6379
	DefaultCORSMaxAge       = 86400
	DefaultRateLimitMax     = 100
	DefaultAsynqConcurrency = 10
)

type (
	// Config holds all configuration for the application.
	Config struct {
		App               App               `mapstructure:"app"`
		HTTP              HTTP              `mapstructure:"http"`
		Log               Log               `mapstructure:"log"`
		Postgres          Postgres          `mapstructure:"postgres"`
		Redis             Redis             `mapstructure:"redis"`
		JWT               JWT               `mapstructure:"jwt"`
		CORS              CORS              `mapstructure:"cors"`
		RateLimit         RateLimit         `mapstructure:"rate_limit"`
		Asynq             Asynq             `mapstructure:"asynq"`
		Email             Email             `mapstructure:"email"`
		Metrics           Metrics           `mapstructure:"metrics"`
		Swagger           Swagger           `mapstructure:"swagger"`
		Storage           Storage           `mapstructure:"storage"`
		Cache             Cache             `mapstructure:"cache"`
		Lock              Lock              `mapstructure:"lock"`
		AuditLog          AuditLog          `mapstructure:"audit_log"`
		EmailVerification EmailVerification `mapstructure:"email_verification"`
		PasswordReset     PasswordReset     `mapstructure:"password_reset"`
		CircuitBreaker    CircuitBreaker    `mapstructure:"circuit_breaker"`
		Telemetry         Telemetry         `mapstructure:"telemetry"`
	}

	// App holds application-specific configuration.
	App struct {
		Name    string `mapstructure:"name"`
		Version string `mapstructure:"version"`
		Env     string `mapstructure:"env"`
	}

	// HTTP holds HTTP server configuration.
	HTTP struct {
		Port            string        `mapstructure:"port"`
		Timeout         time.Duration `mapstructure:"timeout"`          // Network read/write timeout
		IdleTimeout     time.Duration `mapstructure:"idle_timeout"`     // Keep-alive connection timeout
		RequestTimeout  time.Duration `mapstructure:"request_timeout"`  // Handler execution timeout
		BodyLimit       int           `mapstructure:"body_limit"`       // Max request body size in bytes (default: 4MB)
		ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"` // Graceful shutdown timeout
	}

	// Log holds logging configuration.
	Log struct {
		Level           string `mapstructure:"level"`
		File            string `mapstructure:"file"`              // Optional file path for log output (empty = stdout only)
		LogRequestBody  bool   `mapstructure:"log_request_body"`  // Log request bodies (default: false)
		LogResponseBody bool   `mapstructure:"log_response_body"` // Log response bodies (default: false)
		RedactFields    string `mapstructure:"redact_fields"`     // Comma-separated fields to redact from body logs
	}

	// Postgres holds PostgreSQL configuration.
	Postgres struct {
		Host            string        `mapstructure:"host"`
		Port            int           `mapstructure:"port"`
		User            string        `mapstructure:"user"`
		Password        string        `mapstructure:"password"`
		DBName          string        `mapstructure:"dbname"`
		SSLMode         string        `mapstructure:"sslmode"`
		MaxPoolSize     int           `mapstructure:"max_pool_size"`
		MaxIdleConns    int           `mapstructure:"max_idle_conns"`
		ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
		ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
	}

	// Metrics holds metrics configuration.
	Metrics struct {
		Enabled bool `mapstructure:"enabled"`
	}

	// Swagger holds Swagger documentation configuration.
	Swagger struct {
		Enabled bool `mapstructure:"enabled"`
	}

	// Redis holds Redis configuration.
	Redis struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
	}

	// JWT holds JWT configuration.
	JWT struct {
		SecretKey      string        `mapstructure:"secret_key"`
		AccessExpiry   time.Duration `mapstructure:"access_expiry"`
		RefreshExpiry  time.Duration `mapstructure:"refresh_expiry"`
		Algorithm      string        `mapstructure:"algorithm"`        // hs256, rs256, es256
		PrivateKeyPath string        `mapstructure:"private_key_path"` // Path to private key (RS256/ES256)
		PublicKeyPath  string        `mapstructure:"public_key_path"`  // Path to public key (RS256/ES256)
	}

	// CORS holds CORS configuration.
	CORS struct {
		AllowOrigins     string `mapstructure:"allow_origins"`
		AllowMethods     string `mapstructure:"allow_methods"`
		AllowHeaders     string `mapstructure:"allow_headers"`
		AllowCredentials bool   `mapstructure:"allow_credentials"`
		MaxAge           int    `mapstructure:"max_age"`
	}

	// RateLimit holds rate limiting configuration.
	RateLimit struct {
		Max        int           `mapstructure:"max"`
		Expiration time.Duration `mapstructure:"expiration"`
		Store      string        `mapstructure:"store"` // "memory" (default) or "redis" (Phase 3)
	}

	// Asynq holds Asynq task queue configuration.
	Asynq struct {
		Concurrency int           `mapstructure:"concurrency"`
		JobTimeout  time.Duration `mapstructure:"job_timeout"` // Default job timeout
		MaxRetry    int           `mapstructure:"max_retry"`   // Max retry attempts before archiving
	}

	// Email holds email service configuration.
	Email struct {
		Provider  string `mapstructure:"provider"`   // resend, brevo
		APIKey    string `mapstructure:"api_key"`    // API key for the email provider
		FromEmail string `mapstructure:"from_email"` // Sender email address
		FromName  string `mapstructure:"from_name"`  // Sender name
	}

	// Storage holds file storage configuration.
	Storage struct {
		Driver      string `mapstructure:"driver"`        // "local", "s3"
		MaxSize     int64  `mapstructure:"max_size"`      // Max upload size in bytes
		LocalPath   string `mapstructure:"local_path"`    // Local storage base path
		LocalURL    string `mapstructure:"local_url"`     // Local storage public URL
		S3Endpoint  string `mapstructure:"s3_endpoint"`   // S3/MinIO endpoint
		S3Bucket    string `mapstructure:"s3_bucket"`     // S3/MinIO bucket name
		S3Region    string `mapstructure:"s3_region"`     // S3 region
		S3AccessKey string `mapstructure:"s3_access_key"` // S3/MinIO access key
		S3SecretKey string `mapstructure:"s3_secret_key"` // S3/MinIO secret key
		S3UseSSL    bool   `mapstructure:"s3_use_ssl"`    // Use HTTPS for S3
	}

	// Cache holds caching configuration.
	Cache struct {
		Enabled bool          `mapstructure:"enabled"` // Enable/disable caching
		TTL     time.Duration `mapstructure:"ttl"`     // Default cache TTL
		Prefix  string        `mapstructure:"prefix"`  // Key prefix for cache entries
	}

	// AuditLog holds audit logging configuration.
	AuditLog struct {
		Enabled bool `mapstructure:"enabled"` // Enable/disable audit logging
	}

	// Lock holds distributed lock configuration.
	Lock struct {
		Provider string `mapstructure:"provider"` // "noop" (default) or "redis"
	}

	// EmailVerification holds email verification configuration.
	EmailVerification struct {
		Enabled    bool          `mapstructure:"enabled"`     // Enable/disable email verification
		AutoVerify bool          `mapstructure:"auto_verify"` // Auto-verify in development
		TokenTTL   time.Duration `mapstructure:"token_ttl"`   // Verification token TTL
		BaseURL    string        `mapstructure:"base_url"`    // Frontend base URL for verification links
	}

	// PasswordReset holds password reset configuration.
	PasswordReset struct {
		TokenTTL time.Duration `mapstructure:"token_ttl"` // Reset token TTL
		BaseURL  string        `mapstructure:"base_url"`  // Frontend base URL for reset links
	}

	// CircuitBreaker holds circuit breaker configuration.
	CircuitBreaker struct {
		Enabled      bool          `mapstructure:"enabled"`       // Enable/disable circuit breaker
		MaxRequests  uint32        `mapstructure:"max_requests"`  // Max requests in half-open state
		Interval     time.Duration `mapstructure:"interval"`      // Reset interval in closed state
		Timeout      time.Duration `mapstructure:"timeout"`       // Time in open state before half-open
		FailureRatio float64       `mapstructure:"failure_ratio"` // Failure ratio to trip (0.0-1.0)
		MinRequests  uint32        `mapstructure:"min_requests"`  // Min requests before evaluating ratio
	}

	// Telemetry holds OpenTelemetry configuration.
	Telemetry struct {
		Enabled      bool   `mapstructure:"enabled"`       // Enable/disable OpenTelemetry
		OTLPEndpoint string `mapstructure:"otlp_endpoint"` // OTLP gRPC endpoint
		OTLPInsecure bool   `mapstructure:"otlp_insecure"` // Use insecure gRPC connection
	}
)

// DSN returns the PostgreSQL connection string.
func (p *Postgres) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DBName, p.SSLMode,
	)
}

// URL returns the PostgreSQL connection URL.
func (p *Postgres) URL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.DBName, p.SSLMode,
	)
}

// Addr returns the Redis address in host:port format.
func (r *Redis) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// NewConfig reads configuration from file and environment variables.
func NewConfig(configPath string) (*Config, error) {
	cfg := &Config{}

	// Set defaults
	setDefaults()

	// Config file settings
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("/etc/app")
	}

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("config error: %w", err)
		}
		// Config file not found, will use defaults and env vars
	}

	// Environment variables
	viper.SetEnvPrefix("")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Bind specific environment variables
	bindEnvVars()

	// Unmarshal config
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("config unmarshal error: %w", err)
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks required configuration values.
func (c *Config) Validate() error {
	var errs []string

	// JWT secret is required in production
	if c.App.IsProduction() {
		if c.JWT.SecretKey == "" || c.JWT.SecretKey == "change-me-in-production" {
			errs = append(errs, "JWT_SECRET_KEY must be set in production")
		}

		// PostgreSQL SSL is required in production
		if c.Postgres.SSLMode == "disable" {
			errs = append(errs, "POSTGRES_SSLMODE must not be 'disable' in production (use 'require', 'verify-ca', or 'verify-full')")
		}
	}

	// JWT algorithm validation
	switch c.JWT.Algorithm {
	case "hs256", "":
		// HS256 is default, no extra validation needed here
	case "rs256", "es256":
		if c.JWT.PrivateKeyPath == "" {
			errs = append(errs, "JWT_PRIVATE_KEY_PATH is required for "+c.JWT.Algorithm)
		}
		if c.JWT.PublicKeyPath == "" {
			errs = append(errs, "JWT_PUBLIC_KEY_PATH is required for "+c.JWT.Algorithm)
		}
	default:
		errs = append(errs, fmt.Sprintf("JWT_ALGORITHM must be hs256, rs256, or es256 (got %q)", c.JWT.Algorithm))
	}

	// Database config validation
	if c.Postgres.Host == "" {
		errs = append(errs, "POSTGRES_HOST is required")
	}
	if c.Postgres.DBName == "" {
		errs = append(errs, "POSTGRES_DBNAME is required")
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

// Warnings returns non-fatal configuration warnings.
func (c *Config) Warnings() []string {
	var warnings []string

	if c.App.IsProduction() {
		if c.CORS.AllowOrigins == "*" {
			warnings = append(warnings, "CORS_ALLOW_ORIGINS is set to '*' in production — consider restricting to specific origins")
		}

		if !c.Storage.S3UseSSL && c.Storage.Driver == "s3" {
			warnings = append(warnings, "STORAGE_S3_USE_SSL is false in production — consider enabling SSL for S3 connections")
		}
	}

	return warnings
}

func setDefaults() {
	// App defaults
	viper.SetDefault("app.name", "go-boilerplate")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.env", "development")

	// HTTP defaults
	viper.SetDefault("http.port", "8080")
	viper.SetDefault("http.timeout", "15s")          // Network read/write
	viper.SetDefault("http.idle_timeout", "60s")     // Keep-alive connections
	viper.SetDefault("http.request_timeout", "30s")  // Handler execution
	viper.SetDefault("http.body_limit", 4*1024*1024) // 4MB
	viper.SetDefault("http.shutdown_timeout", "15s") // Graceful shutdown

	// Log defaults
	viper.SetDefault("log.level", "debug")
	viper.SetDefault("log.file", "")                                                         // Empty = stdout only
	viper.SetDefault("log.log_request_body", false)                                          //nolint:revive // explicit false
	viper.SetDefault("log.log_response_body", false)                                         //nolint:revive // explicit false
	viper.SetDefault("log.redact_fields", "password,token,secret,authorization,credit_card") //nolint:revive // default

	// Postgres defaults
	viper.SetDefault("postgres.host", "localhost")
	viper.SetDefault("postgres.port", DefaultPostgresPort)
	viper.SetDefault("postgres.user", "postgres")
	viper.SetDefault("postgres.password", "postgres")
	viper.SetDefault("postgres.dbname", "app")
	viper.SetDefault("postgres.sslmode", "disable")
	viper.SetDefault("postgres.max_pool_size", DefaultPostgresPoolSize)
	viper.SetDefault("postgres.max_idle_conns", DefaultPostgresIdleConn)
	viper.SetDefault("postgres.conn_max_lifetime", "1h")
	viper.SetDefault("postgres.conn_max_idle_time", "30m")

	// Metrics defaults
	viper.SetDefault("metrics.enabled", true)

	// Swagger defaults
	viper.SetDefault("swagger.enabled", true)

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", DefaultRedisPort)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// JWT defaults
	viper.SetDefault("jwt.secret_key", "change-me-in-production")
	viper.SetDefault("jwt.access_expiry", "15m")
	viper.SetDefault("jwt.refresh_expiry", "168h")
	viper.SetDefault("jwt.algorithm", "hs256")
	viper.SetDefault("jwt.private_key_path", "")
	viper.SetDefault("jwt.public_key_path", "")

	// CORS defaults
	viper.SetDefault("cors.allow_origins", "*")
	viper.SetDefault("cors.allow_methods", "GET,POST,PUT,DELETE,OPTIONS,PATCH")
	viper.SetDefault("cors.allow_headers", "Origin,Content-Type,Accept,Authorization,X-Request-ID")
	viper.SetDefault("cors.allow_credentials", false)
	viper.SetDefault("cors.max_age", DefaultCORSMaxAge)

	// RateLimit defaults
	viper.SetDefault("rate_limit.max", DefaultRateLimitMax)
	viper.SetDefault("rate_limit.expiration", "1m")
	viper.SetDefault("rate_limit.store", "memory")

	// Asynq defaults
	viper.SetDefault("asynq.concurrency", DefaultAsynqConcurrency)
	viper.SetDefault("asynq.job_timeout", "5m")
	viper.SetDefault("asynq.max_retry", 3)

	// Email defaults
	viper.SetDefault("email.provider", "resend")
	viper.SetDefault("email.api_key", "")
	viper.SetDefault("email.from_email", "noreply@example.com")
	viper.SetDefault("email.from_name", "Go Boilerplate")

	// Storage defaults
	viper.SetDefault("storage.driver", "s3")
	viper.SetDefault("storage.max_size", 10485760) // 10MB
	viper.SetDefault("storage.local_path", "./uploads")
	viper.SetDefault("storage.local_url", "http://localhost:8080/uploads")
	viper.SetDefault("storage.s3_endpoint", "localhost:9000")
	viper.SetDefault("storage.s3_bucket", "app-uploads")
	viper.SetDefault("storage.s3_region", "us-east-1")
	viper.SetDefault("storage.s3_access_key", "minioadmin")
	viper.SetDefault("storage.s3_secret_key", "minioadmin")
	viper.SetDefault("storage.s3_use_ssl", false)

	// Cache defaults
	viper.SetDefault("cache.enabled", true)
	viper.SetDefault("cache.ttl", "5m")
	viper.SetDefault("cache.prefix", "app:")

	// Lock defaults
	viper.SetDefault("lock.provider", "noop")

	// AuditLog defaults
	viper.SetDefault("audit_log.enabled", false)

	// EmailVerification defaults
	viper.SetDefault("email_verification.enabled", true)
	viper.SetDefault("email_verification.auto_verify", true) // Auto-verify in development
	viper.SetDefault("email_verification.token_ttl", "24h")
	viper.SetDefault("email_verification.base_url", "http://localhost:3000")

	// PasswordReset defaults
	viper.SetDefault("password_reset.token_ttl", "1h")
	viper.SetDefault("password_reset.base_url", "http://localhost:3000")

	// Telemetry defaults
	viper.SetDefault("telemetry.enabled", false)
	viper.SetDefault("telemetry.otlp_endpoint", "localhost:4317")
	viper.SetDefault("telemetry.otlp_insecure", true)

	// CircuitBreaker defaults
	viper.SetDefault("circuit_breaker.enabled", true)
	viper.SetDefault("circuit_breaker.max_requests", 3)
	viper.SetDefault("circuit_breaker.interval", "60s")
	viper.SetDefault("circuit_breaker.timeout", "30s")
	viper.SetDefault("circuit_breaker.failure_ratio", 0.5)
	viper.SetDefault("circuit_breaker.min_requests", 10)
}

//nolint:errcheck // BindEnv only errors if key is empty, which is controlled by us.
func bindEnvVars() {
	// App
	viper.BindEnv("app.name", "APP_NAME")
	viper.BindEnv("app.version", "APP_VERSION")
	viper.BindEnv("app.env", "APP_ENV")

	// HTTP
	viper.BindEnv("http.port", "HTTP_PORT")
	viper.BindEnv("http.timeout", "HTTP_TIMEOUT")
	viper.BindEnv("http.idle_timeout", "HTTP_IDLE_TIMEOUT")
	viper.BindEnv("http.request_timeout", "HTTP_REQUEST_TIMEOUT")
	viper.BindEnv("http.body_limit", "HTTP_BODY_LIMIT")
	viper.BindEnv("http.shutdown_timeout", "HTTP_SHUTDOWN_TIMEOUT")

	// Log
	viper.BindEnv("log.level", "LOG_LEVEL")
	viper.BindEnv("log.file", "LOG_FILE")
	viper.BindEnv("log.log_request_body", "LOG_REQUEST_BODY")
	viper.BindEnv("log.log_response_body", "LOG_RESPONSE_BODY")
	viper.BindEnv("log.redact_fields", "LOG_REDACT_FIELDS")

	// Postgres
	viper.BindEnv("postgres.host", "POSTGRES_HOST", "DB_HOST")
	viper.BindEnv("postgres.port", "POSTGRES_PORT", "DB_PORT")
	viper.BindEnv("postgres.user", "POSTGRES_USER", "DB_USER")
	viper.BindEnv("postgres.password", "POSTGRES_PASSWORD", "DB_PASSWORD")
	viper.BindEnv("postgres.dbname", "POSTGRES_DBNAME", "DB_NAME")
	viper.BindEnv("postgres.sslmode", "POSTGRES_SSLMODE", "DB_SSLMODE")
	viper.BindEnv("postgres.max_pool_size", "POSTGRES_MAX_POOL_SIZE", "DB_MAX_POOL_SIZE")

	// Metrics
	viper.BindEnv("metrics.enabled", "METRICS_ENABLED")

	// Swagger
	viper.BindEnv("swagger.enabled", "SWAGGER_ENABLED")

	// Redis
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")

	// JWT
	viper.BindEnv("jwt.secret_key", "JWT_SECRET_KEY")
	viper.BindEnv("jwt.access_expiry", "JWT_ACCESS_EXPIRY")
	viper.BindEnv("jwt.refresh_expiry", "JWT_REFRESH_EXPIRY")
	viper.BindEnv("jwt.algorithm", "JWT_ALGORITHM")
	viper.BindEnv("jwt.private_key_path", "JWT_PRIVATE_KEY_PATH")
	viper.BindEnv("jwt.public_key_path", "JWT_PUBLIC_KEY_PATH")

	// CORS
	viper.BindEnv("cors.allow_origins", "CORS_ALLOW_ORIGINS")
	viper.BindEnv("cors.allow_methods", "CORS_ALLOW_METHODS")
	viper.BindEnv("cors.allow_headers", "CORS_ALLOW_HEADERS")
	viper.BindEnv("cors.allow_credentials", "CORS_ALLOW_CREDENTIALS")
	viper.BindEnv("cors.max_age", "CORS_MAX_AGE")

	// RateLimit
	viper.BindEnv("rate_limit.max", "RATE_LIMIT_MAX")
	viper.BindEnv("rate_limit.expiration", "RATE_LIMIT_EXPIRATION")
	viper.BindEnv("rate_limit.store", "RATE_LIMIT_STORE")

	// Asynq
	viper.BindEnv("asynq.concurrency", "ASYNQ_CONCURRENCY")
	viper.BindEnv("asynq.job_timeout", "ASYNQ_JOB_TIMEOUT")
	viper.BindEnv("asynq.max_retry", "ASYNQ_MAX_RETRY")

	// Email
	viper.BindEnv("email.provider", "EMAIL_PROVIDER")
	viper.BindEnv("email.api_key", "EMAIL_API_KEY")
	viper.BindEnv("email.from_email", "EMAIL_FROM_EMAIL")
	viper.BindEnv("email.from_name", "EMAIL_FROM_NAME")

	// Storage
	viper.BindEnv("storage.driver", "STORAGE_DRIVER")
	viper.BindEnv("storage.max_size", "STORAGE_MAX_SIZE")
	viper.BindEnv("storage.local_path", "STORAGE_LOCAL_PATH")
	viper.BindEnv("storage.local_url", "STORAGE_LOCAL_URL")
	viper.BindEnv("storage.s3_endpoint", "STORAGE_S3_ENDPOINT")
	viper.BindEnv("storage.s3_bucket", "STORAGE_S3_BUCKET")
	viper.BindEnv("storage.s3_region", "STORAGE_S3_REGION")
	viper.BindEnv("storage.s3_access_key", "STORAGE_S3_ACCESS_KEY")
	viper.BindEnv("storage.s3_secret_key", "STORAGE_S3_SECRET_KEY")
	viper.BindEnv("storage.s3_use_ssl", "STORAGE_S3_USE_SSL")

	// Cache
	viper.BindEnv("cache.enabled", "CACHE_ENABLED")
	viper.BindEnv("cache.ttl", "CACHE_TTL")
	viper.BindEnv("cache.prefix", "CACHE_PREFIX")

	// Lock
	viper.BindEnv("lock.provider", "LOCK_PROVIDER")

	// AuditLog
	viper.BindEnv("audit_log.enabled", "AUDIT_LOG_ENABLED")

	// EmailVerification
	viper.BindEnv("email_verification.enabled", "EMAIL_VERIFICATION_ENABLED")
	viper.BindEnv("email_verification.auto_verify", "EMAIL_VERIFICATION_AUTO_VERIFY")
	viper.BindEnv("email_verification.token_ttl", "EMAIL_VERIFICATION_TOKEN_TTL")
	viper.BindEnv("email_verification.base_url", "EMAIL_VERIFICATION_BASE_URL")

	// PasswordReset
	viper.BindEnv("password_reset.token_ttl", "PASSWORD_RESET_TOKEN_TTL")
	viper.BindEnv("password_reset.base_url", "PASSWORD_RESET_BASE_URL")

	// Telemetry
	viper.BindEnv("telemetry.enabled", "TELEMETRY_ENABLED")
	viper.BindEnv("telemetry.otlp_endpoint", "TELEMETRY_OTLP_ENDPOINT")
	viper.BindEnv("telemetry.otlp_insecure", "TELEMETRY_OTLP_INSECURE")

	// CircuitBreaker
	viper.BindEnv("circuit_breaker.enabled", "CIRCUIT_BREAKER_ENABLED")
	viper.BindEnv("circuit_breaker.max_requests", "CIRCUIT_BREAKER_MAX_REQUESTS")
	viper.BindEnv("circuit_breaker.interval", "CIRCUIT_BREAKER_INTERVAL")
	viper.BindEnv("circuit_breaker.timeout", "CIRCUIT_BREAKER_TIMEOUT")
	viper.BindEnv("circuit_breaker.failure_ratio", "CIRCUIT_BREAKER_FAILURE_RATIO")
	viper.BindEnv("circuit_breaker.min_requests", "CIRCUIT_BREAKER_MIN_REQUESTS")
}
