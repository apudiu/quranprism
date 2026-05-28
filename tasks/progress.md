# Progress Dashboard

Live index of all work in this repo — single source of truth for "where are we / what's next".
Rules + lifecycle: see `CLAUDE.md` → **Progress tracking**. Keep terse. No historical log: shipped work folds into Project State and its per-task file is deleted.

_Last updated: 2026-05-28 · apu@laptop (initial scaffold)_

## Project State

Current working reality of the repo (what is built + committed unless noted).

- **Initial scaffold** (staged, NOT committed yet): polyglot monorepo mirroring the ytlistkeeper layout. Bun workspaces (`@qp/web`, `@qp/admin`, `@qp/ui`, `@qp/http`), Go module `github.com/apudiu/quranprism/api` with three `cmd/` entrypoints (api/worker/cron — only api has a working `main.go` with chi `/hc` route), `pkg/ui` and `pkg/http` carried over with namespace renamed. `docker-compose.yml` boots postgres 18 + api (`:3000`) + web (`:3002`) + admin (`:3001`). Admin app has a blank SolidStart shell only. No DB, auth, business logic, or product requirements defined yet.
- **Dev stack** (staged): root `docker-compose.yml` mirrors ytlistkeeper's setup. Container/volume prefix `qp` (`qp-postgres`, `qp_postgres_data`, etc.).
- **Docs site** (staged): mkdocs-material via `docker compose --profile docs up docs` → http://localhost:8000.

## Active Tasks

Copy `_TEMPLATE.md` → `tasks/<id>-<slug>.md` to start; assign next free `T-NNN`.

| ID | Task | Status | Owner/Machine | File |
|----|------|--------|---------------|------|
| _(none yet)_ | | | | |

## Planned (backlog)

Ordered. Promote to Active when started.

- [ ] Define hard constraints in root `CLAUDE.md` (Quran text accuracy, translation/recitation licensing, GDPR)
- [ ] Write initial `prd/` domain files (recitation, translation, user prefs, subscription, privacy, compliance)
- [ ] DB schema + goose migrations (`apps/api/migrations/`)
- [ ] sqlc query layer (`apps/api/internal/db/`)
- [ ] `cmd/worker` + `cmd/cron` entrypoints
- [ ] River job queue wiring
- [ ] Source/ingest pipeline for Quran text + translation corpora + recitation audio
- [ ] Audio streaming/playback architecture
- [ ] Auth model (anonymous listening vs accounts; OAuth provider selection)
- [ ] Web app: listening UI with configurable translation stack
- [ ] Admin app: content + user management UI
- [ ] k8s deploy wiring (`infra/k8s/`)

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
