# Commerce Verification

Updated: October 12, 2026.

## Scope and Revisions

The initial Commerce workflow is implemented in three independent private repositories. Each owns its module, native HTTP handlers, migrations, data, tests, configuration, Dockerfile, and CI. The common test fixture belongs to Order. There is no shared Go workspace.

| Application | Engine | Repository | Verified application revision | Local selective coverage |
|-------------|--------|------------|-------------------------------|--------------------------|
| Order | Gin | [commerce-order](https://github.com/zoe606/commerce-order) | `45baeae241f37218d2f3f5f758b566fc174de900` | 85.6% |
| Inventory | stdlib | [commerce-inventory](https://github.com/zoe606/commerce-inventory) | `1dd4467f4097216f392cae019e4687b439d12d3a` | 85.2% |
| Billing | Fiber | [commerce-billing](https://github.com/zoe606/commerce-billing) | `2cc3899243ad7a7f2e5fe90ae2eb6a0bd625a9ca` | 87.0% |

The applications were generated from gobase `2177675e4ea4409e56905f4b2f966dd4d3b755e5` and adopted the generator workflow test from `631ff36515fce259c0eda237c7716feb359cc96d`. These production changes merged in [PR #59](https://github.com/zoe606/gobase/pull/59) as `a611e8406f491bb259d8a120b89f86b1f7ebaab5`.

Local application paths are `.worktrees/commerce-order`, `.worktrees/commerce-inventory`, and `.worktrees/commerce-billing`. They are separate Git repositories under an ignored directory. Product implementation files are not part of the gobase repository.

## Implemented Workflow

One order contains one SKU and quantity. Monetary amounts are integer IDR values. Inventory owns the price. Order persists the accepted quantity and price. Billing records an invoice and a trusted operator's payment reference. This operation does not charge a payment provider.

```mermaid
sequenceDiagram
    actor Customer
    participant Order as Order / Gin
    participant Inventory as Inventory / stdlib
    participant Billing as Billing / Fiber
    actor Operator
    Customer->>Order: Create with JWT and idempotency key
    Note over Order: Commit pending order
    Order->>Inventory: Reserve by order ID
    Inventory->>Order: Read order intent
    Note over Inventory: Commit reservation and stock change
    Inventory-->>Order: Reservation and price
    Note over Order: Commit reserved price
    Order->>Billing: Create invoice by order ID
    Billing->>Order: Read price snapshot
    Note over Billing: Commit one invoice
    Billing-->>Order: Invoice
    Note over Order: Commit invoiced state
    Operator->>Billing: Record payment reference
    Note over Billing: Commit paid state
    Billing->>Order: Confirm payment
    Order->>Billing: Verify paid invoice
    Order->>Inventory: Consume reservation once
    Note over Order: Commit completed state
    Order-->>Billing: Completed order
    Note over Billing: Commit confirmed notification
```

No database transaction remains open during an HTTP call. A failure preserves the last committed state. The same order key resumes a pending or reserved order. A paid invoice with failed confirmation remains identifiable and can be retried. Repeated reservation, payment, consumption, and confirmation preserve their operation identity.

Internal service and operator credentials are distinct. Inventory's Order credential can read snapshots but cannot confirm payments. Customer reads require ownership. Each application documents exact endpoints and environment variables in its own `docs/commerce.md`.

## Local Verification

The environment uses Go 1.27.2 and golangci-lint 2.14.0 on macOS arm64. PostgreSQL 17 uses TLS with three separately owned databases. Redis 7 uses database numbers 0, 1, and 2. Explicit migrations run before production startup. Fixture host ports are 55430 and 56380.

| Check | Result |
|-------|--------|
| Each application's format, lint, reachable vulnerability scan, race tests, and 85% coverage gate | Passed |
| PostgreSQL repositories and native business handlers with COMMERCE_TEST_DSN | Passed |
| Required credentials, peer URL validation, request IDs, redirects, deadlines, and response limits | Passed |
| Initial workflow with independently running native processes | Ten scenarios passed |
| Non-root Docker images with production configuration | Ten scenarios passed on the recorded application revisions |
| Container identity, config, migrations, CA bundle, temporary files, and writable storage | Passed in local Docker run |
| Restart, retained business state, SIGTERM, and clean final shutdown | Passed |

Coverage excludes the original infrastructure exclusions and generated example HTTP handlers whose routes are disabled in each product. Business handlers, use cases, and repositories remain included. Dormant foundation package unit tests still run. The threshold is 85% in each Makefile and CI workflow.

Inventory locks reservation operations using the canonical PostgreSQL UUID value. A regression test holds the order's lock and verifies that an uppercase UUID cannot bypass it or restore stock before the lock is released. The test failed before the correction and passed afterward. Native and Docker workflow checks passed again with the corrected Inventory revision.

The ten workflow scenarios are:

1. Authentication, access rules, and invalid input.
2. Order, reservation, invoice, payment, stable identities, and conflicts.
3. Eight concurrent orders competing for three stock units. Three succeed and five report insufficient stock.
4. Concurrent creates with the same customer and key. One order and reservation remain.
5. Billing outage after reservation. The same key resumes after recovery.
6. Order outage after payment. The paid invoice retains pending confirmation and recovers.
7. Inventory outage after payment. Retried confirmation consumes stock once.
8. Lost reservation response after its transaction commits. Retrying uses the same reservation.
9. A slow downstream response. The two-second client timeout returns a failure within the tested three-second bound.
10. Application restart. Completed order and confirmed invoice remain persisted.

Repeat from Order after cloning all three application repositories:

```bash
INVENTORY_PROJECT=/absolute/path/to/commerce-inventory \
BILLING_PROJECT=/absolute/path/to/commerce-billing \
bash scripts/commerce-e2e.sh
```

Add `COMMERCE_E2E_DOCKER=1` to run the same workflow with product images. The Docker harness uses `host.docker.internal` on macOS with Docker Desktop or OrbStack. Linux uses host networking so loopback dependencies, peer APIs, and fault proxies remain reachable. Both modes passed in Ubuntu CI as recorded below. Containers have one CPU and 512 MiB each. These limits describe the fixture; they are not measured capacity or a production sizing recommendation.

Process logs and temporary Docker environment files are retained under Order's ignored `integration-test/commerce/.runtime/evidence-<uuid>` directory. Runtime environment files have mode 600. Fixture services retain their test data. Application processes and containers stop after verification.

## Remote CI

Each application's CI applies migrations to PostgreSQL 17 and runs the native handler and repository tests. It also checks formatting, lint, all-package builds, race tests, the coverage threshold, and reachable vulnerabilities. CI results for the business revisions recorded above are:

| Application | GitHub CI | Result |
|-------------|-----------|--------|
| Order | [Run 37912166073](https://github.com/zoe606/commerce-order/actions/runs/37912166073) | Passed |
| Inventory | [Run 37947971768](https://github.com/zoe606/commerce-inventory/actions/runs/37947971768) | Passed |
| Billing | [Run 37912450130](https://github.com/zoe606/commerce-billing/actions/runs/37912450130) | Passed |

These independent quality runs do not execute the full three-application workflow. Order now owns a separate cross-repository workflow, verified below.

### Linux HTTP Workflow

[Order PR #6](https://github.com/zoe606/commerce-order/pull/6) adds Ubuntu native and Docker jobs using the existing ten scenarios. On October 9, 2026, [Commerce E2E run 37954108670](https://github.com/zoe606/commerce-order/actions/runs/37954108670) passed both jobs. [Quality run 37954108608](https://github.com/zoe606/commerce-order/actions/runs/37954108608) also passed. The run head is `265caa41b99b2ddd64e8b9c249a5df08bc8d786a`. The PR checkout is GitHub's merge commit, recorded separately below. The workflow merged on October 11 and is enabled on Order's default branch.

Both artifact revision records identify the same checked-out application commits:

| Application | Checked-out revision |
|-------------|----------------------|
| Order | `658cc72c9702b5ffa1fdab410bdeff423f303df0` |
| Inventory | `1dd4467f4097216f392cae019e4687b439d12d3a` |
| Billing | `7e219a0fe6e9c2a31a9416c406a499e601f94a77` |

The logs confirm Go 1.27.2 on Linux amd64 and ten passed scenarios in each mode. Docker runtime probes passed for all three containers. PostgreSQL TLS, explicit migrations, retained business state, recovery, and clean application shutdown passed. Local native and Docker reruns also passed on the updated harness. Order's local `make check-all` retained 85.6% coverage with the unchanged 85% gate.

Inventory and Billing checkouts use separate read-only deploy keys stored as Order Actions secrets. The keys were verified as read-only. Artifacts contain revision records and runtime logs. Inspection confirmed that environment files and TLS keys are absent. The workflow runs on Order changes and accepts manual candidate peer revisions. Peer pushes run their own CI and do not automatically trigger this common workflow. Each repository remains independently owned.

### Post-Merge Verification

On October 11, [Commerce E2E run 38151098408](https://github.com/zoe606/commerce-order/actions/runs/38151098408) passed all ten scenarios in each mode. [Order Quality run 38151098378](https://github.com/zoe606/commerce-order/actions/runs/38151098378) passed. Logs confirm non-root runtime probes for all three Docker containers. Both jobs record Order `d4fb7e78a1e183a11665730aa008b796d29854dd`, Inventory `1dd4467f4097216f392cae019e4687b439d12d3a`, and Billing `7e219a0fe6e9c2a31a9416c406a499e601f94a77`.

[Gobase PR #62](https://github.com/zoe606/gobase/pull/62) merged as `620c8614dd4ac72fe54ce024013593d888db28a1`. Post-merge [Quality](https://github.com/zoe606/gobase/actions/runs/38150670114) and all six full/minimal [engine jobs](https://github.com/zoe606/gobase/actions/runs/38150670107) passed. Issues #60 and #56 are closed as completed.

### Panic Recovery Follow-up

On October 12, `TestEnginePanicRecovery` reproduced a foundation defect in the three applications and the source checkout. Gin and stdlib returned HTTP 200 when gzip closed during panic handling. Plain requests returned text errors. Fiber recovery also returned text errors instead of the shared JSON envelope.

The fix returns HTTP 500 with `INTERNAL_ERROR` and the request ID. Recovery runs inside compression. Panic details remain in server logs. The source Fiber checkout, full-profile templates, and all three applications adopt the fix. Gin and stdlib cannot replace headers that have already been sent.

`go test -race -count=1 ./internal/handlers/http/...` passed in the source and each application. The regression uses a test use case with the production middleware chain and a live HTTP server. It checks plain and gzip responses, supplied and generated request IDs, hidden panic details, and a successful health request after recovery.

Fresh full-profile Gin, stdlib, and Fiber projects passed module verification, builds, race tests, CRUD generation and wiring, Swagger checks, and lint through `go run ./pkg/tools/verifyengines -lint`. Minimal output is outside this change.

All ten existing Commerce scenarios passed locally in native and Docker modes after the changes. These process tests verify business behavior, dependency failures, recovery, restart, and shutdown. The immutable v0.1.0 release remains evidence for its original revision. Candidate commits need their own CI results.

[Gobase PR #64](https://github.com/zoe606/gobase/pull/64) merged as `339ad3c777afe3197831b063c57c92be109fbe86`. Post-merge [Quality](https://github.com/zoe606/gobase/actions/runs/38202661062) and all six full/minimal [engine jobs](https://github.com/zoe606/gobase/actions/runs/38202661056) passed. The fix is included in v0.1.1.

The application fixes merged through [Order #7](https://github.com/zoe606/commerce-order/pull/7), [Inventory #6](https://github.com/zoe606/commerce-inventory/pull/6), and [Billing #6](https://github.com/zoe606/commerce-billing/pull/6). Their main revisions are Order `0de05cd6e6d09c2106adc8b8d93b477727878125`, Inventory `daabb313c14b2a32e4963659b421b5d74a39eee5`, and Billing `55f2801a042eb774cfd6b799f19bc3e9815a8958`. Post-merge Quality passed for [Order](https://github.com/zoe606/commerce-order/actions/runs/38203192568), [Inventory](https://github.com/zoe606/commerce-inventory/actions/runs/38202732455), and [Billing](https://github.com/zoe606/commerce-billing/actions/runs/38202773047).

Both modes passed all ten scenarios in [Order's post-merge workflow](https://github.com/zoe606/commerce-order/actions/runs/38203192678). Logs record Order `0de05cd6e6d09c2106adc8b8d93b477727878125` and retained peer pins Inventory `6abd22297b9925b44d86b60da25b004468fa110c` and Billing `6406560bc1ae1487022a9793243eef9b54feec73`. Each peer pin is a parent of its main merge commit.

## Remaining Work

The initial workflow and local production configuration are verified within the limits above. This is not a complete production-readiness claim.

- Run the common workflow with candidate peer revisions before adopting them. Update the pinned revisions after verification.
- Fix reusable foundation defects in gobase when found and adopt those fixes in affected applications.

The [foundation capability plan](plans/foundation-capabilities.md) records completed Asynq, WebSocket, and application gRPC work in the template. The existing Commerce workflow remains regression coverage. Application gRPC and Asynq verification are recorded below.

## Application gRPC Interoperability

Updated: October 13, 2026. Tracking: [#79](https://github.com/zoe606/gobase/issues/79).

These checks use the three existing application repositories, separately from Phase 6.3's generated template fixtures. Each application selectively adopts the shared gRPC runtime from gobase `7ae56879be7d8737295172f1c694dde6436fb003`. Each owns its proto, generated Go code, optional server, and generated peer clients. HTTP business operations retain their existing protocol and behavior.

| Application | Engine | Proposed revision | Merged revision | Local selective coverage |
|-------------|--------|-------------------|-----------------|--------------------------|
| Order | Gin | `79a18865486f68e7e3a7fd48e0c680c10387159a` | `31710fec02b5205b6d78656fea6d205884613db4` | 85.7% |
| Inventory | stdlib | `a3f40c21432cccb9ffd27d9c1cdb7e5d36726bc0` | `c12c4acd09700eedcadfdf57bce1d40786b0e655` | 85.4% |
| Billing | Fiber | `ac4d16d3ebda6944407466b2f834faa90a2078e2` | `2b586a43f623c0022bfaaa28d38ba14ebfe83851` | 87.0% |

The application changes merged through [Order #8](https://github.com/zoe606/commerce-order/pull/8), [Inventory #7](https://github.com/zoe606/commerce-inventory/pull/7), and [Billing #7](https://github.com/zoe606/commerce-billing/pull/7).

### Contract and Operation

`commerce.probe.v1.ProbeService/Probe` accepts non-empty UTF-8 text up to 4096 bytes and at most two distinct configured peer names. A response contains the message, actual services visited, and the request ID. Order can call Inventory, which then calls Billing using Inventory's independently generated client. All six directed pairs and all three rotations of the three-service chain are tested.

The separate `cmd/grpc` process uses the existing JWT service. It forwards the access token, request ID, and bounded context. Each configured peer connection verifies TLS. The fixture shares JWT trust between the RPC processes; service-specific production trust and permissions require an application decision. The runtime remains optional and disabled by default.

`Echo` is a sample RPC name, not an HTTP framework. Gin, stdlib, and Fiber retain their native HTTP handlers. Application gRPC uses `google.golang.org/grpc`. Official `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc` generate messages, clients, and server interfaces. Application code implements the handlers.

Each repository's `make proto` pins protoc v36.2, protoc-gen-go v1.36.11, and protoc-gen-go-grpc v1.6.2. The runner compares the three Probe contracts and requires unchanged regenerated output. Runtime versions are grpc-go v1.83.2 and protobuf-go v1.36.11. The source template's module is replaced by each application's own module through the generator import mapping.

### Results and CI

Local `make check-all COVERAGE_THRESHOLD=85` passed independently in each application with its migrated PostgreSQL fixture. Race tests, lint, vulnerability scans, and the unchanged coverage gate passed. Real TCP tests verify active deadline and cancellation propagation to a controlled peer. All three applications regenerate protobuf code without a diff.

Both local native and non-root Docker process runs passed all 14 RPC scenarios and the ten existing HTTP scenarios. The RPC scenarios cover six directed pairs, three chain rotations, HTTP-issued JWTs, invalid input and credentials, verified TLS, canceled/expired client contexts, bounded failure after peer shutdown, restart recovery, rejection of forwarded credentials by a peer with different JWT trust, request correlation, and clean independent shutdown. Active deadline/cancellation propagation is verified by the controlled TCP tests, rather than by slowing normal application handlers.

| Application or common check | PR CI | Post-merge CI |
|-----------------------------|-------|---------------|
| Inventory Quality | [38394262477](https://github.com/zoe606/commerce-inventory/actions/runs/38394262477), passed | [38395404273](https://github.com/zoe606/commerce-inventory/actions/runs/38395404273), passed |
| Billing Quality | [38394476275](https://github.com/zoe606/commerce-billing/actions/runs/38394476275), passed | [38395499843](https://github.com/zoe606/commerce-billing/actions/runs/38395499843), passed |
| Order Quality | [38395727478](https://github.com/zoe606/commerce-order/actions/runs/38395727478), passed | [38420737749](https://github.com/zoe606/commerce-order/actions/runs/38420737749), tracked in #79 |
| Common native and Docker | [38395727490](https://github.com/zoe606/commerce-order/actions/runs/38395727490), both passed | [38420737754](https://github.com/zoe606/commerce-order/actions/runs/38420737754), tracked in #79 |

The common PR jobs checked out Order's synthetic merge revision `b0bb91bd8b94d03ce829ec05497972d40b1431be`, Inventory `c12c4acd09700eedcadfdf57bce1d40786b0e655`, and Billing `2b586a43f623c0022bfaaa28d38ba14ebfe83851`. Both raw logs contain passing results for each RPC scenario and the existing HTTP workflow. The post-merge jobs use Order `31710fec02b5205b6d78656fea6d205884613db4` with the same peer pins. Issue #79 records the final inspection of those runs.

### Reproduce and Limits

Run from the Order repository with the recorded peer revisions checked out:

```bash
INVENTORY_PROJECT=/absolute/path/to/commerce-inventory \
BILLING_PROJECT=/absolute/path/to/commerce-billing \
bash scripts/commerce-grpc-e2e.sh

COMMERCE_E2E_DOCKER=1 \
INVENTORY_PROJECT=/absolute/path/to/commerce-inventory \
BILLING_PROJECT=/absolute/path/to/commerce-billing \
bash scripts/commerce-grpc-e2e.sh
```

Order owns the common runner and CI. Each peer retains its independent repository, module, database, deployment, and quality checks. Credentials and fixture certificates remain under ignored `.runtime` directories; CI uploads only logs and revision records. The new local worktrees are `.worktrees/commerce-order-grpc`, `.worktrees/commerce-inventory-grpc`, and `.worktrees/commerce-billing-grpc`.

These results establish protobuf generation and actual RPC interoperability. Business workflows still use HTTP. A broker, monorepo, gateway, business streaming, production deployment, workload capacity, and a new release tag are outside this verification. Maintain the existing paths and choose further capabilities from concrete requirements.

The items below are unverified Commerce operations and product capabilities outside the active plan. They are not boilerplate features or completion requirements for the current engine validation.

- Select an external deployment environment and verify TLS, credentials, migrations, and runtime behavior there.
- Verify backup and restore into isolated databases. Existing application data must remain intact.
- Agree on workload, resource limits, latency, throughput, and error criteria before measuring performance.
- Define coordinated customer cancellation, reservation release, invoice voiding, and refunds. Inventory release currently requires an operator to stop concurrent retries and reconcile the order and invoice.
- Select and test a payment provider when real charging is required. Operator payment recording is the current substitute.
- Replace inherited example Swagger specifications with the Commerce API before enabling Swagger.
- Verify credential rotation and database certificate identity. The local fixture uses `sslmode=require`, which encrypts transport but does not verify certificate identity.

Issues [#60](https://github.com/zoe606/gobase/issues/60) and [#56](https://github.com/zoe606/gobase/issues/56) are closed. Later Commerce operations are separate work. [#55](https://github.com/zoe606/gobase/issues/55) is closed as not planned. The [execution plan](plans/commerce-execution.md) and [roadmap](roadmap.md) track maintenance and deferred work.

## Application Asynq Verification

Updated: October 14, 2026. Tracking: [#81](https://github.com/zoe606/gobase/issues/81).

The three independent repositories adopt the existing task policy, context handling, worker configuration, and malformed-JSON behavior from gobase PR #70. The source files are taken from master `3dcb68d70a5495c1994eb2ab0add246a5045b031`. This adoption does not enable WebSocket or change Commerce HTTP business calls.

### Contract and Test Processes

Producer configuration applies `ASYNQ_MAX_RETRY` and `ASYNQ_JOB_TIMEOUT` to newly queued tasks. Existing tasks retain their recorded options. Welcome-email enqueue remains best effort. Image enqueue errors remain logged without failing a completed upload. Malformed JSON is archived without retry. Worker errors report each task's actual retry limit.

Order owns `scripts/commerce-asynq-e2e.sh`. Each repository builds and runs its normal `cmd/worker` against its own migrated PostgreSQL database and a separate Redis database. All three normal workers process retained technical email tasks with the noop sender. Order's existing customer registration route verifies pending task metadata with two retries and a three-second timeout before a normal worker consumes its welcome email. Inventory and Billing keep customer registration routes disabled.

The PostgreSQL fixture health check uses TCP. Its temporary initialization server accepts only Unix-socket connections, so a socket probe can pass before the three application databases are ready. A fresh isolated fixture confirmed TCP health passes after initialization and all three databases exist before migrations begin.

Each application also compiles its own worker constructor, email handler, and Asynq wrapper into a separate test process. A test-only sender injects temporary failures, permanent failures, delays, and a crash after an effect is stored. Its Redis result record provides the observable idempotent effect. Normal application routes expose no failure controls.

### Local Results

Local native and non-root Docker runs passed five controlled-worker groups in each repository: retry/archive, pending/repeated execution, active kill recovery, timeout/active shutdown, and Redis interruption/recovery. The suites verify exact attempt and retry counts, one logical effect after repeated execution, malformed JSON archived without retry, bounded explicit enqueue errors, restart, and clean SIGTERM. Active kill recovery took about 61 seconds while the lease expired.

Each mode also passed the three normal-worker cases and Order's HTTP producer case. The common runner retains all ten HTTP and fourteen gRPC regression cases, which passed in both modes. Docker worker images and fixture processes run as UID/GID `65532`. Normal workers use their standard shutdown budget; controlled fixtures use shorter retry and shutdown intervals.

`make check-all COVERAGE_THRESHOLD=85` passed in each app with the migrated PostgreSQL fixture. Local selective coverage is Order 85.6%, Inventory 85.3%, and Billing 86.9%. Race tests, lint, vulnerability scans, workflow lint, and shell syntax checks passed. The coverage threshold remains 85%.

| Application | Engine | Proposed revision | Merged revision | Local selective coverage |
|-------------|--------|-------------------|-----------------|--------------------------|
| Order | Gin | `c4b10b38a6b743ed78e8f11551ed2cfb145e8c7b` | `a0f60e9acf22140fd60f9578e4d515eb641a54d5` | 85.6% |
| Inventory | stdlib | `9782d94105d781b93a7134178901a80b60216d6a` | `f2551248b1f0da75f776ba85ba850a2068e82574` | 85.3% |
| Billing | Fiber | `89756b939925f147e3e576f9e1d6bbeee0f83269` | `00c314962377d36aa343dbc0e5947bd1c80f66fa` | 86.9% |

Order [PR #9](https://github.com/zoe606/commerce-order/pull/9), Inventory [PR #8](https://github.com/zoe606/commerce-inventory/pull/8), and Billing [PR #8](https://github.com/zoe606/commerce-billing/pull/8) are merged. Their PR and merged trees match. Order pins the recorded merged peer revisions.

| Application or common check | PR CI | Post-merge CI |
|-----------------------------|-------|---------------|
| Inventory Quality | [38428950919](https://github.com/zoe606/commerce-inventory/actions/runs/38428950919), passed | [38429680173](https://github.com/zoe606/commerce-inventory/actions/runs/38429680173), passed |
| Billing Quality | [38429025299](https://github.com/zoe606/commerce-billing/actions/runs/38429025299), passed | [38429727274](https://github.com/zoe606/commerce-billing/actions/runs/38429727274), passed |
| Order Quality | [38430570253](https://github.com/zoe606/commerce-order/actions/runs/38430570253), passed | [38507221124](https://github.com/zoe606/commerce-order/actions/runs/38507221124), final result in #81 |
| Common native and Docker | [38430570404](https://github.com/zoe606/commerce-order/actions/runs/38430570404), both passed | [38507221012](https://github.com/zoe606/commerce-order/actions/runs/38507221012), final results in #81 |

Both common PR artifacts record Order's synthetic merge revision `9f51c2492839b4b6892509dfac464ab1b15db546` and the two merged peer revisions. The Order tree matches the proposed head and the main merge commit. Each raw log contains all fifteen controlled-worker groups, four normal-worker/HTTP producer cases, ten HTTP cases, and fourteen gRPC cases. Each mode uploads 45 files containing only logs and revision records. Issue #81 records final inspection of post-merge runs and source documentation checks.

### Reproduce and Limits

Run from Order with separate peer checkouts:

```bash
INVENTORY_PROJECT=/absolute/path/to/commerce-inventory \
BILLING_PROJECT=/absolute/path/to/commerce-billing \
bash scripts/commerce-asynq-e2e.sh

COMMERCE_E2E_DOCKER=1 \
INVENTORY_PROJECT=/absolute/path/to/commerce-inventory \
BILLING_PROJECT=/absolute/path/to/commerce-billing \
bash scripts/commerce-asynq-e2e.sh
```

Each repo documents standalone Redis checks in `docs/asynq-verification.md`. The common runner and CI keep logs under Order's ignored `.runtime` directory. CI retains only logs and exact application revisions. It does not upload environment files, certificates, or private keys. Fixture cleanup deletes only its own queues and result keys. Database volumes are retained.

Normal-worker checks use a noop sender. Controlled worker checks verify fixture idempotency, not external email delivery or exactly-once execution. Generated-template HTTP and image tests remain separate evidence. No new Commerce business feature, shared broker, production deployment, capacity measurement, WebSocket adoption, or release tag is part of this slice. WebSocket verification through the mini-apps and release readiness follow separately.
