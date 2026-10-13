package config

import (
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type WebSocket struct {
	Enabled          bool          `mapstructure:"enabled"`
	AllowedOrigins   string        `mapstructure:"allowed_origins"`
	HandshakeTimeout time.Duration `mapstructure:"handshake_timeout"`
	ReadTimeout      time.Duration `mapstructure:"read_timeout"`
	WriteTimeout     time.Duration `mapstructure:"write_timeout"`
	PingInterval     time.Duration `mapstructure:"ping_interval"`
	MaxMessageSize   int64         `mapstructure:"max_message_size"`
	MaxConnections   int           `mapstructure:"max_connections"`
}

func (w WebSocket) Origins() []string {
	var origins []string
	for _, origin := range strings.Split(w.AllowedOrigins, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

func (w WebSocket) AllowsOrigin(origin string) bool {
	for _, allowed := range w.Origins() {
		if origin == allowed {
			return true
		}
	}
	return false
}

func (w WebSocket) validationErrors() []string {
	if !w.Enabled {
		return nil
	}
	var errs []string
	if len(w.Origins()) == 0 {
		errs = append(errs, "WEBSOCKET_ALLOWED_ORIGINS must list explicit origins when WebSocket is enabled")
	}
	for _, origin := range w.Origins() {
		if !validWebSocketOrigin(origin) {
			errs = append(errs, "WEBSOCKET_ALLOWED_ORIGINS must contain HTTP or HTTPS origins without paths or wildcards")
		}
	}
	if w.HandshakeTimeout <= 0 || w.ReadTimeout <= 0 || w.WriteTimeout <= 0 || w.PingInterval <= 0 || w.PingInterval >= w.ReadTimeout {
		errs = append(errs, "WebSocket timeouts must be positive and ping_interval must be less than read_timeout")
	}
	if w.MaxMessageSize <= 0 || w.MaxConnections <= 0 {
		errs = append(errs, "WebSocket message and connection limits must be positive")
	}
	return errs
}

func validWebSocketOrigin(origin string) bool {
	u, err := url.Parse(origin)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" && !strings.Contains(u.Host, "*") && u.User == nil && u.Path == "" && u.RawQuery == "" && u.Fragment == ""
}

func setWebSocketDefaults() {
	viper.SetDefault("websocket.enabled", false)
	viper.SetDefault("websocket.allowed_origins", "")
	viper.SetDefault("websocket.handshake_timeout", "5s")
	viper.SetDefault("websocket.read_timeout", "60s")
	viper.SetDefault("websocket.write_timeout", "10s")
	viper.SetDefault("websocket.ping_interval", "25s")
	viper.SetDefault("websocket.max_message_size", 65536)
	viper.SetDefault("websocket.max_connections", 128)
}
