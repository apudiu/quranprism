# quranprism

Multilingual Quran listening service. Plays Quran recitation interleaved with user-configurable translations — Arabic → English → Spanish → Bangla → … verse by verse. One source text refracted into every language the listener wants.

## Repository shape

Polyglot monorepo. Bun manages JS/TS workspaces. The Go module is independent (not a Bun workspace).

- `apps/api/` — Go backend, three `cmd/` entrypoints (`api`, `worker`, `cron`) + `cmd/river-migrate` helper. Modular monolith composed with Uber fx; see `docs/decisions/0001-di-fx-modular-monolith.md`.
- `apps/api/internal/platform/` — cross-cutting infra (config, logger, db, cache, pubsub, queue, mailer, jwt). Never imports `internal/modules/*`.
- `apps/api/internal/transport/http/` — router, middleware, JSON envelope, typed-error → status mapping. Leaf `router.Registrar` interface (`http/router/`) breaks the modules-vs-app import cycle.
- `apps/api/internal/modules/<name>/` — domain modules (NestJS-style bundle: service + repo + handler + dto + errors + tests). Each exports an `fx.Module`; handlers self-register on the chi router via `group:"routes"`.
- `apps/web/` — SolidStart user app. Workspace `@qp/web`.
- `apps/admin/` — SolidStart admin app. Workspace `@qp/admin`.
- `pkg/` — Shared TS packages, added on demand. Naming `@qp/<name>`.
- `deploy/k8s/` — production manifests. `deploy/nats/dev.conf` — dev-only NATS config (JetStream + WebSocket).
- `docker-compose.yml` — Local dev stack (postgres + redis + nats + mailpit + api + web + admin), repo root.
- `docs/` — Architecture, ADRs, legal.
- `prd/` — Product requirements + business decisions (see **Product requirements**).
- `tasks/` — Progress tracking (see **Progress tracking**).

## Pinned versions

Authoritative. Prefer latest stable for new deps; bump pins deliberately and update this list in the same PR.

- Go 1.26 (min 1.26.0) · Bun 1.3.14 (`.bun-version`) · PostgreSQL 18.4 · Redis 7.4 · NATS 2.10 · Mailpit v1.20 (dev SMTP)
- SolidStart `@solidjs/start` 2.0.0-alpha + `@solidjs/vite-plugin-nitro-2` (alpha — pin exact, bump deliberately) · `solid-js` 1.9.x
- DI: `go.uber.org/fx` v1.24.x (Nest-style modular composition, see `docs/decisions/0001-di-fx-modular-monolith.md`)
- pgx `github.com/jackc/pgx/v5` v5.9.x · sqlc v1.31.x · goose `github.com/pressly/goose/v3` v3.27.x
- River `github.com/riverqueue/river` v0.38.x (Postgres-atomic job queue) · NATS `github.com/nats-io/nats.go` v1.52.x (pub/sub + WS realtime)
- Cache `github.com/redis/go-redis/v9` v9.20.x · JWT `github.com/golang-jwt/jwt/v5` v5.3.x (HS256, rotating secret list)
- Mail `github.com/wneessen/go-mail` v0.7.x · Env `github.com/caarlos0/env/v11` v11.4.x · UUID `github.com/google/uuid` v1.6.x
- chi `github.com/go-chi/chi/v5` v5.x · AWS SDK Go `aws-sdk-go-v2`

## Stack

Go single binary (api/worker/cron) composed by **Uber fx** · Bun runtime/tooling · SolidStart + TS strict · PostgreSQL (RDS prod, compose local) · **Redis** (cache + rate-limit) · **NATS JetStream** (pub/sub fan-out + browser WebSocket gateway at :8080) · **River** queue (Postgres-atomic; see ADR-0001) · **Mailpit** dev SMTP / AWS SES prod via `wneessen/go-mail` · goose migrations (River migrates itself via `rivermigrate`) · sqlc query layer, no ORM · chi router · JWT HS256 with rotating secret list · OAuth via `golang.org/x/oauth2` (forward-looking, if/when social login is added) · Kubernetes on EC2 · Cloudflare DNS/CDN at the ingress.

## Hard constraints (inviolable)

These mirror `prd/compliance.md` CMP-1 to CMP-6. Violating any of them is grounds to reject a design or a PR.

- **No ads, ever, anywhere in the app.** Marketing pages, profile, settings, listening view — zero ad slots, zero ad SDK references, zero affiliate sneaks. Revenue is Paid-tier subscriptions only.
- **No AI-generated or TTS audio anywhere.** Every audio file (Arabic recitation and translation narration alike) is human-recorded and human-uploaded by admin. Implements `CMP-2`.
- **Translator-purity within a (Surah, Language) slot.** Structurally enforced because each Reciter is bound to exactly one Translator (`REC-2`). Implements `CMP-3`.
- **Whole-Surah Reciter coverage.** A Reciter has audio for ALL Ayahs of any Surah they appear in. Partial-Surah coverage is rejected at publish (`REC-6`). Implements `CMP-4`.
- **Arabic recitation is never time-stretched or pitch-shifted.** When playback-speed control is added in a future release, it applies only to non-Arabic audio tracks; Arabic always plays at 1.0×. Preserves Tajweed pacing. Implements `CMP-5`.
- **Arabic Quran text source is fixed**: ingest only from the Quran.com curated Uthmani Hafs dataset (`CON-4`). No editing of Arabic text by humans or machines once ingested. Implements `CMP-6`.
- **GDPR-equivalent**: data export and deletion honored within 30 days of a verified request (`PRV-1`, `PRV-4`). Implements the standard privacy obligation.

## Things to never do

- Never edit a migration already applied in any environment — write a new one to correct it.
- Never commit secrets — use env vars from Kubernetes secrets.
- Never use npm, pnpm, or yarn. Bun only.
- Never add Go code to a Bun workspace package.
- Never use an ORM in Go — sqlc is the query layer.
- Never downgrade `@solidjs/start` to 1.x, or mix 1.x and 2.x conventions.
- Never leave AI-tooling traces anywhere in the repo or history: no `Co-Authored-By:` agent trailers, "Generated with …" lines, 🤖 markers, or model/agent names in commits, PRs, code, or docs. Human author identity only. (`CLAUDE.md` filenames are the one intentional exception.)
- Never commit, push, or rewrite git history unless explicitly asked. No auto-commits — finish, then await a commit instruction.

## Product rules

Product requirements and business decisions live in `prd/`. The canonical tier matrix (Free vs Paid) is in `prd/README.md`. Free covers everything personal (unlimited playlists, bookmarks, listening, direct user-to-user sharing). Paid unlocks public link share, public profile, and collaborative playlists with sync listening. No ads on either tier.

## Go conventions

- Standard layout: `cmd/` + `internal/`. No `pkg/` inside the Go module.
- Wrap errors `fmt.Errorf("context: %w", err)`. Sentinel errors are exported package vars.
- `log/slog`: JSON handler in prod, text in dev.
- Every function touching DB/HTTP/external services takes `ctx context.Context` first.
- No global state except logger and metrics registry; everything else via struct fields.
- Tests table-driven in `_test.go`, no testify mocks — handwritten doubles.
- `golangci-lint` must pass (CI fails on any lint error).

## SQL conventions

- Forward-only migrations in prod; correct mistakes with a new migration.
- Explicit foreign keys with explicit `ON DELETE`.
- `TIMESTAMPTZ` always, never bare `TIMESTAMP`.
- `snake_case` tables/columns. Indexes named `idx_<table>_<column(s)>`.

## Comments

Write what a human teammate would, only where it earns its place.

- Comment the **why**/intent, not the obvious what.
- Short plain-language note on each non-trivial feature, exported function, type/class, config option — enough to grasp purpose and gotchas at a glance.
- Brief and natural; one line beats a paragraph.
- No bloat: no signature restating, banner art, commented-out code, or narrating trivial lines.

## Frontend conventions

- TS strict throughout. Tailwind only (no CSS modules / styled-components).
- Kobalte for accessible primitives. TanStack Solid Query for server state.
- No prop drilling beyond two levels — lift to context.
- Shared UI primitives in `pkg/ui/` once both apps need them, not before.

## Workspace and package management

Bun only. Shared code in `pkg/<name>/` with `"name": "@qp/<name>"`; root `package.json` already globs `pkg/*`; apps import `@qp/<name>` (Bun symlinks, no dev build step).

- `bun add <pkg>` · `bun add -d <pkg>` (dev) · `bun add --filter @qp/web <pkg>` (one workspace)
- `bun --filter '*' <script>` (all) · `bun --filter @qp/web <script>` (one)
- New shared pkg: create `pkg/<name>/package.json` (`@qp/<name>`), then `bun add @qp/<name>@workspace:*` from a consumer.

## Environment configuration

Runtime config is env-driven and per-workspace. Each app owns three env
files in its own directory:

- `apps/api/.env.dev` / `.env.stg` / `.env.live`
- `apps/web/.env.dev` / `.env.stg` / `.env.live`
- `apps/admin/.env.dev` / `.env.stg` / `.env.live`

`.env.dev` carries real dev-tier values and is committed; `.env.stg` and
`.env.live` are committed templates with placeholder values (real
production secrets land via Kubernetes Secret, never the repo). The
compose stack is dev-only and references each service's `.env.dev`
through `env_file:`. The postgres compose service shares
`apps/api/.env.dev` because the api's `DATABASE_URL` encodes the same
credentials — one source of truth avoids dev drift. Never inline env
values in `docker-compose.yml` — add new keys to the corresponding
`apps/<workspace>/.env.{dev,stg,live}` files instead.

## Development workflow

- Full dev stack: `docker compose up -d` (repo root → postgres + api + web + admin)
- API / worker / cron: `cd apps/api && make run-api` (`run-worker`, `run-cron`)
- Web / admin: `bun run dev:web` · `bun run dev:admin`
- Migrations: `make migrate-up` · `make migrate-down` · `make river-migrate-up` (in `apps/api`)
- Tests: `cd apps/api && make test` · `bun test --filter '*'`
- Lint: `bun run lint` · `cd apps/api && golangci-lint run`
- Docs site (mkdocs-material): `docker compose --profile docs up docs` → http://localhost:8000. Renders `docs/` + the PRD via the `docs/prd` symlink (edit `prd/`, not `docs/prd`).

## Product requirements

Product requirements and business decisions live in `prd/` — per-domain files with stable requirement IDs, indexed by `prd/README.md`.

- Before building or changing a module, read its `prd/<domain>.md` for expected behavior and tier gating.
- Cite requirement IDs in code comments and commits (e.g. "implements `REC-2`").
- New product/business rules go to `prd/`, never inline in code or this file. Engineering/how decisions go to `docs/decisions/` (ADRs). Operational task state goes to `tasks/`.

## Progress tracking

Work state lives in `tasks/` so any agent on any machine can resume cold.

- `tasks/progress.md` — dashboard: Project State snapshot + one-line task index + planned backlog + durable gotchas. Keep ~1 screen; no task detail.
- `tasks/<id>-<slug>.md` — one self-contained handoff file per active/deferred task. Copy `tasks/_TEMPLATE.md` to start.

Rules:
- Session start: read `tasks/progress.md`, then the relevant task file. Don't touch another task's file.
- Keep the task file's "Current state" + checklist accurate; bump `Updated` and the dashboard `Last updated`.
- Lifecycle: planned → active (copy template to `tasks/<id>-<slug>.md`, add dashboard row, next free `T-NNN`) → shipped (fold outcome into Project State as 1–2 lines, write a `docs/decisions/` ADR if durable, then delete the task file + remove its row).
- One file per task; concurrent tasks = concurrent files.
- Durable architecture decisions → `docs/decisions/` ADRs; only task-transient choices live in the task file.
- No history: dashboard holds current + planned + deferred only. Prune stale notes.

## Reference documents

Read the relevant doc before working in its area.

- `prd/` — product requirements, business decisions, free vs paid breakdown
- `docs/architecture.md` — services and boundaries
- `docs/database-schema.md` — schema + rationale
- `docs/privacy-policy.md` · `docs/terms-of-service.md`
- `docs/decisions/` — ADRs

## Scaffolding

`apps/web` and `apps/admin` are SolidStart 2 apps (`bun create solid`: TypeScript, SSR on, names `@qp/web` / `@qp/admin`, tsconfig extends `tsconfig.base.json`). To re-scaffold from scratch, repeat that and `bun install` at the root.
