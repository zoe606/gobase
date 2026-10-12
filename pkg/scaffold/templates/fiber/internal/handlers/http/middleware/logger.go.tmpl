package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"go-boilerplate/config"
	pkgjson "go-boilerplate/pkg/json"
)

// ParseRedactFields splits a comma-separated string into a slice of trimmed, non-empty field names.
func ParseRedactFields(s string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))

	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

// RedactJSON replaces values of specified top-level keys with "[REDACTED]" in a JSON string.
// Returns the input as-is if it is not valid JSON or no fields are given.
func RedactJSON(body string, fields []string) string {
	if body == "" || len(fields) == 0 {
		return body
	}

	var parsed map[string]interface{}
	if err := pkgjson.Unmarshal([]byte(body), &parsed); err != nil {
		return body
	}

	redactSet := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		redactSet[strings.ToLower(f)] = struct{}{}
	}

	for key := range parsed {
		if _, ok := redactSet[strings.ToLower(key)]; ok {
			parsed[key] = "[REDACTED]"
		}
	}

	out, err := pkgjson.Marshal(parsed)
	if err != nil {
		return body
	}

	return string(out)
}

// StructuredLogger returns a Fiber middleware that logs HTTP requests as structured zap fields.
// Log level is determined by response status code: 2xx/3xx->info, 4xx->warn, 5xx->error.
func StructuredLogger(zapLogger *zap.Logger, cfg config.Log) fiber.Handler {
	redactFields := ParseRedactFields(cfg.RedactFields)

	return func(ctx *fiber.Ctx) error {
		start := time.Now()

		// Capture request body before handler if enabled
		var reqBody string
		if cfg.LogRequestBody {
			reqBody = string(ctx.Body())
		}

		// Process request
		err := ctx.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get request ID from requestid middleware
		requestID := ctx.GetRespHeader("X-Request-Id")
		if requestID == "" {
			if id, ok := ctx.Locals("requestid").(string); ok {
				requestID = id
			}
		}

		// Build structured fields
		status := ctx.Response().StatusCode()
		fields := []zap.Field{
			zap.String("method", ctx.Method()),
			zap.String("path", ctx.OriginalURL()),
			zap.Int("status", status),
			zap.Float64("latency_ms", float64(latency.Nanoseconds())/1e6),
			zap.String("request_id", requestID),
			zap.String("ip", ctx.IP()),
			zap.String("user_agent", ctx.Get("User-Agent")),
			zap.Int("bytes_out", len(ctx.Response().Body())),
		}

		// Add user_id if present in context
		if userID, ok := ctx.Locals(UserIDKey).(uint); ok && userID > 0 {
			fields = append(fields, zap.Uint64("user_id", uint64(userID)))
		}

		// Add request body if enabled (with redaction)
		if cfg.LogRequestBody && reqBody != "" {
			redacted := RedactJSON(reqBody, redactFields)
			fields = append(fields, zap.String("request_body", redacted))
		}

		// Add response body if enabled
		if cfg.LogResponseBody {
			fields = append(fields, zap.String("response_body", string(ctx.Response().Body())))
		}

		// Log at appropriate level based on status code
		switch {
		case status >= 500:
			zapLogger.Error("HTTP request", fields...)
		case status >= 400:
			zapLogger.Warn("HTTP request", fields...)
		default:
			zapLogger.Info("HTTP request", fields...)
		}

		return err
	}
}
