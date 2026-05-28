# T-001 · API architecture lock + User & Auth module

Status: in review
Owner/Machine: apu@laptop
Started: 2026-05-28 · Updated: 2026-05-28
Branch: main (local; not pushed) · Dashboard: [progress.md](./progress.md)

## Goal

Lock the API spine every future module will sit on (DI, queue, pub/sub,
realtime, cache, auth crypto, transport conventions) and land the first
domain module — User + Auth (signup → verify → login → refresh → logout →
forgot/reset/change) per `prd/accounts.md` — as the reference template.

## Context

Scaffold is in place (`e202d0c`); v1 PRDs are locked (`8b6eac2`). We
need the architecture decision to be made once, up front, because
retrofitting across 15 modules is the expensive mistake. Decisions are
durable so they live in `docs/decisions/0001-di-fx-modular-monolith.md`;
only task-transient notes stay in this file.

## Plan / Checklist

- [x] Compose stack: add `redis`, `nats` (JetStream + WS), `mailpit`; `deploy/nats/dev.conf`; api `depends_on` extended.
- [x] go.mod: fx v1.24, pgx v5.9, River v0.38, NATS 1.52, go-redis v9, jwt v5.3, bcrypt, goose v3.27, uuid, go-mail, env/v11.
- [x] `internal/platform/{config,logger,db,cache,pubsub,queue,mailer,jwt}` — one fx.Module each.
- [x] `internal/transport/http/{response,httperr,middleware,router}` — envelope, typed-error mapping, logger / CORS / redis-Lua ratelimit / JWT auth / email-verified middleware, leaf `router.Registrar` to break the import cycle.
- [x] 6 goose migrations: `users` (citext + pgcrypto), ACL tables, verify tokens, reset tokens, refresh sessions, login attempts.
- [x] sqlc queries (`internal/db/queries/*.sql`) → generated (`internal/db/sqlc/`).
- [x] Module `acl` (scaffold): 22-permission catalog, 3 system groups, idempotent boot-time seed; `JoinGroupByName` / `ListGroupsForUser` exposed for auth/user.
- [x] Module `user`: service + repo + handler, `/v1/me` (GET/PATCH/DELETE/data-export).
- [x] Module `auth`: bcrypt12 + random/sha256 tokens + signup/verify/login/refresh/logout/forgot/reset/change handlers + HttpOnly refresh-cookie scoped to `/v1/auth`.
- [x] cmd entrypoints: `cmd/api` (`fx.New(app.HTTPApp).Run()`), `cmd/worker`, `cmd/cron`, `cmd/river-migrate`.
- [x] Smoke test end-to-end via docker compose (signup → mailpit → verify → login → /me → refresh → lockout → logout → forgot-password).
- [x] `go test -race -short ./...` green; `golangci-lint run ./...` 0 issues.
- [x] ADR-0001 written; mkdocs nav + CLAUDE.md pinned versions updated.
- [ ] User accepts the work → commit → fold into Project State as 1–2 lines → delete this file + remove dashboard row.

## Current state

All scope items implemented and verified end-to-end on the local
compose stack. Branch is `main`, ahead 0 of origin (commits `e202d0c`
scaffold + `8b6eac2` PRDs are pushed; this work is unstaged on top).

**Next action**: hand the diff to the user for review. On accept, stage
+ commit + fold into Project State (1–2 lines), then delete this file.

## Decisions (task-local)

Durable decisions live in `docs/decisions/0001-di-fx-modular-monolith.md`.
Task-local choices captured here:

- Refresh cookie is `Path=/v1/auth` only, not site-wide — minimises the
  surface where it's transmitted. JWT goes in the response body so the
  SolidStart client controls header attachment.
- ACC-6 lockout enforced twice: redis-Lua rate-limit middleware (defence
  in depth, returns 429) **and** service-level `CountRecentFailedLogins`
  query (PRD intent, returns 423). Smoke test shows the middleware
  fires first under the same IP — the service-level check covers the
  case where a distributed attacker rotates IPs.
- `EmailVerifier` is a small interface defined in `middleware/auth.go`
  so `transport/http/middleware` doesn't import any domain module;
  user.Service implements it and fx injects it where needed.
- Change-password lives on `auth.Handler` mounted as a flat path
  (`r.With(mw).Post("/v1/me/change-password", ...)`) so it doesn't
  collide with the user module's `r.Route("/v1/me", ...)` block.

## Blockers / open questions

- none — ready for review.

## Verification

```
cd /home/apu/Projects/quranprism && docker compose up -d
# expected: postgres, redis, nats, mailpit, api all healthy

cd apps/api && go test -race -short ./...
# expected: ok across all packages

PATH=$PATH:$(go env GOPATH)/bin golangci-lint run ./...
# expected: 0 issues.

# Signup → verify → login → /me → refresh → logout
curl -X POST http://localhost:3000/v1/auth/signup -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"correct-horse-battery-staple","name":"Alice","tos_accepted":true}'
# expected: 202 {"data":{"user_id":"...","verification_sent":true}}

# Then read the verification token from http://localhost:8025 (Mailpit),
# POST it to /v1/auth/verify-email, then login + /v1/me as shown in the
# smoke transcript folded into the most recent session.
```

ACL idempotency:
```
docker exec qp-postgres psql -U qp -d qp -tA -c "SELECT COUNT(*) FROM permissions"   # → 22
docker exec qp-postgres psql -U qp -d qp -tA -c "SELECT COUNT(*) FROM groups WHERE is_system" # → 3
docker restart qp-api && sleep 10
# counts unchanged.
```

## Key files

- `docker-compose.yml` + `deploy/nats/dev.conf` — extended dev stack.
- `apps/api/go.mod` — locked deps.
- `apps/api/internal/platform/*/` — one fx.Module each: config, logger, db, cache, pubsub, queue, mailer, jwt.
- `apps/api/internal/transport/http/{response,httperr,middleware,router}/` — uniform envelope + typed-error mapping + cross-cutting middleware + leaf Registrar interface.
- `apps/api/migrations/2026052812*.sql` — 6 goose migrations.
- `apps/api/internal/db/queries/*.sql` + `internal/db/sqlc/` — query layer.
- `apps/api/internal/modules/{acl,user,auth}/` — three domain modules.
- `apps/api/cmd/{api,worker,cron,river-migrate}/main.go` — entrypoints.
- `docs/decisions/0001-di-fx-modular-monolith.md` — durable architecture decisions.
