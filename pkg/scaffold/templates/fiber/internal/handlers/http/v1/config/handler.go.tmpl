// Package config provides HTTP handlers for application configuration endpoints.
package config

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"go-boilerplate/config"
	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
	"go-boilerplate/pkg/response"
)

// ConfigResponse represents the sanitized config returned to clients.
// Sensitive fields are masked or omitted.
type ConfigResponse struct {
	App       AppConfig       `json:"app"`
	HTTP      HTTPConfig      `json:"http"`
	Log       LogConfig       `json:"log"`
	Postgres  PostgresConfig  `json:"postgres"`
	Redis     RedisConfig     `json:"redis"`
	JWT       JWTConfig       `json:"jwt"`
	CORS      CORSConfig      `json:"cors"`
	RateLimit RateLimitConfig `json:"rate_limit"`
	Asynq     AsynqConfig     `json:"asynq"`
	Email     EmailConfig     `json:"email"`
	Metrics   MetricsConfig   `json:"metrics"`
	Swagger   SwaggerConfig   `json:"swagger"`
	CachedAt  time.Time       `json:"cached_at"`
}

// AppConfig holds app configuration (safe to expose).
type AppConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Env     string `json:"env"`
}

// HTTPConfig holds HTTP configuration.
type HTTPConfig struct {
	Port           string `json:"port"`
	Timeout        string `json:"timeout"`
	IdleTimeout    string `json:"idle_timeout"`
	RequestTimeout string `json:"request_timeout"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string `json:"level"`
	File  string `json:"file"`
}

// PostgresConfig holds database configuration (passwords masked).
type PostgresConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	User            string `json:"user"`
	Password        string `json:"password"` // Masked
	DBName          string `json:"dbname"`
	SSLMode         string `json:"sslmode"`
	MaxPoolSize     int    `json:"max_pool_size"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	ConnMaxLifetime string `json:"conn_max_lifetime"`
	ConnMaxIdleTime string `json:"conn_max_idle_time"`
}

// RedisConfig holds Redis configuration (passwords masked).
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"` // Masked
	DB       int    `json:"db"`
}

// JWTConfig holds JWT configuration (secrets masked).
type JWTConfig struct {
	SecretKey     string `json:"secret_key"` // Masked
	AccessExpiry  string `json:"access_expiry"`
	RefreshExpiry string `json:"refresh_expiry"`
}

// CORSConfig holds CORS configuration.
type CORSConfig struct {
	AllowOrigins     string `json:"allow_origins"`
	AllowMethods     string `json:"allow_methods"`
	AllowHeaders     string `json:"allow_headers"`
	AllowCredentials bool   `json:"allow_credentials"`
	MaxAge           int    `json:"max_age"`
}

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	Max        int    `json:"max"`
	Expiration string `json:"expiration"`
}

// AsynqConfig holds Asynq configuration.
type AsynqConfig struct {
	Concurrency int    `json:"concurrency"`
	JobTimeout  string `json:"job_timeout"`
}

// EmailConfig holds email configuration (API keys masked).
type EmailConfig struct {
	Provider  string `json:"provider"`
	APIKey    string `json:"api_key"` // Masked
	FromEmail string `json:"from_email"`
	FromName  string `json:"from_name"`
}

// MetricsConfig holds metrics configuration.
type MetricsConfig struct {
	Enabled bool `json:"enabled"`
}

// SwaggerConfig holds Swagger configuration.
type SwaggerConfig struct {
	Enabled bool `json:"enabled"`
}

// Handler handles config-related HTTP requests.
type Handler struct {
	cfg        *config.Config
	jwtService jwt.Service
	l          logger.Interface

	// Cache
	cache    *ConfigResponse
	cacheMu  sync.RWMutex
	cacheTTL time.Duration
	cachedAt time.Time
}

// New creates a new config handler.
func New(cfg *config.Config, jwtService jwt.Service, l logger.Interface) *Handler {
	return &Handler{
		cfg:        cfg,
		jwtService: jwtService,
		l:          l,
		cacheTTL:   5 * time.Minute, // Cache config for 5 minutes
	}
}

// RegisterRoutes registers config routes.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	cfg := router.Group("/config")
	{
		// All config routes require admin or superadmin role
		cfg.Use(middleware.JWTAuth(h.jwtService, h.l))
		cfg.Use(middleware.RequireRole("admin", "superadmin"))

		cfg.Get("", h.GetConfig)
		cfg.Post("/cache/invalidate", h.InvalidateCache)
	}
}

// GetConfig returns the current application configuration.
// @Summary     Get application configuration
// @Description Returns sanitized application configuration (admin only)
// @Tags        config
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} response.Response[ConfigResponse]
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Router      /config [get]
func (h *Handler) GetConfig(c *fiber.Ctx) error {
	// Try to get from cache
	h.cacheMu.RLock()
	if h.cache != nil && time.Since(h.cachedAt) < h.cacheTTL {
		cached := h.cache
		h.cacheMu.RUnlock()
		h.l.Debug("Serving config from cache")
		return response.OK(c, cached)
	}
	h.cacheMu.RUnlock()

	// Build and cache response
	h.cacheMu.Lock()
	defer h.cacheMu.Unlock()

	// Double-check after acquiring write lock
	if h.cache != nil && time.Since(h.cachedAt) < h.cacheTTL {
		return response.OK(c, h.cache)
	}

	h.l.Debug("Building config response and caching")
	configResp := h.buildConfigResponse()
	h.cache = configResp
	h.cachedAt = time.Now()

	return response.OK(c, configResp)
}

// InvalidateCache clears the config cache.
// @Summary     Invalidate config cache
// @Description Clears the cached configuration (admin only)
// @Tags        config
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} response.Response[map[string]string]
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Router      /config/cache/invalidate [post]
func (h *Handler) InvalidateCache(c *fiber.Ctx) error {
	h.cacheMu.Lock()
	h.cache = nil
	h.cachedAt = time.Time{}
	h.cacheMu.Unlock()

	h.l.Info("Config cache invalidated by user",
		"user_id", middleware.GetUserID(c),
		"email", middleware.GetEmail(c),
	)

	return response.OK(c, map[string]string{
		"message": "Config cache invalidated successfully",
	})
}

// buildConfigResponse creates a sanitized config response.
func (h *Handler) buildConfigResponse() *ConfigResponse {
	return &ConfigResponse{
		App: AppConfig{
			Name:    h.cfg.App.Name,
			Version: h.cfg.App.Version,
			Env:     h.cfg.App.Env,
		},
		HTTP: HTTPConfig{
			Port:           h.cfg.HTTP.Port,
			Timeout:        h.cfg.HTTP.Timeout.String(),
			IdleTimeout:    h.cfg.HTTP.IdleTimeout.String(),
			RequestTimeout: h.cfg.HTTP.RequestTimeout.String(),
		},
		Log: LogConfig{
			Level: h.cfg.Log.Level,
			File:  h.cfg.Log.File,
		},
		Postgres: PostgresConfig{
			Host:            h.cfg.Postgres.Host,
			Port:            h.cfg.Postgres.Port,
			User:            h.cfg.Postgres.User,
			Password:        maskSecret(h.cfg.Postgres.Password),
			DBName:          h.cfg.Postgres.DBName,
			SSLMode:         h.cfg.Postgres.SSLMode,
			MaxPoolSize:     h.cfg.Postgres.MaxPoolSize,
			MaxIdleConns:    h.cfg.Postgres.MaxIdleConns,
			ConnMaxLifetime: h.cfg.Postgres.ConnMaxLifetime.String(),
			ConnMaxIdleTime: h.cfg.Postgres.ConnMaxIdleTime.String(),
		},
		Redis: RedisConfig{
			Host:     h.cfg.Redis.Host,
			Port:     h.cfg.Redis.Port,
			Password: maskSecret(h.cfg.Redis.Password),
			DB:       h.cfg.Redis.DB,
		},
		JWT: JWTConfig{
			SecretKey:     maskSecret(h.cfg.JWT.SecretKey),
			AccessExpiry:  h.cfg.JWT.AccessExpiry.String(),
			RefreshExpiry: h.cfg.JWT.RefreshExpiry.String(),
		},
		CORS: CORSConfig{
			AllowOrigins:     h.cfg.CORS.AllowOrigins,
			AllowMethods:     h.cfg.CORS.AllowMethods,
			AllowHeaders:     h.cfg.CORS.AllowHeaders,
			AllowCredentials: h.cfg.CORS.AllowCredentials,
			MaxAge:           h.cfg.CORS.MaxAge,
		},
		RateLimit: RateLimitConfig{
			Max:        h.cfg.RateLimit.Max,
			Expiration: h.cfg.RateLimit.Expiration.String(),
		},
		Asynq: AsynqConfig{
			Concurrency: h.cfg.Asynq.Concurrency,
			JobTimeout:  h.cfg.Asynq.JobTimeout.String(),
		},
		Email: EmailConfig{
			Provider:  h.cfg.Email.Provider,
			APIKey:    maskSecret(h.cfg.Email.APIKey),
			FromEmail: h.cfg.Email.FromEmail,
			FromName:  h.cfg.Email.FromName,
		},
		Metrics: MetricsConfig{
			Enabled: h.cfg.Metrics.Enabled,
		},
		Swagger: SwaggerConfig{
			Enabled: h.cfg.Swagger.Enabled,
		},
		CachedAt: time.Now(),
	}
}

// maskSecret masks a secret string for safe display.
// Shows first 4 chars if length > 8, otherwise shows "****".
func maskSecret(s string) string {
	if s == "" {
		return "(not set)"
	}
	if len(s) > 8 {
		return s[:4] + "****"
	}
	return "****"
}
