# ADR-0001: fx-composed modular monolith for the Go API

Date: 2026-05-28
Status: accepted

## Context

The Go API will grow into ~15 domain modules (auth, user, ACL, content,
translation, recitation, playlists, playback, bookmarks, sharing,
collaboration, notifications, admin, …). We need a composition model
that:

1. Feels like NestJS (per-module bundles of providers + handlers + lifecycle).
2. Resolves dependencies + lifecycle ordering without per-module boilerplate.
3. Detects circular dependencies loudly at startup, not silently at runtime.
4. Doesn't lock us in if we later split a domain into its own service.
5. Stays idiomatic Go — no codegen-heavy magic for a single-team service.

We also need to decide which transport / queue / cache backs the platform.

## Options considered

### DI / composition

| Option | Why considered | Outcome |
|---|---|---|
| **Uber fx v1.24** | NestJS-like `fx.Module(...)`, runtime cycle detection with traces, `group:"routes"` for self-registering handlers, lifecycle hooks. | **Chosen.** |
| Google wire | Compile-time DI, no runtime cost. | **Eliminated** — archived by Google on 2025-08-25. |
| Manual wiring in `cmd/api/main.go` | No framework. | Eliminated — clean at ≤5 modules, ugly at 15 with lifecycle ordering. |

### Job queue

| Option | Why considered | Outcome |
|---|---|---|
| **River v0.38** (Postgres-backed) | `InsertTx` enqueues atomically inside the same `pgx.Tx` as the DB write. No outbox needed. | **Chosen.** |
| NATS JetStream as queue | One broker for everything. | Eliminated — forces a transactional outbox (extra table, dispatcher, idempotency envelope) to recover the atomicity River gives for free. Trendy ≠ cheaper at our scale. |

### Pub/sub + realtime

| Option | Why considered | Outcome |
|---|---|---|
| **NATS JetStream + NATS WebSocket** | One broker for cross-replica fan-out and direct browser subscribe. Sub-200ms for collab sync. | **Chosen.** |
| Centrifugo | Drop-in dedicated WS server, presence + history. | Escape hatch — adopt if NATS subject auth proves too coarse. |
| Self-rolled WS | Cheapest in $. | Eliminated — re-implements Centrifugo features we'd otherwise get for free from NATS. |
| Soketi | Pusher-compatible OSS. | Eliminated — effectively dormant since March 2024. |
| SSE-only | ~40% less memory at scale. | Tactical fallback for the notifications inbox; doesn't cover full-duplex collab. |

### Cache, JWT

- **Redis 7**: rate-limit counters, hot lookups, effective-permission cache later.
- **HS256 with rotating secret list** (`JWT_SECRETS=current,previous`): verifier tries each, prepend on rotate. RS256 is overkill for a monolith.
- **`github.com/golang-jwt/jwt/v5`** is the current incumbent (last major fork after `dgrijalva` died).

## Decision

Use Uber fx as the DI framework. Each domain module ships as a
self-contained package under `internal/modules/<name>/` exporting an
`fx.Module` that wires its `Repository → Service → Handler`. Handlers
self-register on the chi router by emitting a `router.Registrar` into
the fx value group `group:"routes"`, which the router builder in
`internal/app/http.go` consumes.

Cross-cutting infrastructure (DB, cache, queue, pub/sub, mailer, JWT,
logger, config) lives in `internal/platform/<name>/`. Platform packages
never import `internal/modules/*`, so the dependency arrow stays
one-way and cycles are impossible by construction.

Three cmd entrypoints share the same fx graph minus the HTTP listener:

- `cmd/api/main.go` runs `fx.New(app.HTTPApp).Run()`.
- `cmd/worker/main.go` runs `fx.New(app.WorkerApp).Run()` — same platform + domains, no HTTP server, workers added later.
- `cmd/cron/main.go` mirrors worker for periodic tasks.

Job queue is River (Postgres-backed). Pub/sub + realtime is NATS
JetStream + NATS WebSocket. Cache is Redis. JWT is HS256 with rotating
secrets.

## Consequences

**Positive**

- Adding a new module = one new directory + one line in `internal/app/module.go`. No edits to existing modules; cycle detection at startup catches mistakes.
- DB-atomic enqueue stays — we never paint ourselves into a transactional-outbox corner.
- One broker (NATS) covers fan-out and realtime; we don't run Centrifugo unless we need it.
- The platform interface is stable: future modules see config / logger / db / cache / pubsub / queue / mailer / jwt with no surprises.

**Negative**

- fx reflection cost is paid once at boot (~ms). Not a hot-path concern, but it's there.
- Each module touches `router.Registrar` and the chi router via fx value groups — slightly more ceremony than a single central route table, but it scales with the module count without forcing one file to know everyone.
- Bumping the React-of-DI (fx) is now a coordinated event. Pin in `CLAUDE.md` and bump deliberately.

**Neutral**

- If we ever split a domain into its own service, the module's `fx.Module` is the unit we lift out — the boundaries already match.

## References

- `internal/app/module.go` — `Platform`, `Domains`, `HTTPApp`, `WorkerApp`, `CronApp` composition.
- `internal/transport/http/router/registrar.go` — the leaf `Registrar` interface that breaks the import cycle.
- `internal/modules/auth/module.go` — canonical example of a domain `fx.Module`.
- `prd/accounts.md`, `prd/acl.md` — the requirements the User+Auth module implements.
