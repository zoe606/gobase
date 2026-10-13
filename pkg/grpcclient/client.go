package grpcclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Options struct {
	Address  string
	Insecure bool
	CAFile   string
}

func New(options Options) (*grpc.ClientConn, error) {
	var transport credentials.TransportCredentials
	if options.Insecure {
		if options.CAFile != "" {
			return nil, fmt.Errorf("CA file cannot be used with insecure transport")
		}
		transport = insecure.NewCredentials()
	} else {
		cfg, err := tlsConfig(options.CAFile)
		if err != nil {
			return nil, err
		}
		transport = credentials.NewTLS(cfg)
	}
	return grpc.NewClient(options.Address, grpc.WithTransportCredentials(transport))
}

func tlsConfig(caFile string) (*tls.Config, error) {
	cfg := &tls.Config{MinVersion: tls.VersionTLS13}
	if caFile == "" {
		return cfg, nil
	}
	pem, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("gRPC CA file: %w", err)
	}
	cfg.RootCAs = x509.NewCertPool()
	if !cfg.RootCAs.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("gRPC CA file has no valid certificates")
	}
	return cfg, nil
}
