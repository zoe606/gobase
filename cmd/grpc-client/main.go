package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc/metadata"

	echov1 "go-boilerplate/pkg/api/echo/v1"
	"go-boilerplate/pkg/grpcclient"
)

func main() {
	address := flag.String("address", "127.0.0.1:50051", "gRPC address")
	plaintext := flag.Bool("insecure", false, "explicit plaintext development transport")
	ca := flag.String("ca-file", "", "private CA PEM file")
	message := flag.String("message", "hello", "Echo message")
	timeout := flag.Duration("timeout", 5*time.Second, "RPC timeout")
	flag.Parse()
	if err := run(grpcclient.Options{Address: *address, Insecure: *plaintext, CAFile: *ca}, *message, *timeout, os.Getenv("GRPC_ACCESS_TOKEN")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(options grpcclient.Options, message string, timeout time.Duration, token string) error {
	if token == "" {
		return fmt.Errorf("set GRPC_ACCESS_TOKEN to an access token")
	}
	if timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	conn, err := grpcclient.New(options)
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	response, err := echov1.NewEchoServiceClient(conn).Echo(ctx, &echov1.EchoRequest{Message: message})
	if err != nil {
		return err
	}
	fmt.Printf("message=%s request_id=%s\n", response.GetMessage(), response.GetRequestId())
	return nil
}
