package echohandler_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go-boilerplate/config"
	echohandler "go-boilerplate/internal/handlers/grpc/echo"
	echousecase "go-boilerplate/internal/usecase/echo"
	echov1 "go-boilerplate/pkg/api/echo/v1"
	"go-boilerplate/pkg/grpcclient"
	"go-boilerplate/pkg/grpcserver"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

type echoFunc func(context.Context, string) (string, error)

func (f echoFunc) Echo(ctx context.Context, message string) (string, error) { return f(ctx, message) }

func rpcConfig() config.GRPC {
	return config.GRPC{Enabled: true, Address: "127.0.0.1:0", Insecure: true, RequestTimeout: time.Second, ShutdownTimeout: time.Second, ConnectionTimeout: time.Second, MaxMessageSize: 65536, MaxConcurrentStreams: 128}
}

func fixture(t *testing.T, cfg config.GRPC, auth jwt.Service, usecase echohandler.UseCase) (*grpcserver.Server, *grpc.ClientConn) {
	t.Helper()
	server, err := grpcserver.New(cfg, auth, logger.New("error"))
	require.NoError(t, err)
	echov1.RegisterEchoServiceServer(server.GRPC, echohandler.New(usecase))
	require.NoError(t, server.Start())
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		require.NoError(t, server.Shutdown(ctx))
	})
	conn, err := grpcclient.New(grpcclient.Options{Address: server.Address(), Insecure: cfg.Insecure, CAFile: cfg.TLSCertFile})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return server, conn
}

func token(t *testing.T, svc jwt.Service) string {
	t.Helper()
	value, _, err := svc.GenerateAccessToken(7, "rpc@example.com", "user", nil)
	require.NoError(t, err)
	return value
}

func authContext(t *testing.T, svc jwt.Service) context.Context {
	t.Helper()
	return metadata.AppendToOutgoingContext(t.Context(), "authorization", "Bearer "+token(t, svc), "x-request-id", "grpc-contract")
}

func TestGRPCContract(t *testing.T) {
	svc := jwt.New("rpc-test", time.Hour, time.Hour)
	server, conn := fixture(t, rpcConfig(), svc, echousecase.New())
	client := echov1.NewEchoServiceClient(conn)
	ctx := authContext(t, svc)
	var headers metadata.MD
	response, err := client.Echo(ctx, &echov1.EchoRequest{Message: "hello 世界"}, grpc.Header(&headers))
	require.NoError(t, err)
	require.Equal(t, "hello 世界", response.GetMessage())
	require.Equal(t, "grpc-contract", response.GetRequestId())
	require.Equal(t, []string{"grpc-contract"}, headers.Get("x-request-id"))
	for _, message := range []string{"", " \n", strings.Repeat("a", 4097)} {
		_, err := client.Echo(ctx, &echov1.EchoRequest{Message: message})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	}
	_, err = client.Echo(ctx, &echov1.EchoRequest{Message: strings.Repeat("a", 70000)})
	require.Equal(t, codes.ResourceExhausted, status.Code(err))
	health := healthv1.NewHealthClient(conn)
	for _, name := range []string{"", echov1.EchoService_ServiceDesc.ServiceName} {
		result, err := health.Check(t.Context(), &healthv1.HealthCheckRequest{Service: name})
		require.NoError(t, err)
		require.Equal(t, healthv1.HealthCheckResponse_SERVING, result.GetStatus())
	}
	_, err = health.Check(t.Context(), &healthv1.HealthCheckRequest{Service: "missing"})
	require.Equal(t, codes.NotFound, status.Code(err))
	err = conn.Invoke(ctx, "/gobase.echo.v1.EchoService/Missing", &echov1.EchoRequest{}, &echov1.EchoResponse{})
	require.Equal(t, codes.Unimplemented, status.Code(err))
	require.Error(t, server.Start())
}

func TestGRPCCredentials(t *testing.T) {
	svc := jwt.New("rpc-test", time.Hour, time.Hour)
	_, conn := fixture(t, rpcConfig(), svc, echousecase.New())
	refresh, _, err := svc.GenerateRefreshToken()
	require.NoError(t, err)
	values := []string{"", "Basic invalid", "Bearer invalid", "Bearer " + refresh, "Bearer " + token(t, jwt.New("rpc-test", -time.Hour, time.Hour)), "Bearer " + token(t, jwt.New("other-key", time.Hour, time.Hour))}
	for _, value := range values {
		ctx := metadata.AppendToOutgoingContext(t.Context(), "authorization", value)
		_, err := echov1.NewEchoServiceClient(conn).Echo(ctx, &echov1.EchoRequest{Message: "hello"})
		require.Equal(t, codes.Unauthenticated, status.Code(err))
	}
}

func TestGRPCDeadlinesAndCancellation(t *testing.T) {
	for _, name := range []string{"client deadline", "server timeout", "cancellation", "token expiry"} {
		t.Run(name, func(t *testing.T) {
			cfg := rpcConfig()
			svc := jwt.New("rpc-test", time.Hour, time.Hour)
			if name == "server timeout" {
				cfg.RequestTimeout = 40 * time.Millisecond
			}
			if name == "token expiry" {
				svc = jwt.New("rpc-test", 2*time.Second, time.Hour)
				cfg.RequestTimeout = 5 * time.Second
			}
			entered, stopped := make(chan struct{}), make(chan struct{})
			_, conn := fixture(t, cfg, svc, echoFunc(func(ctx context.Context, _ string) (string, error) {
				close(entered)
				<-ctx.Done()
				close(stopped)
				return "", ctx.Err()
			}))
			budget := 3 * time.Second
			if name == "client deadline" {
				budget = 40 * time.Millisecond
			}
			ctx, cancel := context.WithTimeout(authContext(t, svc), budget)
			defer cancel()
			result := make(chan error, 1)
			go func() {
				_, err := echov1.NewEchoServiceClient(conn).Echo(ctx, &echov1.EchoRequest{Message: "hello"})
				result <- err
			}()
			select {
			case <-entered:
			case <-time.After(time.Second):
				t.Fatal("handler did not start")
			}
			if name == "cancellation" {
				cancel()
			}
			select {
			case err := <-result:
				code := codes.DeadlineExceeded
				if name == "cancellation" {
					code = codes.Canceled
				}
				require.Equal(t, code, status.Code(err))
			case <-time.After(4 * time.Second):
				t.Fatal("RPC did not finish")
			}
			select {
			case <-stopped:
			case <-time.After(time.Second):
				t.Fatal("handler context did not stop")
			}
		})
	}
}

func TestGRPCRecovery(t *testing.T) {
	for _, usecase := range []echohandler.UseCase{
		echoFunc(func(context.Context, string) (string, error) { panic("private panic value") }),
		echoFunc(func(context.Context, string) (string, error) { return "", errors.New("private error value") }),
	} {
		svc := jwt.New("rpc-test", time.Hour, time.Hour)
		_, conn := fixture(t, rpcConfig(), svc, usecase)
		_, err := echov1.NewEchoServiceClient(conn).Echo(authContext(t, svc), &echov1.EchoRequest{Message: "hello"})
		require.Equal(t, codes.Internal, status.Code(err))
		require.Equal(t, "internal error", status.Convert(err).Message())
	}
}

func TestGRPCShutdown(t *testing.T) {
	for _, force := range []bool{false, true} {
		t.Run(map[bool]string{false: "graceful", true: "fallback"}[force], func(t *testing.T) {
			svc := jwt.New("rpc-test", time.Hour, time.Hour)
			entered, release := make(chan struct{}), make(chan struct{})
			var once sync.Once
			finish := func() { once.Do(func() { close(release) }) }
			defer finish()
			server, conn := fixture(t, rpcConfig(), svc, echoFunc(func(context.Context, string) (string, error) { close(entered); <-release; return "finished", nil }))
			result := make(chan error, 1)
			go func() {
				_, err := echov1.NewEchoServiceClient(conn).Echo(authContext(t, svc), &echov1.EchoRequest{Message: "hello"})
				result <- err
			}()
			select {
			case <-entered:
			case <-time.After(time.Second):
				t.Fatal("handler did not start")
			}
			ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
			defer cancel()
			shutdown := make(chan error, 1)
			go func() { shutdown <- server.Shutdown(ctx) }()
			if !force {
				time.Sleep(20 * time.Millisecond)
				finish()
			}
			select {
			case err := <-shutdown:
				if force {
					require.ErrorIs(t, err, context.DeadlineExceeded)
				} else {
					require.NoError(t, err)
				}
			case <-time.After(time.Second):
				t.Fatal("shutdown exceeded budget")
			}
			select {
			case err := <-result:
				if force {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			case <-time.After(time.Second):
				t.Fatal("RPC did not stop")
			}
			finish()
			select {
			case err := <-server.Notify():
				require.NoError(t, err)
			case <-time.After(time.Second):
				t.Fatal("server did not exit")
			}
			_, err := echov1.NewEchoServiceClient(conn).Echo(authContext(t, svc), &echov1.EchoRequest{Message: "hello"})
			require.Equal(t, codes.Unavailable, status.Code(err))
		})
	}
}
