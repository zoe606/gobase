# Email Verification and Password Reset

Tracking: [#49](https://github.com/zoe606/gobase/issues/49).
Baseline: v0.2.0, `369deaef2b9dfa591e05c0c44666331002c6604c`.
Status: Implemented and verified locally. CI acceptance and PR merge are tracked in #49.

## HTTP Contract

Full projects expose these native routes for Gin, stdlib, and Fiber. Minimal omits authentication. Existing register, login, refresh, logout, and current-user contracts remain unchanged. Verification does not add a login requirement.

| Method and path | Input | Authentication | Success |
|-----------------|-------|----------------|---------|
| `POST /v1/auth/verify-email` | JSON `token` | Token possession | 200 with a message |
| `POST /v1/auth/resend-verification` | No body | Existing access JWT; only its user | 202 with a message |
| `POST /v1/auth/forgot-password` | JSON `email` | Public | 202 with the same message for known and unknown email addresses |
| `POST /v1/auth/reset-password` | JSON `token`, `new_password` | Token possession | 200 with a message |

Use the existing JSON envelope, request IDs, body parsing, and validation errors. Invalid, expired, or consumed tokens return HTTP 400 with `INVALID_VERIFICATION_TOKEN` or `INVALID_RESET_TOKEN`. Missing or malformed fields retain `VALIDATION_ERROR`. Queue failures return HTTP 503 with `AUTH_EMAIL_UNAVAILABLE`. Disabled verification returns HTTP 503 with `EMAIL_VERIFICATION_DISABLED`.

## Processing and Configuration

Queue password-reset requests before looking up the email address. The worker performs the lookup and sends only for existing users. Known and unknown addresses therefore use the same HTTP queue path, including queue failures. Queue verification requests by authenticated user ID. Workers use the existing sender and Asynq retry/timeout configuration. Email delivery failure is retryable. Invalid task payloads are archived without retry.

Identical request payloads share a task ID within each calendar minute. This reduces duplicate email requests and is not a rolling per-account rate limit. A new delivery attempt replaces prior tokens for that user and purpose. A retry can invalidate an earlier link.

Normal token-email delivery requires configured Resend credentials and a sender address. Missing credentials produce HTTP 503 before account lookup. Noop remains available for existing development welcome-email processing. Noop is not evidence of token-email delivery. Development auto-verification applies only to the authenticated verification request. Production does not permit auto-verification through that route. Verification token consumption does not require the provider to be available. Password-reset token consumption also remains available during a provider outage.

Email links use the configured frontend base URL and `/verify-email#token=...` or `/reset-password#token=...`. The frontend reads the fragment and submits the token in the corresponding POST body. Use configured absolute HTTP(S) URLs, not request Host headers. Production URLs require HTTPS. This work does not add a frontend.

Store only token hashes in PostgreSQL. Tokens are random, expiring, and consumed once. Token consumption, user updates, and password-reset refresh-token deletion use a database transaction. Conditional token consumption rejects concurrent replay. Existing access JWTs expire according to their existing TTL; password reset invalidates refresh tokens.

## Verification

- Reuse focused auth, repository, email, worker, and shared HTTP contract tests.
- Exercise actual HTTP processes, real PostgreSQL and Redis, a separate worker, and a controlled email sender for every generated Full engine.
- Inspect email links and consume tokens. Cover expiry, repeated and concurrent consumption, unknown email, disabled/unavailable configuration, malformed payloads, and delivery retry/failure.
- Preserve the existing Full and Minimal generation and runtime checks and 85% coverage threshold. Run `make check-all COVERAGE_THRESHOLD=85` before committing.
- Record controlled delivery separately from external Resend delivery. Do not claim provider delivery without provider evidence.

Complete Swagger, DTOs, generated instructions, README, roadmap, and issue evidence with the implementation. Keep #50 deferred. This work does not publish another release automatically.

## Local Results (2026-10-15)

`make check-all COVERAGE_THRESHOLD=85` passed with 87% coverage, clean lint, race tests, and no called vulnerabilities reported by `govulncheck`. Swagger regeneration, focused configuration and auth tests, scaffold race tests, and workflow validation with `actionlint` passed.

| Full engine | Generate, build, race tests, CRUD, protobuf, Swagger | Lint | Real HTTP, PostgreSQL, Redis, and worker fixture |
|-------------|---------------------------------------------------|------|------------------------------------------------|
| Gin | Passed | Passed | 12 scenarios passed |
| stdlib | Passed | Passed | 12 scenarios passed |
| Fiber | Passed | Passed | 12 scenarios passed |

Generation used `go run ./pkg/tools/verifyengines -engine=<engine> -output=<directory>` for each engine. Each generated project ran `go test -count=1 -race -timeout=4m -v ./integration-test/authemail` with `AUTH_EMAIL_TEST_ENABLED=true` and dedicated PostgreSQL and Redis configuration. The [setup guide](../auth-email.md#verification) records the complete fixture command and its cleanup behavior.

The fixture verifies transaction rollback, concurrent replay rejection, refresh-token revocation, expiry, known and unknown addresses, retry exhaustion, malformed tasks and HTTP bodies, queue failure, and configuration behavior. It also checks that auth tokens and replacement passwords are absent from process logs when request body logging is enabled, including mixed-case paths and trailing slashes.

The Engines workflow now runs this fixture for every Full engine. Existing Full and Minimal jobs remain required regressions. Provider delivery uses a controlled sender in these tests. Real Resend inbox delivery requires separate provider verification.
