# apps/api

The Go backend. Single binary with three entrypoints. Not part of the Bun workspace.

## Module path

`github.com/apudiu/quranprism/api`.

## Entrypoints

- `cmd/api/` — HTTP server. Serves the public API consumed by `apps/web/` and `apps/admin/`.
- `cmd/worker/` — Background job processor. Consumes the River queue.
- `cmd/cron/` — Periodic scheduler. Enqueues due jobs into River. Uses Postgres advisory locks so only one instance schedules at a time.

## Internal package layout

Internal packages are added under `internal/<domain>/` as the application grows. Standard cross-cutting packages we expect:

- `internal/config/` — Environment-based configuration loading. Single source of truth for env vars.
- `internal/db/` — sqlc-generated code. Do not hand-edit files in `internal/db/sqlc/`; regenerate via `make sqlc`.
- `internal/auth/` — Authentication and session management (defined once we pick the auth model).
- `internal/jobs/` — River job definitions. One file per job kind.
- `internal/email/` — SES client wrapper plus email templates.
- `internal/httpapi/` — HTTP handlers. Wires routes (chi), middleware, request/response types.
- `internal/admin/` — Admin-only handlers. Same package conventions but scoped under a separate router with auth middleware.

Domain-specific internal packages (recitation, translation, user preferences, subscription, etc.) are added as the product surface is defined in `prd/`.

## Conventions

- Error wrapping: `fmt.Errorf("recitation: load surah: %w", err)`. Include enough context that the wrap chain reads as a sentence.
- Sentinel errors are exported. Example: `var ErrSurahNotFound = errors.New("surah not found")` in the relevant package's `errors.go`.
- All exported functions that perform I/O take `ctx context.Context` as the first parameter.
- Logging: use `*slog.Logger` injected via struct field, never `slog.Default()`.
- Tests: table-driven. Use `t.Run` for subtests. No `testify` mocks; write handwritten test doubles in `_test.go` files. Use `github.com/stretchr/testify/require` for assertions only.
- Integration tests use a real Postgres via `testcontainers-go`, not a mock.

## Migrations

- Domain migrations live in `apps/api/migrations/` and are managed by goose.
- River's internal migrations are managed separately via `rivermigrate.New(...).Migrate()`. They live in River's package, not in our migrations directory.
- Migrations are forward-only in production. If a deployed migration was wrong, write a new one to correct it.

## Job queue patterns

- Every job worker implements `river.Worker[ArgsType]`. Args are the only input.
- Side-effecting state lives in struct fields injected at worker construction.
- Idempotency: assume any job may run more than once.
- Long-running jobs check `ctx.Done()` periodically and return cleanly on cancellation.

## Adding a new HTTP endpoint

1. Define request/response types in `internal/httpapi/types.go`.
2. Write the handler in `internal/httpapi/<area>.go`.
3. Register the route in `internal/httpapi/router.go`.
4. Add a table-driven test in `internal/httpapi/<area>_test.go`.
5. If the endpoint requires a new DB query, add SQL to `internal/db/queries/` and regenerate sqlc: `make sqlc`.
