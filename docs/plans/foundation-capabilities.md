# Foundation Capability Validation

Updated: October 15, 2026.

Status: Phase 6 and application gRPC, Asynq, and WebSocket verification are complete. Final post-merge verification is recorded under #79, #81, and #83. v0.2.0 readiness is complete under #85. The release notes record the final tag revision and passing CI runs.

## Purpose and Baseline

The standalone Order, Inventory, and Billing applications exist to test gobase's foundation with Gin, stdlib, and Fiber. Their existing business workflow is sufficient for that purpose. Customer cancellation, invoice voiding, refunds, real payment processing, and further product features are outside the active plan.

Phases 1 through 6 are complete. The current release is v0.2.0. The [release report](../release-readiness.md#v020-capability-release) records the dependency baseline and verification scope. The [roadmap](../roadmap.md) records completed work. The [Commerce report](../commerce-verification.md) retains the existing verification results and their limits.

The Full profile includes Asynq clients, email and image task handlers, and `cmd/worker`. Engine CI checks processing, failure recovery, and worker shutdown. Optional WebSocket and application gRPC use the existing JWT service and remain disabled by default. The telemetry trace exporter is separate from the application RPC server.

## Execution Order and Tracking

| Slice | Outcome | Status | Tracking |
|-------|---------|--------|----------|
| 6.1 | Verify and harden existing Asynq processing | Complete; PR #70 merged as `ab7a30d`, source and six engine/profile jobs passed after merge | [#66](https://github.com/zoe606/gobase/issues/66) |
| 6.2 | Add optional WebSocket support with native engine integration | Complete; PR #77 merged as `e3edeef`, source and six engine/profile jobs passed after merge | [#67](https://github.com/zoe606/gobase/issues/67) |
| 6.3 | Add an optional application gRPC server and client example | Complete; PR #78 merged as `7ae5687`, with seven checks passing before and after merge | [#68](https://github.com/zoe606/gobase/issues/68) |
| Follow-up | Verify application-to-application RPC calls | Implemented and merged through Order #8, Inventory #7, and Billing #7; native/Docker checks pass; final evidence in issue | [#79](https://github.com/zoe606/gobase/issues/79) |
| Follow-up | Verify Asynq through the three application repositories | Merged through Order #9, Inventory #8, and Billing #8; native/Docker PR checks pass; final post-merge evidence in issue | [#81](https://github.com/zoe606/gobase/issues/81) |
| Follow-up | Verify WebSocket through the three application repositories | Merged through Order #10, Inventory #9, and Billing #9; thirty WebSocket groups and existing regressions pass in native/Docker PR CI; final post-merge evidence in issue | [#83](https://github.com/zoe606/gobase/issues/83) |
| Release readiness | Prepare a verified v0.2.0 capability release candidate | Complete; PR #86 merged as `862c65d`; all seven post-merge checks, six fresh outputs, and documented quick starts passed | [#85](https://github.com/zoe606/gobase/issues/85) |
| Release publication | Publish the verified template with current creation instructions | Exact tag revision and passing CI runs are recorded in the release notes | [v0.2.0](https://github.com/zoe606/gobase/releases/tag/v0.2.0) |

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

The implementation follows the Full profile used in 6.2. [gRPC setup](../grpc.md) defines the unary protobuf contract, JWT metadata, TLS configuration, health, limits, and independent process lifecycle. `cmd/grpc` starts only the RPC server. `cmd/grpc-client` uses the generated client. Protobuf generation pins the compiler and both Go plugins. Real TCP/TLS tests cover credentials, deadlines, cancellation, status codes, recovery, and graceful/forced shutdown. Issue #68 records verification and merge evidence. Minimal gRPC support, mTLS, HTTP gateways, business streaming RPCs, and new Commerce flows are outside this slice.

## After Phase 6

The Asynq follow-up is implemented and merged in all three mini-apps. Each app compiles its own worker implementation. Order owns the common native/Docker runner. The normal `cmd/worker` processes email tasks with a noop sender in all three apps. Order's existing registration route verifies the HTTP producer policy. Inventory and Billing do not expose customer registration routes. Controlled sender failures use test-only worker processes and real Redis. Both modes passed retry exhaustion, timeout, pending restart, active crash recovery, SIGTERM, and Redis recovery. The [Asynq application report](../commerce-verification.md#application-asynq-verification) records commands, revisions, and limits. Issue #81 records final CI inspection. Fixture idempotency does not prove real email delivery.

The WebSocket follow-up is implemented and merged in all three mini-apps. It reuses `/v1/ws`, the existing JWT service, and native engine upgrades. Activation remains explicit. Order owns the common runner. Actual processes pass header authentication, explicit origins, frame validation, deadlines, ping/pong, capacity, slow clients, reconnect, and SIGTERM with close code 1001. Each client example connects to its running app. The fixture obtains a token through Order's registration route and shares JWT trust for peer calls. Inventory and Billing keep customer registration disabled. The [application WebSocket report](../commerce-verification.md#application-websocket-verification) records revisions, commands, CI evidence, and limits. Issue #83 records final post-merge inspection. Browser authentication and business broadcasts remain outside this slice.

Release readiness is complete under [#85](https://github.com/zoe606/gobase/issues/85). Candidate `862c65d2542b93a48c9ae14df13da50363ab120c` passed Quality and all six engine/profile checks after PR #86 merged. Fresh output verification used source `b88a84025d0c0ea93e65098a76630b4630a45628`. The [candidate report](../release-readiness.md#v020-capability-release-candidate) records pins, generated projects, documented commands, application evidence, and release notes. The [v0.2.0 release report](../release-readiness.md#v020-capability-release) records publication scope. Readiness covered:

- Record the source revision and review dependency, Go version, and capability documentation against it.
- Generate Full and Minimal projects for all three engines from that revision. Verify quick start, code generation, default startup, and the documented optional worker, WebSocket, and gRPC commands with the existing checks.
- Record the six engine/profile results and the application regressions. Prepare release notes and select v0.2.0 before the separate publication task.

The approved RPC interoperability follow-up is implemented in the existing mini-apps. They adopt the shared runtime from revision `7ae56879be7d8737295172f1c694dde6436fb003`. Each repository owns its proto, generated code, and optional server. Order owns the common runner. Native and Docker checks verify all six directed pairs, three three-service chains, TLS, JWT, bounded failures, and recovery. Revisions, reproducible commands, and CI evidence are recorded in #79 and the [Commerce report](../commerce-verification.md#application-grpc-interoperability). The technical Probe RPC does not add business workflows.

Maintain the verified Asynq, WebSocket, and gRPC paths and fix reproduced foundation defects. Select further capabilities from a concrete application requirement. Browser WebSocket authentication, mTLS, and application RPC streaming remain possible follow-up work, not automatic new phases. Version v0.2.0 is published. The approved next task is [auth email integration #49](auth-email-flows.md), including native handlers, workers, and token-consumption verification for all Full engines. Keep #50 deferred. A new release tag requires its own readiness verification.

## Verification and Compatibility

- Keep the existing 85% coverage gate. Run `make check-all COVERAGE_THRESHOLD=85` before committing.
- Keep all three Full and all three Minimal engine jobs passing. Existing successful HTTP, generation, worker, and Docker checks remain regression coverage.
- Use focused tests for each slice. Use real Redis, worker processes, WebSocket connections, or gRPC connections when verifying those capabilities.
- Generate and verify one engine at a time when the combined verifier reaches its 15-minute limit. Record the commands and actual results.
- Record source, generated-project, and application revisions separately. Do not report an enabled configuration or installed dependency as verified runtime support.
- Update template and generated application documentation when implemented capabilities change setup or feature availability.
- Complete the resumed email verification and password reset slice under #49 with its dedicated plan and verification. Keep metrics export deferred under #50.

Monorepo generation, new Commerce business flows, an additional cross-service broker, and a general feature-composition system are outside this plan. gRPC and WebSocket are optional capabilities. Minimal projects retain infrastructure-free default startup.
