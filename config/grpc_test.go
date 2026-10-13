package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGRPCConfig(t *testing.T) {
	valid := GRPC{Enabled: true, Address: "127.0.0.1:50051", Insecure: true, RequestTimeout: time.Second, ShutdownTimeout: time.Second, ConnectionTimeout: time.Second, MaxMessageSize: 65536, MaxConcurrentStreams: 128}
	require.Empty(t, valid.validationErrors(false))
	require.Empty(t, (GRPC{}).validationErrors(true))
	require.NotEmpty(t, valid.validationErrors(true))
	tls := valid
	tls.Insecure = false
	require.NotEmpty(t, tls.validationErrors(false))
	tls.TLSCertFile, tls.TLSKeyFile = "cert.pem", "key.pem"
	require.Empty(t, tls.validationErrors(true))
	for _, address := range []string{"localhost", "localhost:0", "localhost:65536", "localhost:abc"} {
		cfg := valid
		cfg.Address = address
		require.NotEmpty(t, cfg.validationErrors(false))
	}
	valid.RequestTimeout = 0
	valid.MaxMessageSize = 0
	require.Len(t, valid.validationErrors(false), 2)
}

func TestGRPCEnvironment(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "localhost")
	t.Setenv("POSTGRES_DBNAME", "testdb")
	t.Setenv("GRPC_ENABLED", "false")
	cfg, err := NewConfig("")
	require.NoError(t, err)
	require.False(t, cfg.GRPC.Enabled)
	t.Setenv("GRPC_ENABLED", "true")
	t.Setenv("GRPC_INSECURE", "true")
	t.Setenv("GRPC_ADDRESS", "127.0.0.1:50052")
	cfg, err = NewConfig("")
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1:50052", cfg.GRPC.Address)
	require.Equal(t, uint32(128), cfg.GRPC.MaxConcurrentStreams)
	t.Setenv("GRPC_MAX_CONCURRENT_STREAMS", "-1")
	_, err = NewConfig("")
	require.Error(t, err)
	t.Setenv("GRPC_MAX_CONCURRENT_STREAMS", "128")
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET_KEY", "grpc-config-test")
	t.Setenv("POSTGRES_SSLMODE", "require")
	_, err = NewConfig("")
	require.ErrorContains(t, err, "GRPC_INSECURE")
}
