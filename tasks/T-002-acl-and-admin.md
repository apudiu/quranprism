# T-002 · ACL middleware + admin CRUD + extensible `qp` CLI + audit log

Status: in review
Owner/Machine: apu@laptop
Started: 2026-05-29 · Updated: 2026-05-29
Branch: dev · Dashboard: [progress.md](./progress.md)

## Goal

Land the full ACL enforcement surface: `RequirePermission("resource:action")`
middleware, admin CRUD endpoints for groups/permissions/users/memberships,
audit log for every mutation, and an extensible `qp` CLI that consolidates
the existing entrypoints + provides the first-admin bootstrap command.
Done = an operator can `qp admin:grant` a user, that user can manage
groups/permissions/memberships via JSON API, every mutation lands an
audit row, and the system can never lock itself out of admin access.

## Context

Next backlog item after T-001's API spine landed. Three problems
converge: (1) no permission gate exists yet, (2) audit logging is a v1
hard requirement (ACL-10 / ADM-12 / ADM-14) and must land before any
mutating admin surface, (3) without a bootstrap path the first admin
can't exist after deploy. Locked design choices (see plan file):

- Permission strings are `resource:action`, all lowercase, snake_case
  for multi-word. **No wildcards.**
- **No system groups.** Drop `groups.is_system`, drop all seeded groups,
  drop signup auto-join. PRD ACL-6 and ACL-9 superseded as part of this
  task.
- **Default = open.** Any authenticated user reaches any route without
  `RequirePermission`. Ownership is a service/data-layer concern.
- `/v1/admin/permissions` is read-only.
- Single binary `qp` consolidates `cmd/{api,worker,cron,river-migrate}`
  as `qp serve:api|worker|cron`, `qp migrate:up|down|status|queue`,
  `qp admin:grant --email=… [--group=Admin]`.
- Membership endpoints are atomic; one audit row per mutation.
- Admin-side user surface in T-002 is list + detail only, both gated by
  `user:view`.
- Pagination is `?limit=&offset=` (defaults 20, max 100); response
  envelope `{"data":{"items":[…],"total":N,"limit":L,"offset":O}}`.
- Self-protection: any mutation that would leave zero users holding
  `group:update` rejects with `409 last_admin_protected`.

Full plan: `~/.claude/plans/now-ultrathink-to-have-lazy-acorn.md`.

## Plan / Checklist

- [x] Phase 1 — Schema + catalog reset (drop is_system, 7-perm catalog, no auto-join)
- [x] Phase 2 — `audit_log` table + `internal/modules/audit/`
- [x] Phase 3 — sqlc admin queries (groups/perms CRUD, count guards, ListUsers)
- [x] Phase 4 — pagination helper + `RequirePermission` middleware (with per-request cache)
- [x] Phase 5 — `acl.Handler` + 13 admin endpoints + audit-in-tx wrapper
- [x] Phase 6 — last-grantor self-protection (3 paths)
- [x] Phase 7 — `cmd/qp/*` cobra root + `admin:grant`; delete old cmd/* entrypoints; Makefile + .air.toml + Dockerfile + compose + CLAUDE.md updates
- [x] Phase 8 — audit validation tests, perm + pagination middleware tests, PRD supersede ACL-6/ACL-9, smoke verify per plan (legacy purge migration added at `20260529120200`)
- [ ] User accepts the work → commit → fold into Project State as 1–2 lines → delete this file + remove dashboard row.

## Current state

All scope items implemented. End-to-end smoke verified on the local
compose stack: signup → verify → `qp admin:grant` → login → list groups
→ create/link/detail/delete second group → three last-admin self-protect
paths all 409 `last_admin_protected` → audit_log table populated (cli
`admin.grant` + user `group.create`/`group_permission.add`/`group.delete`
rows) → bob (zero groups) reaches `/v1/me` but is 403'd at
`/v1/admin/groups` with `insufficient_permissions`. `go test -race
-short ./...` green; `golangci-lint run ./...` 0 issues.

**Next action**: hand the diff to the user for review. On accept, stage
+ commit + fold into Project State (1–2 lines), then delete this file.

## Decisions (task-local)

Durable architecture decisions go to `docs/decisions/`. Task-local
choices captured here:

- The `auth.Service.aclSvc` field stays in the struct after removing
  the auto-join — only the offending call line goes. Avoids touching
  the constructor signature and breaking existing tests.
- `ErrLastAdmin` maps to `httperr.New(409, "last_admin_protected",
  ...)` explicitly, not `httperr.Conflict(...)` — the generic
  `"conflict"` code is too coarse for the front-end to switch on.
- Audit `Record` is invoked inside the same `pgx.Tx` as the mutation
  via a `RecordTx(ctx, qtx, params)` overload. Commits atomically with
  the change; rolls back together on guard rejection.
- `RequirePermission` caches the user's effective perm slice in a
  private `permsKey{}` ctx value so chained `RequirePermission` calls
  on the same request hit the slice, not the DB.
- Pagination convention (limit/offset with `COUNT(*) OVER()` total) is
  set repo-wide here as `internal/transport/http/pagination/parse.go`;
  every future list endpoint reuses it.

## Blockers / open questions

- none

## Verification

```
# Build qp + apply migrations
cd apps/api && make build && ./bin/qp migrate:up && ./bin/qp migrate:queue
docker compose up -d api

# Signup + verify alice (Mailpit at :8025 to fetch the token)
curl -s -X POST http://localhost:3000/v1/auth/signup ...
curl -s -X POST http://localhost:3000/v1/auth/verify-email ...

# Bootstrap admin
docker compose exec api /app/qp admin:grant --email=alice@example.com --group=Admin

# Login + admin endpoints
ACCESS=$(curl -s -X POST http://localhost:3000/v1/auth/login ... | jq -r .data.access_token)
curl -s http://localhost:3000/v1/admin/groups -H "Authorization: Bearer $ACCESS"
# → {"data":{"items":[{"name":"Admin",…}],"total":1,…}}

# Last-admin self-protect
curl -i -X DELETE http://localhost:3000/v1/admin/users/$ALICE/groups/$ADMIN_GID \
  -H "Authorization: Bearer $ACCESS"
# → 409 Conflict {"error":{"code":"last_admin_protected",…}}
```

Plus: `go test -race -short ./...` green; `golangci-lint run ./...`
zero issues.

## Key files

- `apps/api/internal/modules/acl/{service,handler,catalog,seed,dto,errors,module}.go`
- `apps/api/internal/modules/audit/{module,service,types}.go` (new)
- `apps/api/internal/transport/http/middleware/perm.go` (new)
- `apps/api/internal/transport/http/pagination/parse.go` (new)
- `apps/api/internal/db/queries/{acl,users,audit}.sql`
- `apps/api/migrations/20260529120000_drop_groups_is_system.sql` (new)
- `apps/api/migrations/20260529120100_create_audit_log.sql` (new)
- `apps/api/cmd/qp/{main,serve,migrate,admin}.go` (new)
- `apps/api/cmd/{api,worker,cron,river-migrate}/main.go` (delete)
- `apps/api/{Makefile,.air.toml,Dockerfile,CLAUDE.md}`, `docker-compose.yml`, root `CLAUDE.md`
- `prd/acl.md` (ACL-6, ACL-9 supersession)
