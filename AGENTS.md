# gobase — Agent Conventions

This file is the **canonical source of project rules** for AI agents. Tool-specific guidance must point here instead of maintaining a separate project description.

## Check the workspace first

Before describing the project, reviewing changes, or editing files, run:

```bash
pwd
git status --short --branch
git log -1 --format='%h %s'
git worktree list
```

- Identify the active checkout and revision before making claims about the project.
- If HEAD is detached, say which revision you inspected. Do not describe it as the latest project state.
- For the current development baseline, compare the checkout with `origin/master`. Fetch when current remote state matters. State when the remote reference has not been refreshed.
- If the checkout and baseline differ, inspect the relevant files at both revisions. Keep task-specific work on its intended branch.
- Preserve local changes and commits when switching branches. Do not remove other worktrees during routine cleanup.

## Overview

- **Stack**: Go
- **Purpose**: Generate Go applications with one selected HTTP engine and profile.
- **HTTP engines**: `gin` (default), `stdlib` (`net/http` with `http.ServeMux`), and `fiber`.
- **Profiles**: `full` (default) and `minimal`.
- **Test runner**: go test ./...

When `pkg/scaffold/scaffold.go` exists, this is the source template. Read `README.md`, `docs/http-engines.md`, and `.github/CONTRIBUTING.md` for the current architecture. Read `docs/roadmap.md` for project status and linked verification. Historical plans and old feature worktrees do not establish the current baseline.

The source template's runnable HTTP application uses Fiber for compatibility. Generated applications use independent native engine templates. Reading only `internal/handlers/http` does not establish which engines gobase supports.

In generated applications, inspect `.gobase.json` and the local README for the selected engine and profile. The metadata controls generation, not runtime engine switching. Do not assume template maintenance tools or documents exist in generated output.

## Source locations

These paths apply to the source template:

| Responsibility | Source |
|---|---|
| Supported engines and profiles | `pkg/project`, `pkg/tools/init` |
| Project generation | `pkg/scaffold/scaffold.go`, `pkg/scaffold/minimal.go` |
| Full endpoint templates | `pkg/scaffold/templates/{gin,stdlib,fiber}` |
| Shared Gin and stdlib helpers | `pkg/scaffold/templates/nethttp` |
| Minimal service templates | `pkg/scaffold/templates/minimal` |
| CRUD handler generation | `pkg/codegen/generator/templates/handler/{gin,stdlib,fiber}` |
| Shared business logic and persistence | `internal/usecase`, `internal/dto`, `internal/entity`, `internal/repo` |
| Cross-engine HTTP contracts | `internal/handlers/http/engine_contract_test.go` |

For shared HTTP changes, update the affected engine templates and the runnable checkout. Keep handlers native to their engine. Put shared business rules in use cases. Use `.github/CONTRIBUTING.md` for optional capability source locations and verification.

## Conventions

- Match existing code style; do not introduce a new framework or pattern without explicit approval.
- Prefer editing existing files over creating new ones.
- Write no comments unless the *why* is non-obvious.
- Use `Handler -> UseCase -> Repository` for application flow.
- Keep each usecase method and its tests in corresponding files. Keep constructors and shared helpers in the feature's main file. Keep domain errors in `errors.go`.
- Keep handler methods in separate files. Related methods may share a file when that matches the existing pattern.
- Use request and response DTOs under `internal/dto/<feature>`. Use package names such as `articledto` to avoid Swagger import conflicts.
- Use `errors.Is` for repository sentinel errors. Map domain errors to HTTP status codes in handlers.
- Use sequential SQL migrations with matching `.up.sql` and `.down.sql` files. Use plural table names.
- Use `pkg/json` for application JSON handling.

## Testing

Run focused tests for the affected packages. `make test` runs the source unit suite with race detection and coverage. `go test ./...` also includes integration tests that need a running application and infrastructure.

Run `/check` or `make check-all` for format, lint, vulnerability checks, tests, and coverage validation. Run `make check-all` before every commit. Do not bypass hooks or ignore failed checks.

In the source template, run `make test-engines` for generated HTTP behavior and `make test-minimal` for minimal output. Use the engine verifier's `-lint` option when required by the contribution guide. Report which checks actually ran and which environment they used.

## Build / Run

- Read the Go version from `go.mod` instead of copying an old version into guidance.
- `make init ENGINE=gin PROFILE=full MODULE=example.com/api APP_NAME=api OUTPUT=../api` creates a new application. The output directory must not exist.
- `make run` runs the HTTP app. `make dev` uses Air hot reload.
- `make build` builds the HTTP app. `make run-worker` runs the Full profile's worker.
- `make docker-services` starts the source template's PostgreSQL and Redis. `make docker-dev` starts its full stack.
- Read `config/config.example.yaml` and `docs/deployment.md` before changing runtime configuration. Keep local configuration and secrets out of commits.

## Don'ts

- Don't commit `.env` files or any secrets.
- Don't run destructive operations (force-push, `git reset --hard`, dropping tables) without explicit user approval.
- Don't assume CI passes; run `/check` locally first.
- Don't bypass checks with `--no-verify`, force-push, or discard unrelated changes.

## Project-specific notes

Generated applications own their code. Updating the source template does not update existing applications. Distinguish source inspection, recorded verification, and checks run in the current session.
