package config

import (
	"net"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

type GRPC struct {
	Enabled              bool          `mapstructure:"enabled"`
	Address              string        `mapstructure:"address"`
	Insecure             bool          `mapstructure:"insecure"`
	TLSCertFile          string        `mapstructure:"tls_cert_file"`
	TLSKeyFile           string        `mapstructure:"tls_key_file"`
	RequestTimeout       time.Duration `mapstructure:"request_timeout"`
	ShutdownTimeout      time.Duration `mapstructure:"shutdown_timeout"`
	ConnectionTimeout    time.Duration `mapstructure:"connection_timeout"`
	MaxMessageSize       int           `mapstructure:"max_message_size"`
	MaxConcurrentStreams uint32        `mapstructure:"max_concurrent_streams"`
}

func (g GRPC) validationErrors(production bool) []string {
	if !g.Enabled {
		return nil
	}
	var errs []string
	_, port, err := net.SplitHostPort(g.Address)
	portNumber, portErr := strconv.Atoi(port)
	if err != nil || portErr != nil || portNumber < 1 || portNumber > 65535 {
		errs = append(errs, "GRPC_ADDRESS must be a host:port address with a valid port")
	}
	if g.Insecure {
		if production {
			errs = append(errs, "GRPC_INSECURE must be false in production")
		}
	} else if g.TLSCertFile == "" || g.TLSKeyFile == "" {
		errs = append(errs, "GRPC_TLS_CERT_FILE and GRPC_TLS_KEY_FILE are required when gRPC TLS is enabled")
	}
	if g.RequestTimeout <= 0 || g.ShutdownTimeout <= 0 || g.ConnectionTimeout <= 0 {
		errs = append(errs, "gRPC timeouts must be positive")
	}
	if g.MaxMessageSize <= 0 || g.MaxConcurrentStreams == 0 {
		errs = append(errs, "gRPC message and stream limits must be positive")
	}
	return errs
}

func setGRPCDefaults() {
	viper.SetDefault("grpc.enabled", false)
	viper.SetDefault("grpc.address", "127.0.0.1:50051")
	viper.SetDefault("grpc.insecure", false)
	viper.SetDefault("grpc.tls_cert_file", "")
	viper.SetDefault("grpc.tls_key_file", "")
	viper.SetDefault("grpc.request_timeout", "5s")
	viper.SetDefault("grpc.shutdown_timeout", "15s")
	viper.SetDefault("grpc.connection_timeout", "5s")
	viper.SetDefault("grpc.max_message_size", 65536)
	viper.SetDefault("grpc.max_concurrent_streams", 128)
}
