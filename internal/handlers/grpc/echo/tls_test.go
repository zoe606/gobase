package echohandler_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	echousecase "go-boilerplate/internal/usecase/echo"
	echov1 "go-boilerplate/pkg/api/echo/v1"
	"go-boilerplate/pkg/grpcclient"
	"go-boilerplate/pkg/grpcserver"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

func certificate(t *testing.T) (certFile, keyFile string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	cert := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "RPC fixture"}, NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, BasicConstraintsValid: true, IsCA: true, IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}}
	der, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	require.NoError(t, err)
	private, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)
	dir := t.TempDir()
	certFile, keyFile = filepath.Join(dir, "server.pem"), filepath.Join(dir, "server.key")
	require.NoError(t, os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600))
	require.NoError(t, os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: private}), 0o600))
	return certFile, keyFile
}

func TestGRPCTLS(t *testing.T) {
	cfg := rpcConfig()
	cfg.Insecure = false
	cfg.TLSCertFile, cfg.TLSKeyFile = certificate(t)
	svc := jwt.New("rpc-tls", time.Hour, time.Hour)
	server, conn := fixture(t, cfg, svc, echousecase.New())
	response, err := echov1.NewEchoServiceClient(conn).Echo(authContext(t, svc), &echov1.EchoRequest{Message: "TLS"})
	require.NoError(t, err)
	require.Equal(t, "TLS", response.GetMessage())
	for _, test := range []struct {
		name    string
		options grpcclient.Options
	}{
		{"plaintext", grpcclient.Options{Address: server.Address(), Insecure: true}},
		{"untrusted CA", grpcclient.Options{Address: server.Address()}},
	} {
		t.Run(test.name, func(t *testing.T) {
			client, err := grpcclient.New(test.options)
			require.NoError(t, err)
			t.Cleanup(func() { _ = client.Close() })
			ctx, cancel := context.WithTimeout(authContext(t, svc), time.Second)
			defer cancel()
			_, err = echov1.NewEchoServiceClient(client).Echo(ctx, &echov1.EchoRequest{Message: "untrusted"})
			require.Equal(t, codes.Unavailable, status.Code(err), "transport rejection: %v", err)
		})
	}
	t.Run("wrong hostname", func(t *testing.T) {
		transport, err := credentials.NewClientTLSFromFile(cfg.TLSCertFile, "")
		require.NoError(t, err)
		// Keep the TCP address fixed so this checks certificate identity without DNS or IPv6 retries.
		client, err := grpc.NewClient(server.Address(), grpc.WithTransportCredentials(transport), grpc.WithAuthority("localhost"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = client.Close() })
		ctx, cancel := context.WithTimeout(authContext(t, svc), time.Second)
		defer cancel()
		_, err = echov1.NewEchoServiceClient(client).Echo(ctx, &echov1.EchoRequest{Message: "wrong hostname"})
		require.Equal(t, codes.Unavailable, status.Code(err), "transport rejection: %v", err)
		require.ErrorContains(t, err, "certificate")
		require.ErrorContains(t, err, "localhost")
	})
}

func TestGRPCTransportConfiguration(t *testing.T) {
	cfg := rpcConfig()
	cfg.Enabled = false
	_, err := grpcserver.New(cfg, nil, logger.New("error"))
	require.ErrorContains(t, err, "GRPC_ENABLED")
	cfg.Enabled, cfg.Insecure = true, false
	_, err = grpcserver.New(cfg, nil, logger.New("error"))
	require.ErrorContains(t, err, "TLS certificate")
	for _, options := range []grpcclient.Options{
		{Insecure: true, CAFile: "bad"}, {CAFile: "missing"},
	} {
		_, err := grpcclient.New(options)
		require.Error(t, err)
	}
	invalid := filepath.Join(t.TempDir(), "invalid.pem")
	require.NoError(t, os.WriteFile(invalid, []byte("invalid"), 0o600))
	_, err = grpcclient.New(grpcclient.Options{CAFile: invalid})
	require.ErrorContains(t, err, "no valid certificates")
}
