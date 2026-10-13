package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go-boilerplate/config"
	echohandler "go-boilerplate/internal/handlers/grpc/echo"
	echousecase "go-boilerplate/internal/usecase/echo"
	echov1 "go-boilerplate/pkg/api/echo/v1"
	"go-boilerplate/pkg/grpcserver"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

func main() {
	path := flag.String("config", "", "path to config file")
	flag.Parse()
	if err := run(*path); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(path string) error {
	cfg, err := config.NewConfig(path)
	if err != nil {
		return err
	}
	auth, err := jwt.NewFromConfig(cfg.JWT)
	if err != nil {
		return err
	}
	l := logger.New(cfg.Log.Level)
	defer func() { _ = l.Sync() }()
	server, err := grpcserver.New(cfg.GRPC, auth, l)
	if err != nil {
		return err
	}
	echov1.RegisterEchoServiceServer(server.GRPC, echohandler.New(echousecase.New()))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := server.Start(); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
	case err = <-server.Notify():
	}
	shutdown, cancel := context.WithTimeout(context.Background(), cfg.GRPC.ShutdownTimeout)
	defer cancel()
	if shutdownErr := server.Shutdown(shutdown); shutdownErr != nil {
		if !errors.Is(shutdownErr, context.DeadlineExceeded) {
			return shutdownErr
		}
		l.Warn("gRPC forced shutdown after timeout")
	}
	return err
}
