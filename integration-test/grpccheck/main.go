package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	echov1 "go-boilerplate/pkg/api/echo/v1"
	"go-boilerplate/pkg/grpcclient"
)

func main() {
	address := flag.String("address", "127.0.0.1:50051", "gRPC address")
	base := flag.String("http-url", "http://localhost:8080", "HTTP app for obtaining an access token")
	plaintext := flag.Bool("insecure", false, "explicit plaintext development transport")
	ca := flag.String("ca-file", "", "private CA file")
	flag.Parse()
	if err := run(grpcclient.Options{Address: *address, Insecure: *plaintext, CAFile: *ca}, *base); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(options grpcclient.Options, base string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := grpcclient.New(options)
	if err != nil {
		return err
	}
	defer conn.Close()
	health, err := healthv1.NewHealthClient(conn).Check(ctx, &healthv1.HealthCheckRequest{}, grpc.WaitForReady(true))
	if err != nil {
		return err
	}
	if health.GetStatus() != healthv1.HealthCheckResponse_SERVING {
		return fmt.Errorf("gRPC health is not serving")
	}
	token, err := register(ctx, base)
	if err != nil {
		return err
	}
	client := echov1.NewEchoServiceClient(conn)
	if _, err := client.Echo(ctx, &echov1.EchoRequest{Message: "unauthorized"}); status.Code(err) != codes.Unauthenticated {
		return fmt.Errorf("missing token status: %w", err)
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token, "x-request-id", "grpc-runtime")
	response, err := client.Echo(ctx, &echov1.EchoRequest{Message: "runtime"})
	if err != nil {
		return err
	}
	if response.GetMessage() != "runtime" || response.GetRequestId() != "grpc-runtime" {
		return fmt.Errorf("invalid gRPC echo response")
	}
	if _, err := client.Echo(ctx, &echov1.EchoRequest{}); status.Code(err) != codes.InvalidArgument {
		return fmt.Errorf("invalid input status: %w", err)
	}
	fmt.Println("gRPC authenticated echo, validation, and health verified")
	return nil
}

func register(ctx context.Context, base string) (string, error) {
	body := fmt.Sprintf(`{"email":"grpc-%d@example.com","password":"RuntimeTest123!","name":"gRPC runtime"}`, time.Now().UnixNano())
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/v1/auth/register", bytes.NewBufferString(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{Timeout: 5 * time.Second}).Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("registration status %d", response.StatusCode)
	}
	var result struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Data.AccessToken, nil
}
