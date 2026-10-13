package echohandler_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	echohandler "go-boilerplate/internal/handlers/grpc/echo"
	echousecase "go-boilerplate/internal/usecase/echo"
	echov1 "go-boilerplate/pkg/api/echo/v1"
	"go-boilerplate/pkg/grpcclient"
	"go-boilerplate/pkg/grpcserver"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

func TestGRPCLogCorrelation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "grpc.log")
	file, err := os.Create(path)
	require.NoError(t, err)
	defer file.Close()
	svc := jwt.New("rpc-log", time.Hour, time.Hour)
	server, err := grpcserver.New(rpcConfig(), svc, logger.New("info", logger.WithFileOutput(file)))
	require.NoError(t, err)
	echov1.RegisterEchoServiceServer(server.GRPC, echohandler.New(echousecase.New()))
	require.NoError(t, server.Start())
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		require.NoError(t, server.Shutdown(ctx))
	}()
	conn, err := grpcclient.New(grpcclient.Options{Address: server.Address(), Insecure: true})
	require.NoError(t, err)
	defer conn.Close()
	accessToken := token(t, svc)
	ctx := metadata.AppendToOutgoingContext(t.Context(), "authorization", "Bearer "+accessToken)
	result, err := echov1.NewEchoServiceClient(conn).Echo(ctx, &echov1.EchoRequest{Message: "private message body"})
	require.NoError(t, err)
	require.NotEmpty(t, result.GetRequestId())
	require.Eventually(t, func() bool {
		data, readErr := os.ReadFile(path)
		return readErr == nil && strings.Contains(string(data), "request_id="+result.GetRequestId())
	}, time.Second, time.Millisecond)
	logs, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(logs), "code=OK")
	require.Contains(t, string(logs), echov1.EchoService_Echo_FullMethodName)
	require.NotContains(t, string(logs), accessToken)
	require.NotContains(t, string(logs), "private message body")
}
