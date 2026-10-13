# Optional WebSocket Support

Implementation and verification are tracked in [#67](https://github.com/zoe606/gobase/issues/67). This document defines the Phase 6.2 contract. Passing runtime and CI results are recorded separately in that issue.

## Scope and Activation

The Full profile supports WebSocket with Gin, stdlib, and Fiber. The feature is disabled by default. Minimal projects retain their existing startup and do not include this authenticated endpoint.

Enable the feature with exported environment variables or the corresponding `websocket` settings in `config/config.yaml`:

```bash
export WEBSOCKET_ENABLED=true
export WEBSOCKET_ALLOWED_ORIGINS=http://localhost:3000
make run
```

Local Go processes read exported variables and YAML. They do not load `.env` automatically. Docker uses the same environment variables. Origins must be explicit HTTP or HTTPS origins without paths. Wildcards and missing Origin headers are rejected. WebSocket origin checks are independent of HTTP CORS.

## Handshake and Messages

Connect to `ws://localhost:8080/v1/ws` using GET with `Authorization: Bearer <access-token>` and an allowed `Origin`. Use `wss://` when connecting over TLS. The existing JWT service checks the signature and expiry. The connection closes when the token expires. A refresh token cannot authenticate this endpoint.

The example uses a Go client that can set handshake headers. Browser WebSocket APIs cannot set an Authorization header. Browser authentication requires a separate handshake-ticket or cookie design and is outside this slice. Tokens are not accepted in URLs.

```bash
export WS_ACCESS_TOKEN='your-access-token'
go run ./cmd/ws-client -url=ws://localhost:8080/v1/ws -origin=http://localhost:3000
```

The server sends a notification after connecting:

```json
{"type":"notification","data":"connected","request_id":"request-id"}
```

Send a text frame containing:

```json
{"type":"echo","data":"hello"}
```

The response keeps the data and adds the handshake request ID. The server owns `request_id`; client-provided values do not replace it. Binary frames, invalid JSON, and unsupported message types close the connection.

WebSocket messages use this frame contract rather than the HTTP response envelope. Rejected handshakes use the existing HTTP JSON error envelope. Request ID, logging, rate limiting, and recovery middleware still run. HTTP gzip and WebSocket frame compression are separate. Frame compression is disabled.

## Limits and Shutdown

| Setting | Default |
|---------|---------|
| `WEBSOCKET_HANDSHAKE_TIMEOUT` | `5s` |
| `WEBSOCKET_READ_TIMEOUT` | `60s` |
| `WEBSOCKET_WRITE_TIMEOUT` | `10s` |
| `WEBSOCKET_PING_INTERVAL` | `25s` |
| `WEBSOCKET_MAX_MESSAGE_SIZE` | `65536` bytes |
| `WEBSOCKET_MAX_CONNECTIONS` | `128` per process |

The server sends ping frames. Pong frames extend the read deadline up to the JWT expiry. A client must keep reading to process these frames. Each connection has one reader and one writer. Outgoing echo buffering is bounded, and writes have deadlines. Connection limits include upgrades in progress.

The application tracks connections separately from ordinary HTTP requests. SIGTERM rejects new connections, sends close code 1001 to active clients, and closes their sockets within `HTTP_SHUTDOWN_TIMEOUT`. Handshake and connection failures release their reserved slot. An upgrade that never starts its session releases the slot when the handshake timeout expires. Reconnecting creates a new connection and needs a valid token.

Gin and stdlib use `github.com/gorilla/websocket` v1.5.3. Fiber v2 uses `github.com/gofiber/contrib/websocket` v1.3.4. Native handlers perform the upgrade. Shared session code implements messages, deadlines, connection tracking, and shutdown.

Fiber uses `github.com/fasthttp/websocket` v1.5.12. Its adapter closes the underlying socket so active reads stop before the fasthttp callback returns. It retains the framework's default ownership of hijack wrappers.

## Verification

```bash
go test -count=1 -race ./internal/handlers/http -run '^TestEngineWebSocket'
go test -count=1 -race ./config ./pkg/websocket
go run ./integration-test/websocketcheck -url=http://localhost:8080 -origin=http://localhost:3000
```

The runtime check creates a test user through the existing registration endpoint. Use a separate development database. Its `-wait-shutdown` mode holds an authenticated connection and requires close code 1001 when the app stops. Unit fixtures use real TCP connections without database or Redis. Each generated engine runs the same wire-level scenarios.
