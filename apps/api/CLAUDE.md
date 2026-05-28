# apps/api

The Go backend. Single binary with three entrypoints. Not part of the Bun workspace.

## Module path

`github.com/apudiu/quranprism/api`.

## Entrypoints

Single binary `cmd/qp/` (cobra root). Long-running services + one-shot ops share the same binary, picked by subcommand. Colon-namespaced names (`qp serve:api`, `qp admin:grant`).

- `qp serve:api` — HTTP server. Serves the public API consumed by `apps/web/` and `apps/admin/`.
- `qp serve:worker` — Background job processor. Consumes the River queue.
- `qp serve:cron` — Periodic scheduler. Enqueues due jobs into River. Uses Postgres advisory locks so only one instance schedules at a time.
- `qp migrate:up` / `:down` / `:status` — Apply / roll back / list goose migrations (embedded via `go:embed`).
- `qp migrate:queue` — Apply River's internal migrations.
- `qp seed:run` — Reconcile the ACL permission catalog into Postgres (same logic that runs on every `serve:*` boot). Idempotent; lets deploy pipelines populate perms without booting a server.
- `qp admin:grant --email=… [--group=Admin]` — Bootstrap admin: upsert group, link every catalog perm to it, join the user. Calls the seed inline so it's safe on a freshly-migrated DB. Idempotent. Writes one `audit_log` row with `actor_kind='cli'`.

**Deploy pipeline order (k8s Jobs / CI):** `qp migrate:up` → `qp migrate:queue` → (`qp seed:run` OR run `qp admin:grant` which seeds inline) → start `serve:*` pods. The seed also runs on every `serve:*` boot, so a hot redeploy of an existing cluster picks up new catalog entries automatically.

Adding a new subcommand: drop a `newFooCmd() *cobra.Command` constructor in a new or existing `cmd/qp/<family>.go`, register it in `cmd/qp/main.go`. One-shot commands that don't need the full fx graph load `DATABASE_URL` via the `databaseURL()` helper; long-running services build `fx.New(app.XApp).Run()`.

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
