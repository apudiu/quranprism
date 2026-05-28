# Progress Dashboard

Live index of all work in this repo — single source of truth for "where are we / what's next".
Rules + lifecycle: see `CLAUDE.md` → **Progress tracking**. Keep terse. No historical log: shipped work folds into Project State and its per-task file is deleted.

_Last updated: 2026-05-28 · apu@laptop (v1 PRDs locked)_

## Project State

Current working reality of the repo (what is built + committed unless noted).

- **Initial scaffold** (committed as `e202d0c` on `origin/main`): polyglot monorepo mirroring the ytlistkeeper layout. Bun workspaces (`@qp/web`, `@qp/admin`, `@qp/ui`, `@qp/http`), Go module `github.com/apudiu/quranprism/api` with three `cmd/` entrypoints (api/worker/cron — only api has a working `main.go` with chi `/hc` route), `pkg/ui` and `pkg/http` carried over with namespace renamed. `docker-compose.yml` boots postgres 18 + api (`:3000`) + web (`:3002`) + admin (`:3001`).
- **Dev stack** (committed): root `docker-compose.yml`, container/volume prefix `qp` (`qp-postgres`, `qp_postgres_data`, etc.). All four services healthy locally.
- **Docs site** (committed): mkdocs-material via `docker compose --profile docs up docs` → http://localhost:8000.
- **v1 PRDs locked**: 14 per-domain PRD files written under `prd/` covering Content, Translation, Recitation, Accounts, ACL, Playlists, Playback, Bookmarks, Sharing, Collaboration, Notifications, Admin, Compliance, Privacy. `prd/README.md` carries the canonical tier matrix and domains index. Root `CLAUDE.md` Hard constraints filled in (no ads, no AI/TTS audio, translator-purity, whole-Surah Reciter coverage, Arabic 1.0× speed lock, Arabic text source locked to Quran.com Uthmani Hafs, GDPR 30-day). No code implementing these requirements yet.

## Active Tasks

Copy `_TEMPLATE.md` → `tasks/<id>-<slug>.md` to start; assign next free `T-NNN`.

| ID | Task | Status | Owner/Machine | File |
|----|------|--------|---------------|------|
| _(none yet)_ | | | | |

## Planned (backlog)

Ordered. Promote to Active when started.

- [ ] DB schema + goose migrations (`apps/api/migrations/`) — driven by PRD entity definitions (Users, Groups, Permissions, Languages, Translators, Reciters, Surahs, Ayahs, AudioFiles, TranslationText, Playlists, Bookmarks, etc.)
- [ ] sqlc query layer (`apps/api/internal/db/`)
- [ ] ACL infrastructure (`internal/acl/`) — Permission/Group/User M:N, seed groups (Default user, Super Admin, Content Manager), middleware
- [ ] Auth (`internal/auth/`) — email+password signup with email verification, JWT + refresh-cookie sessions
- [ ] `cmd/worker` + `cmd/cron` entrypoints + River job queue wiring
- [ ] Quran.com ingest pipeline — Surahs, Ayahs, Bismillah/Sajdah/Juz markers, translation text, audio metadata for popular reciters
- [ ] Admin API: catalog ops (Language/Translator/Reciter CRUD), audio upload pipeline, audit log
- [ ] User API: Playlists, per-Surah Reciter selection, language stack, volume mix, save validation
- [ ] Playback API: progress (Surah:Ayah + offset), repeat modes, multi-device LWW
- [ ] Bookmarks + Categories API
- [ ] Sharing: direct user-to-user (Free), public link + profile (Paid), comment-style suggestions, access requests
- [ ] Notifications: in-app inbox, email via SES, coalescing model (wickmeet pattern)
- [ ] Collaboration (Paid): roles, sync sessions, real-time transport
- [ ] Web app: listening UI with configurable language stack, per-Surah reciter picker, smart-default propagation
- [ ] Admin app: catalog management UI, audio upload UI, user/group management
- [ ] i18n framework + English/Bengali UI strings
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
