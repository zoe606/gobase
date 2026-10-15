package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmail_CanSend(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, provider, key, from string
		want                      bool
	}{
		{"resend", "resend", "key", "sender@example.com", true},
		{"default provider", "", "key", "Sender <sender@example.com>", true},
		{"missing key", "resend", "", "sender@example.com", false},
		{"blank key", "resend", "  ", "sender@example.com", false},
		{"missing sender", "resend", "key", "", false},
		{"invalid sender", "resend", "key", "invalid", false},
		{"noop", "noop", "key", "sender@example.com", false},
		{"unsupported provider", "brevo", "key", "sender@example.com", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := Email{Provider: tt.provider, APIKey: tt.key, FromEmail: tt.from}
			require.Equal(t, tt.want, cfg.CanSend())
		})
	}
}

func TestNewConfig_WithDefaults(t *testing.T) {
	// Clear any existing env vars that might interfere
	originalEnvs := map[string]string{
		"POSTGRES_HOST":   os.Getenv("POSTGRES_HOST"),
		"POSTGRES_DBNAME": os.Getenv("POSTGRES_DBNAME"),
	}
	defer func() {
		for k, v := range originalEnvs {
			if v != "" {
				os.Setenv(k, v)
			}
		}
	}()

	// Set minimal required env vars
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_DBNAME", "testdb")

	cfg, err := NewConfig("")
	require.NoError(t, err)

	// Verify defaults
	assert.Equal(t, "go-boilerplate", cfg.App.Name)
	assert.Equal(t, "1.0.0", cfg.App.Version)
	assert.Equal(t, "development", cfg.App.Env)
	assert.Equal(t, "8080", cfg.HTTP.Port)
	assert.Equal(t, 15*time.Second, cfg.HTTP.Timeout)
	assert.Equal(t, "debug", cfg.Log.Level)
	assert.Equal(t, DefaultPostgresPort, cfg.Postgres.Port)
	assert.Equal(t, DefaultPostgresPoolSize, cfg.Postgres.MaxPoolSize)
	assert.Equal(t, DefaultRedisPort, cfg.Redis.Port)
	assert.Equal(t, DefaultAsynqConcurrency, cfg.Asynq.Concurrency)
	assert.Equal(t, "resend", cfg.Email.Provider)
}

func TestConfig_Validate_RequiredFields(t *testing.T) {
	tests := []struct {
		name      string
		cfg       Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid development config",
			cfg: Config{
				App:      App{Env: "development"},
				Postgres: Postgres{Host: "localhost", DBName: "app"},
			},
			wantError: false,
		},
		{
			name: "missing postgres host",
			cfg: Config{
				App:      App{Env: "development"},
				Postgres: Postgres{DBName: "app"},
			},
			wantError: true,
			errorMsg:  "POSTGRES_HOST is required",
		},
		{
			name: "missing postgres dbname",
			cfg: Config{
				App:      App{Env: "development"},
				Postgres: Postgres{Host: "localhost"},
			},
			wantError: true,
			errorMsg:  "POSTGRES_DBNAME is required",
		},
		{
			name: "production without jwt secret",
			cfg: Config{
				App:      App{Env: "production"},
				Postgres: Postgres{Host: "localhost", DBName: "app"},
				JWT:      JWT{SecretKey: ""},
			},
			wantError: true,
			errorMsg:  "JWT_SECRET_KEY must be set in production",
		},
		{
			name: "production with default jwt secret",
			cfg: Config{
				App:      App{Env: "production"},
				Postgres: Postgres{Host: "localhost", DBName: "app"},
				JWT:      JWT{SecretKey: "change-me-in-production"},
			},
			wantError: true,
			errorMsg:  "JWT_SECRET_KEY must be set in production",
		},
		{
			name: "production with insecure ssl",
			cfg: Config{
				App:      App{Env: "production"},
				Postgres: Postgres{Host: "localhost", DBName: "app", SSLMode: "disable"},
				JWT:      JWT{SecretKey: "real-secret-key-here"},
			},
			wantError: true,
			errorMsg:  "POSTGRES_SSLMODE must not be 'disable' in production",
		},
		{
			name: "valid production config",
			cfg: Config{
				App:      App{Env: "production"},
				Postgres: Postgres{Host: "localhost", DBName: "app", SSLMode: "require"},
				JWT:      JWT{SecretKey: "real-secret-key-here"},
			},
			wantError: false,
		},
		{
			name: "rs256 without key paths",
			cfg: Config{
				App:      App{Env: "development"},
				Postgres: Postgres{Host: "localhost", DBName: "app"},
				JWT:      JWT{Algorithm: "rs256"},
			},
			wantError: true,
			errorMsg:  "JWT_PRIVATE_KEY_PATH is required",
		},
		{
			name: "es256 without public key",
			cfg: Config{
				App:      App{Env: "development"},
				Postgres: Postgres{Host: "localhost", DBName: "app"},
				JWT:      JWT{Algorithm: "es256", PrivateKeyPath: "/some/key.pem"},
			},
			wantError: true,
			errorMsg:  "JWT_PUBLIC_KEY_PATH is required",
		},
		{
			name: "invalid jwt algorithm",
			cfg: Config{
				App:      App{Env: "development"},
				Postgres: Postgres{Host: "localhost", DBName: "app"},
				JWT:      JWT{Algorithm: "ps256"},
			},
			wantError: true,
			errorMsg:  "JWT_ALGORITHM must be hs256, rs256, or es256",
		},
		{
			name: "valid rs256 config",
			cfg: Config{
				App:      App{Env: "development"},
				Postgres: Postgres{Host: "localhost", DBName: "app"},
				JWT:      JWT{Algorithm: "rs256", PrivateKeyPath: "/key.pem", PublicKeyPath: "/pub.pem"},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_Validate_AsynqOptions(t *testing.T) {
	tests := []struct {
		name    string
		retry   int
		timeout time.Duration
		err     string
	}{
		{name: "defaults", retry: 3, timeout: 5 * time.Minute},
		{name: "zero values"},
		{name: "one second", retry: 1, timeout: time.Second},
		{name: "negative retry", retry: -1, err: "ASYNQ_MAX_RETRY must not be negative"},
		{name: "negative timeout", timeout: -time.Second, err: "ASYNQ_JOB_TIMEOUT"},
		{name: "subsecond timeout", timeout: 200 * time.Millisecond, err: "ASYNQ_JOB_TIMEOUT"},
		{name: "fractional seconds", timeout: 1500 * time.Millisecond, err: "ASYNQ_JOB_TIMEOUT"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Postgres: Postgres{Host: "localhost", DBName: "app"},
				Asynq:    Asynq{MaxRetry: tt.retry, JobTimeout: tt.timeout},
			}
			if tt.err == "" {
				require.NoError(t, cfg.Validate())
			} else {
				require.ErrorContains(t, cfg.Validate(), tt.err)
			}
		})
	}
}

func TestPostgres_DSN(t *testing.T) {
	pg := Postgres{
		Host:     "localhost",
		Port:     5432,
		User:     "user",
		Password: "pass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	expected := "host=localhost port=5432 user=user password=pass dbname=testdb sslmode=disable"
	assert.Equal(t, expected, pg.DSN())
}

func TestPostgres_URL(t *testing.T) {
	pg := Postgres{
		Host:     "localhost",
		Port:     5432,
		User:     "user",
		Password: "pass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	expected := "postgres://user:pass@localhost:5432/testdb?sslmode=disable"
	assert.Equal(t, expected, pg.URL())
}

func TestRedis_Addr(t *testing.T) {
	r := Redis{
		Host: "localhost",
		Port: 6379,
	}

	assert.Equal(t, "localhost:6379", r.Addr())
}

func TestApp_IsProduction(t *testing.T) {
	tests := []struct {
		env      string
		expected bool
	}{
		{"production", true},
		{"development", false},
		{"staging", false},
		{"", false},
	}

	for _, tt := range tests {
		name := tt.env
		if name == "" {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			app := App{Env: tt.env}
			assert.Equal(t, tt.expected, app.IsProduction())
		})
	}
}

func TestApp_ShouldAutoMigrate(t *testing.T) {
	tests := []struct {
		env      string
		expected bool
	}{
		{"development", true},
		{"", true}, // empty defaults to development
		{"production", false},
		{"staging", false},
	}

	for _, tt := range tests {
		name := tt.env
		if name == "" {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			app := App{Env: tt.env}
			assert.Equal(t, tt.expected, app.ShouldAutoMigrate())
		})
	}
}
