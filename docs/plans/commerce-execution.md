# Commerce Application Verification

Updated: October 9, 2026.

## Approved Direction

Build three standalone applications with different business domains. Order uses Gin, Inventory uses stdlib, and Billing uses Fiber. Each application owns its repository, Go module, data, configuration, CI, and deployment. Use HTTP for the initial workflow. A message broker is an optional later exercise.

Tracking: [applications #60](https://github.com/zoe606/gobase/issues/60) and [HTTP workflows #56](https://github.com/zoe606/gobase/issues/56). The superseded monorepo issue #55 is closed as not planned.

## Initial Implementation Scope

The first business workflow supports one SKU per order and integer monetary amounts in IDR. Inventory owns product prices and stock. Order records the accepted price and quantity. Billing owns invoices and recorded payment state. Real payment-provider integration is outside this first implementation; record that limitation in the evidence report.

Use the full profile to exercise generated authentication, persistence, HTTP middleware, migrations, and runtime behavior. Keep business logic in each application's use cases and repositories. Keep each engine's native handler API. Generated foundation fixes belong in gobase and must be adopted by affected applications.

Local application repositories will be kept separately under the ignored `.worktrees` directory during implementation: `commerce-order`, `commerce-inventory`, and `commerce-billing`. They have independent Git metadata and no shared Go workspace. Remote publication and an external deployment target are not selected yet.

## Business and HTTP Contract

| Application | Operation | Required result |
|-------------|-----------|-----------------|
| Order | Create an authenticated customer's order with an idempotency key | Persist the order, reserve stock, request its invoice, and expose its workflow state |
| Order | Read the customer's order | Return only an order owned by the authenticated customer |
| Order | Receive authenticated payment confirmation | Verify the invoice state and complete stock consumption before completing the order |
| Inventory | Create or update a product through an authenticated operations API | Store SKU, positive unit price, and nonnegative stock |
| Inventory | Reserve stock for an order | Reject insufficient stock; repeat the same operation without reserving twice |
| Inventory | Release or consume a reservation | Apply the state change once and preserve inventory totals |
| Billing | Create an invoice for an order | Verify the order's price snapshot over HTTP and create one invoice per order |
| Billing | Record a payment reference | Persist payment state once and send a separate HTTP confirmation to Order |
| Billing | Retry an unconfirmed payment notification | Preserve payment state and recover the order workflow without recording another payment |

Each service uses explicit credentials for its internal APIs. Public customer requests use the generated authentication flow. Configure internal credentials through local environment or ignored configuration files. No service reads another service's database.

Persist each local state change before network calls. Do not hold a database transaction while calling another service. Keep HTTP client timeouts and response-size limits bounded. Propagate request IDs. Repeated write calls must use their operation identity and reject a changed payload for an existing identity.

A failed invoice request leaves an identifiable order and reservation for safe retry or recovery. A payment confirmation failure leaves a paid invoice with pending confirmation. Repeating confirmation must not consume stock twice. Define the exact status codes and error codes with the implementation and test them against actual processes.

## Verification

- Run the existing generated project checks and focused business tests in each repository.
- Use real PostgreSQL and Redis fixtures where the chosen application uses them.
- Exercise the complete workflow through HTTP, including ownership and service credential failures, insufficient stock, duplicate requests, concurrent reservations, missing and slow downstream services, partial failure, restart, recovery, and shutdown.
- Run production configuration in a controlled container environment. Verify explicit migrations, non-root identity, configuration and storage access, database backup and restore, retained state, and shutdown.
- Define workload criteria before measuring. Record the environment, workload, latency, error rate, resource usage, and limits of the result.
- Keep the production-readiness conclusion limited to recorded evidence. Local production configuration does not prove an unselected external deployment or payment provider.

## Current State

The product domains and engine choices are approved. Commerce application code has not been implemented yet. README, architecture, and roadmap corrections are in progress on `feat/minimal-service-output`.

The October 9 vulnerability scan found newly reported standard-library and x/net findings in the previous Go 1.27.1 baseline. The source now uses Go 1.27.2 and x/net v0.60.0. Source `make check-all` passed with no reachable vulnerabilities. Full and minimal output passed generation, module verification, builds, race tests, and lint for all three engines. Minimal vulnerability scans found no reachable vulnerabilities. The October 8 verification remains evidence for its recorded revision and date.
