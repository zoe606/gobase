# Optional Application gRPC

Phase 6.3 is tracked in [#68](https://github.com/zoe606/gobase/issues/68). This document defines the contract before exposing the optional server. Verification and merge evidence are recorded in that issue.

## Profile and Processes

Full projects include one shared grpc-go server and generated Go client. Gin, stdlib, and Fiber retain their native HTTP handlers. Minimal output omits this capability. HTTP and worker startup do not start a gRPC listener.

Run `cmd/grpc` separately. It reads the existing Full configuration and JWT settings. It does not connect to PostgreSQL, Redis, storage, or email. Full configuration validation still applies. For production, use the same valid Full configuration as the HTTP app.

## Contract

`pkg/api/echo/v1/echo.proto` defines `gobase.echo.v1.EchoService/Echo`. The request contains `message`. The response contains the same `message` and a server request ID. Non-empty UTF-8 text up to 4096 bytes is accepted. Empty or oversized text returns `InvalidArgument`.

Send `authorization: Bearer <access-token>` in RPC metadata. The existing JWT service validates the signature and expiry. Missing, invalid, expired, and refresh credentials return `Unauthenticated`. Echo calls are bounded by token expiry and `GRPC_REQUEST_TIMEOUT`. The server accepts or creates `x-request-id` and returns it in response metadata. Logs contain the method, request ID, duration, and status without tokens or request bodies. Handler panics return generic `Internal` errors.

gRPC uses protobuf messages and status codes. HTTP keeps its existing JSON envelope. There is no HTTP gateway or business streaming RPC in this slice.

The standard `grpc.health.v1.Health/Check` and `Watch` endpoints are public for readiness. The empty service name and `gobase.echo.v1.EchoService` report `SERVING` after startup. Shutdown marks health as `NOT_SERVING` and stops accepting new calls. It waits for active calls, then forces Stop when `GRPC_SHUTDOWN_TIMEOUT` expires. Handlers must honor their context.

## Activation and Credentials

Local development:

```bash
export GRPC_ENABLED=true
export GRPC_INSECURE=true
go run ./cmd/grpc
```

The default address is `127.0.0.1:50051`. Plaintext requires explicit `GRPC_INSECURE=true` and is rejected in production. Local Go commands read exported variables and YAML, not `.env` automatically.

For TLS, set `GRPC_TLS_CERT_FILE` and `GRPC_TLS_KEY_FILE` and leave `GRPC_INSECURE=false`. The server requires TLS 1.3. The client verifies the server name and certificate chain. No skip-verification option is provided. JWT authentication still applies over TLS. Mutual TLS is separate follow-up work.

```bash
export GRPC_ACCESS_TOKEN='your-access-token'
go run ./cmd/grpc-client -address=127.0.0.1:50051 -insecure -message=hello
# TLS client with a private CA:
go run ./cmd/grpc-client -address=localhost:50051 -ca-file=path/to/ca.pem -message=hello
```

Use `-timeout` to bound the client call. HTTP access tokens can be used when both processes share JWT configuration. The server closes independently of the HTTP app.

## Configuration

| Environment variable | Default |
|----------------------|---------|
| `GRPC_ENABLED` | `false` |
| `GRPC_ADDRESS` | `127.0.0.1:50051` |
| `GRPC_INSECURE` | `false` |
| `GRPC_TLS_CERT_FILE` | empty |
| `GRPC_TLS_KEY_FILE` | empty |
| `GRPC_REQUEST_TIMEOUT` | `5s` |
| `GRPC_SHUTDOWN_TIMEOUT` | `15s` |
| `GRPC_CONNECTION_TIMEOUT` | `5s` |
| `GRPC_MAX_MESSAGE_SIZE` | `65536` bytes |
| `GRPC_MAX_CONCURRENT_STREAMS` | `128` per connection |

The optional Docker Compose `grpc` profile builds `TARGET=grpc` and binds container port 50051. It does not add database or Redis dependencies. Set the gRPC values in `.env` before starting that profile. Keep TLS and JWT key files outside the image and mount them read-only.

After setting `GRPC_ENABLED=true` and development transport or TLS settings in `.env`:

```bash
docker compose -f deployment/docker/docker-compose.yml -f deployment/docker/docker-compose.app.yml --profile grpc up --build grpc
docker compose -f deployment/docker/docker-compose.yml -f deployment/docker/docker-compose.app.yml --profile grpc stop grpc
```

These commands start and stop only the optional RPC service. Add read-only mounts through a Compose override when using certificate or JWT key files.

## Protobuf Generation

`make proto` uses protoc v36.2, protoc-gen-go v1.36.11, and protoc-gen-go-grpc v1.6.2. The Go plugins are pinned in `go.mod`. The script verifies the platform-specific protoc archive checksum. Linux and macOS on amd64 and arm64 are supported. Generated code is checked in, so normal builds do not require protoc.

The `.proto` contract uses an import mapping supplied from `go list -m`. This keeps generated descriptors stable when the initializer changes the application's module path. Regenerate and check `git diff --exit-code -- pkg/api/echo/v1` to verify the checked-in output. Each generated engine repeats this check.

Runtime dependencies are grpc-go v1.83.2 and protobuf-go v1.36.11. The grpc-go patch fixes [GO-2026-6443](https://pkg.go.dev/vuln/GO-2026-6443), flagged by govulncheck in the new server call path. The existing telemetry exporter is separate from the application RPC server.

## Verification

```bash
go test -count=1 -race ./config ./pkg/jwt ./internal/usecase/echo ./internal/handlers/grpc/...
make proto
git diff --exit-code -- pkg/api/echo/v1
go run ./integration-test/grpccheck -insecure -http-url=http://localhost:8080
```

The runtime check needs a running HTTP app and optional RPC process sharing JWT configuration. It creates a development user, checks health, uses the HTTP access token for Echo, and verifies authentication and input errors. Native and Docker CI checks repeat this with all three HTTP engines. Tests use real TCP/TLS connections and controlled test-only use cases for deadlines, cancellation, panics, and active shutdown.
