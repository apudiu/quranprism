# Progress Dashboard

Live index of all work in this repo — single source of truth for "where are we / what's next".
Rules + lifecycle: see `CLAUDE.md` → **Progress tracking**. Keep terse. No historical log: shipped work folds into Project State and its per-task file is deleted.

_Last updated: 2026-05-28 · apu@laptop_

## Project State

Current working reality of the repo (what is built + committed unless noted).

- **Initial scaffold** (`e202d0c` on `origin/main`) — polyglot monorepo mirroring ytlistkeeper; container/volume prefix `qp`; compose dev stack healthy locally.
- **v1 PRDs locked** (`8b6eac2` on `origin/main`) — 14 per-domain PRD files in `prd/`; tier matrix + hard constraints canonical in root `CLAUDE.md`.
- **API spine + User & Auth** (`cf32070`, `a61d9b6` on `origin/dev`) — fx-composed modular monolith; platform layer (config/logger/db/cache/pubsub/queue/mailer/jwt) + transport (envelope, typed-error map, middleware: request_id/slog/CORS/recoverer/redis-Lua ratelimit/AuthRequired/EmailVerifiedRequired); `acl`/`user`/`auth` modules; `/v1/auth/*` + `/v1/me*` live; HttpOnly refresh cookie scoped to `/v1/auth`; ACL boot seed (22 perms, 3 system groups) idempotent; ADR-0001 covers the durable decisions; per-workspace `.env.{dev,stg,live}` adopted.
- **Docs site** — mkdocs-material via `docker compose --profile docs up docs` → http://localhost:8000.

## Active Tasks

Copy `_TEMPLATE.md` → `tasks/<id>-<slug>.md` to start; assign next free `T-NNN`.

| ID | Task | Status | Owner/Machine | File |
|----|------|--------|---------------|------|
| _(none)_ | | | | |

## Planned (backlog)

Ordered. Promote to Active when started.

- [ ] Full ACL middleware (`RequirePermission("Subject:action")`) + admin endpoints for groups/permissions CRUD
- [ ] River workers: data-export, post-grace user hard-delete, login-attempt prune cron
- [ ] Notifications module (in-app inbox + SES email, wickmeet coalescing pattern)
- [ ] NATS-WS gateway client-facing wiring (consumed by notifications inbox / sync sessions)
- [ ] Content ingest pipeline (Surahs, Ayahs, Bismillah/Sajdah/Juz, translation text, audio metadata) from Quran.com Uthmani Hafs
- [ ] Admin API: catalog ops (Language/Translator/Reciter CRUD), audio upload pipeline, audit log
- [ ] User API: Playlists (per-Surah Reciter selection, language stack, volume mix)
- [ ] Playback API: progress (Surah:Ayah + offset), repeat modes, multi-device LWW
- [ ] Bookmarks + Categories API
- [ ] Sharing: direct user-to-user (Free), public link + profile (Paid), comments + access requests
- [ ] Collaboration (Paid): roles, transient sync sessions, real-time transport over NATS-WS
- [ ] Web app listening UI + admin app catalog UI
- [ ] i18n framework + English/Bengali UI strings
- [ ] k8s deploy wiring (`deploy/k8s/`)

## Deferred

| Task | Why | Resume trigger | File |
|------|-----|----------------|------|
| _(none)_ | | | |

## Notes / Gotchas

Durable cross-task warnings. Prune when obsolete.

- pg18 image stores data in `/var/lib/postgresql/<major>/` — volume must mount `/var/lib/postgresql`, NOT `/var/lib/postgresql/data` (init aborts otherwise).
- SolidStart 2 alpha SSR returns `[object Object]` under Bun runtime → web/admin dev images run vite on real Node; Bun is install-only.
- Changing `bun.lock` requires `docker compose down -v && docker compose up --build` to re-seed the node_modules named volumes.
- Each new `pkg/<name>` needs a `COPY pkg/<name>/package.json …` line added to BOTH `apps/web/Dockerfile-dev` and `apps/admin/Dockerfile-dev`.
- `bunfig.toml` must NOT set `[install.lockfile] print = "yarn"` (regenerates `yarn.lock`, violates Bun-only rule).
- air writes builds to `apps/api/tmp/`. Go build artifacts (`apps/api/api`, `apps/api/tmp/`) are gitignored — never commit them.
- Docs live-reload: `squidfunk/mkdocs-material:9` ships a broken `serve` watcher. The docs image is built from `python:3.12-alpine` instead (`docs/Dockerfile-dev`); `prd/`/`tasks/` reload via `--watch` since mkdocs won't follow their symlinks.
- k8s api probes hit `/hc` — keep `cmd/api`'s health route named `/hc`.
- **fx graph cycle on shared route prefix**: `chi.Mount` panics if two modules `r.Route("/v1/me", ...)` on the same prefix. Modules sharing a prefix must use flat path mounts (`r.With(mw).Post("/v1/me/change-password", ...)`) instead.
- **sqlc `bool` from boolean expression**: `SELECT x IS NOT NULL AS v` gets inferred as `interface{}` — cast: `SELECT (x IS NOT NULL)::BOOLEAN AS v` to lock the type.
- **`RouteRegistrar` import cycle**: keep the type in a leaf package (`internal/transport/http/router`) — never in `internal/app` (which imports modules). Modules import `router`; app imports `router`; neither depends on the other.
- **River + NATS division of labour**: jobs (atomically enqueued in same `pgx.Tx` as the DB write) → River; pub/sub fan-out + realtime → NATS JetStream + NATS WebSocket. Never the reverse — moving jobs onto NATS forces a transactional-outbox table and an idempotency envelope we don't otherwise need at our scale.
- **Env vars per workspace, never inline in compose**: each `apps/<name>/` owns `.env.dev` / `.env.stg` / `.env.live`. compose `env_file:` references them; postgres shares `apps/api/.env.dev` (DATABASE_URL encodes the same creds). Adding a new env key = edit the three `apps/<workspace>/.env.*` files, not `docker-compose.yml`.
