package grpcserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"

	"go-boilerplate/config"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

type Server struct {
	GRPC         *grpc.Server
	cfg          config.GRPC
	l            logger.Interface
	auth         jwt.Service
	health       *health.Server
	mu           sync.Mutex
	listener     net.Listener
	stopping     bool
	notify       chan error
	shutdownOnce sync.Once
	forceOnce    sync.Once
	shutdownDone chan struct{}
}

func New(cfg config.GRPC, auth jwt.Service, l logger.Interface) (*Server, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("enable GRPC_ENABLED before starting cmd/grpc")
	}
	s := &Server{cfg: cfg, auth: auth, l: l, health: health.NewServer(), notify: make(chan error, 1), shutdownDone: make(chan struct{})}
	options := []grpc.ServerOption{grpc.ChainUnaryInterceptor(s.unary), grpc.MaxRecvMsgSize(cfg.MaxMessageSize), grpc.MaxSendMsgSize(cfg.MaxMessageSize), grpc.MaxConcurrentStreams(cfg.MaxConcurrentStreams), grpc.ConnectionTimeout(cfg.ConnectionTimeout)}
	if !cfg.Insecure {
		certificate, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			return nil, fmt.Errorf("gRPC TLS certificate: %w", err)
		}
		options = append(options, grpc.Creds(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS13, Certificates: []tls.Certificate{certificate}})))
	}
	s.GRPC = grpc.NewServer(options...)
	healthv1.RegisterHealthServer(s.GRPC, s.health)
	s.health.SetServingStatus("", healthv1.HealthCheckResponse_NOT_SERVING)
	return s, nil
}

func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener != nil || s.stopping {
		return fmt.Errorf("gRPC server already started or stopping")
	}
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", s.cfg.Address)
	if err != nil {
		return fmt.Errorf("gRPC listen: %w", err)
	}
	s.listener = listener
	s.health.SetServingStatus("", healthv1.HealthCheckResponse_SERVING)
	for name := range s.GRPC.GetServiceInfo() {
		s.health.SetServingStatus(name, healthv1.HealthCheckResponse_SERVING)
	}
	go func() {
		if err := s.GRPC.Serve(listener); err != nil {
			s.notify <- err
		}
		close(s.notify)
	}()
	s.l.Info("gRPC server started address=%s", listener.Addr().String())
	return nil
}

func (s *Server) Address() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.cfg.Address
}

func (s *Server) Notify() <-chan error { return s.notify }
