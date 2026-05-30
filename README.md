# quranprism

A multilingual Quran listening service. Plays Quran recitation interleaved with user-configurable translations — Arabic → English → Spanish → Bangla, etc., verse by verse.

## Prerequisites

- Go 1.26.0+
- Bun 1.3.14+ (pinned in `.bun-version`)
- Docker (for local Postgres 18.4)
- PostgreSQL client tools (psql) for ad hoc DB work

## Project layout

This is a polyglot monorepo. Bun manages the TypeScript apps in `apps/web/` and `apps/admin/`. The Go module lives at `apps/api/` and is independent of the Bun workspace.

See `CLAUDE.md` for the full conventions document.

## First-time setup

1. Clone the repository.
2. Install Bun if you haven't already: `curl -fsSL https://bun.sh/install | bash`
3. Verify version: `bun --version` should print 1.3.14 or newer.
4. Install root JS deps: `bun install`
5. Initialize the Go module deps:
   - `cd apps/api && go mod tidy`

## Daily development

```
docker compose up -d                                 # full dev stack (postgres + api + web + admin)
cd apps/api && make run-api                          # API server
cd apps/api && make run-worker                       # background worker (separate terminal)
bun run dev:web                                      # web frontend (separate terminal)
bun run dev:admin                                    # admin frontend (separate terminal)
```

## Troubleshooting

**Added a JS dependency but the dev container can't find it** (e.g. `Cannot find module '@tanstack/solid-query'`).

The `web`/`admin` containers keep their installed `node_modules` in **anonymous** volumes that mask the bind-mounted host tree. Rebuild and recreate to re-seed them from the fresh image:

```
docker compose down
docker compose up -d --build web admin
```

`down` removes the containers and orphans their anonymous `node_modules` volumes, so the rebuilt image's dependencies are seeded into brand-new ones. (Tidy up the orphaned volumes occasionally with `docker volume prune`.) A plain `up --build` *without* `down` first won't pick up new deps — Compose preserves the old volume across an in-place recreate.

## Documentation

- `CLAUDE.md` — Repository conventions, version pins, hard constraints
- `docs/architecture.md` — Service boundaries and deployment topology
- `prd/` — Product requirements
- `docs/database-schema.md` — Schema with rationale
- `docs/privacy-policy.md` and `docs/terms-of-service.md` — Legal documents
- `docs/decisions/` — Architecture decision records

- M-1 — move GetUserByEmail inside the admin:grant tx (close TOCTOU race)
- M-3 — align admin:grant timeout from 30s to 60s
- L-2 — --email format check (reject input without @)
- L-3 — .air.toml watch coverage for the root migrations.go
- L-4 — flip SilenceUsage: false so flag errors include a help hint

Out of scope entirely:
- k8s manifests (already deferred by T-002 plan).
- L-5 (idempotent admin:grant description-clobber).
- L-6 (worker/cron compose services).
- L-7 (binary-path standardisation across dev/container/prod).
- Per-mode binary stripping via build tags.
- Test refactor to inject a fake *sqlcdb.Queries for unit-testing acl.Service and audit.Service.Record directly.
- Restructuring colon-namespaced cobra commands into space-separated nested commands.
