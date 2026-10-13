# Foundation Capability Validation

Updated: October 13, 2026.

Status: Asynq is complete and merged through PR #70. Full-profile WebSocket implementation is available; verification and merge evidence are tracked in #67. Application gRPC remains planned.

## Purpose and Baseline

The standalone Order, Inventory, and Billing applications exist to test gobase's foundation with Gin, stdlib, and Fiber. Their existing business workflow is sufficient for that purpose. Customer cancellation, invoice voiding, refunds, real payment processing, and further product features are outside the active plan.

Phases 1 through 5 are complete. The current release is v0.1.1. Master revision `84685ebafbf12c26b49e0fb3fbcca13c72c566da` also includes the Zap v1.28.0 update. The [roadmap](../roadmap.md) records completed work. The [Commerce report](../commerce-verification.md) retains the existing verification results and their limits.

The Full profile includes Asynq clients, email and image task handlers, and `cmd/worker`. Engine CI checks processing, failure recovery, and worker shutdown. Optional WebSocket is disabled by default and uses the existing JWT service. Application gRPC endpoints are not implemented. The existing gRPC dependency supports the OpenTelemetry trace exporter.

## Execution Order and Tracking

| Slice | Outcome | Status | Tracking |
|-------|---------|--------|----------|
| 6.1 | Verify and harden existing Asynq processing | Complete; PR #70 merged as `ab7a30d`, source and six engine/profile jobs passed after merge | [#66](https://github.com/zoe606/gobase/issues/66) |
| 6.2 | Add optional WebSocket support with native engine integration | Full implementation available; verification tracked in issue | [#67](https://github.com/zoe606/gobase/issues/67) |
| 6.3 | Add an optional application gRPC server and client example | Planned; follows 6.2 | [#68](https://github.com/zoe606/gobase/issues/68) |

Each slice has its own issue and implementation PR. Record the proposed revision, verification commands, CI runs, and remaining limitations in that issue. Complete one slice before starting the next. Merge only after the required checks pass. A documentation PR does not complete an implementation issue.

Use small technical examples. Reuse the existing applications or generated-project fixtures without adding Commerce business rules. Reusable runtime code, template changes, and cross-engine verification belong in gobase. The application repositories remain separate.

## 6.1 Asynq Failure and Recovery

Inspect the existing client, enqueue calls, task handlers, retry options, worker configuration, and process lifecycle before changing behavior. Reuse current successful email and image processing coverage. Add tests for failure paths that are not covered.

Use real Redis and controlled tasks with observable results. Keep fixture task handlers and failure injection out of normal application routes. Explicit enqueue failures and existing best-effort welcome-email behavior must retain their documented contracts.

Required scenarios:

- Enqueue a task and verify processing by a separately running worker.
- Fail a task temporarily and verify the configured retry limit and eventual result.
- Submit an invalid payload and verify a documented bounded failure outcome.
- Stop and restart a worker with pending or active work. Verify recovery and the final result.
- Repeat processing of the same logical operation. Verify an idempotent fixture handler preserves one result. Asynq can execute a task more than once; this does not establish exactly-once execution or email delivery.
- Make Redis unavailable. Verify explicit enqueue errors, bounded requests, and worker recovery after Redis returns.
- Send SIGTERM during processing. Verify clean exit and either completed work or recoverable work according to the shutdown timeout.

Confirm that retry configuration matches actual task options. Fix only reproduced foundation defects. Preserve existing valid HTTP responses and the full profile's creation commands. Minimal output must continue to build and start without Redis or a worker.

Completion requires reproducible failure and recovery checks, documented task and worker behavior, passing source quality checks, and passing generated-project CI. Exercise the HTTP-to-queue integration through Gin, stdlib, and Fiber. Shared worker behavior needs one implementation.

The implementation applies producer retry and timeout settings, preserves request contexts, and archives malformed JSON without retry. `integration-test/asynq` provides real Redis checks with separate workers and controlled failures. HTTP tests inspect actual queued options, and existing Docker tests verify email and image processing. See [background tasks](../deployment.md#background-tasks) for commands, configuration, and limits. Issue #66 records completed verification and the merged revision.

## 6.2 Optional WebSocket Support

Start with echo and a small server notification example. Define the route, message format, authentication, origin policy, configuration, and activation before exposing support in generated projects. Activation must be explicit. Existing default routes and startup behavior must remain compatible.

Gin and stdlib use their native `net/http` integration. Fiber uses integration compatible with the template's Fiber v2 dependency. Share message contracts and transport-independent logic where useful. Keep each engine's upgrade and handler API native.

Required scenarios:

- Complete an authenticated handshake and exchange messages with a real WebSocket client.
- Reject missing or invalid credentials, disallowed origins, and invalid upgrade requests.
- Verify ping/pong behavior, read and write deadlines, and message-size limits.
- Disconnect and reconnect a client without leaking connections or goroutines.
- Handle a slow client with bounded writes or buffering.
- Verify interaction with request IDs, logging, compression, and recovery middleware.
- Close active connections during application shutdown within the configured time budget.

Completion requires the same wire-level scenarios against all three engines, race tests, Docker runtime checks, and generated application instructions. Document the supported profiles and library versions. Keep default minimal startup free of new required infrastructure.

The user selected Full first with existing JWT authentication. The implementation adds native upgrades for all three engines and shared bounded sessions. [WebSocket setup](../websocket.md) defines activation, headers, message frames, limits, native library versions, and shutdown. `cmd/ws-client` is the client example. Wire tests use real TCP connections, and `integration-test/websocketcheck` checks the running application. Browser authentication, application broadcasts, and Minimal WebSocket support are outside this slice. Issue #67 records actual verification and merge evidence.

## 6.3 Optional Application gRPC

Use `grpc-go` with a separate server and port. Keep the selected HTTP engine and its native handlers. Start with one unary Echo RPC and a Go client. Reuse application configuration, logging, and use cases where appropriate. Add a dedicated command such as `cmd/grpc` instead of starting another listener unconditionally with the HTTP app.

Define the `.proto` contract, generated Go code, configuration, credentials, supported profiles, and startup commands before implementation. Pin generation tools and provide a reproducible regeneration command. gRPC uses protobuf messages and gRPC status codes; HTTP keeps its JSON response envelope.

Required scenarios:

- Call the service using the generated client and verify valid and invalid requests.
- Verify authentication and documented transport security configuration.
- Verify deadlines, context cancellation, unavailable servers, and error status codes.
- Check health and graceful shutdown with active RPCs and a bounded fallback timeout.
- Run the optional server alongside each HTTP engine and verify independent startup and shutdown.

Completion requires real client/server integration tests, race tests, Docker runtime checks, and setup documentation. Use one shared gRPC implementation. Additional streaming RPC types are separate follow-up work.

## Verification and Compatibility

- Keep the existing 85% coverage gate. Run `make check-all COVERAGE_THRESHOLD=85` before committing.
- Keep all three Full and all three Minimal engine jobs passing. Existing successful HTTP, generation, worker, and Docker checks remain regression coverage.
- Use focused tests for each slice. Use real Redis, worker processes, WebSocket connections, or gRPC connections when verifying those capabilities.
- Generate and verify one engine at a time when the combined verifier reaches its 15-minute limit. Record the commands and actual results.
- Record source, generated-project, and application revisions separately. Do not report an enabled configuration or installed dependency as verified runtime support.
- Update template and generated application documentation when implemented capabilities change setup or feature availability.
- Keep email verification, password reset, and metrics export deferred under issues #49 and #50.

Monorepo generation, new Commerce business flows, an additional cross-service broker, and a general feature-composition system are outside this plan. gRPC and WebSocket are optional capabilities. Minimal projects retain infrastructure-free default startup.
