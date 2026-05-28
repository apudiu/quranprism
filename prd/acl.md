# PRD · ACL (Authorization)

Permission-based authorization. Permissions belong to Groups; Users belong to Groups; effective permission set is the union across the User's Groups.

Mirrors the schema and model used by `~/Projects/order-online/apps/api`.

Status: draft
Related: [accounts.md](./accounts.md), [admin.md](./admin.md)

## Requirements

- **ACL-1** — Tables: `users`, `permissions`, `groups`, `group_user` (M:N), `group_permission` (M:N). One `users` table for everyone; there is no separate `admins` entity. _(both)_
- **ACL-2** — Permission record: `(id, name, subject, description, ...)`. `subject` is the resource (e.g., `Language`, `Reciter`, `Playlist`); `name` is the action (e.g., `create`, `update`, `delete`, `view`). One Permission row = one (subject, action) atom. _(both)_
- **ACL-3** — Group record: `(id, name, description, ...)`. A Group bundles a set of Permissions via `group_permission`. _(both)_
- **ACL-4** — A User belongs to zero or more Groups via `group_user`. Effective permission set = union of Permissions across all the User's Groups. _(both)_
- **ACL-5** — Authorization checks at every API boundary use the user's effective permission set. Code checks permissions by (subject, action), never by "role name" or Group name. _(both)_
- **ACL-6** — _Superseded by T-002 (2026-05-29)_: no system groups are seeded. Admins create groups post-deploy via the `qp` CLI bootstrap (`qp admin:grant --email=… [--group=Admin]`). The seeded permission catalog is the authoritative list; group composition is operator policy. _(both)_
- **ACL-7** — Admin endpoints under `/v1/admin/*` are gated by atomic perms: `group:create`, `group:view`, `group:update`, `group:delete`, `permission:view`, `user:view`. A user with `group:update` can manage memberships and permission assignments. _(both)_
- **ACL-8** — A user may be added to multiple Groups; permissions stack additively. Removing a Group revokes only the permissions unique to that Group. _(both)_
- **ACL-9** — _Superseded by T-002 (2026-05-29)_: self-protection now applies at the permission level — any mutation that would leave zero users holding `group:update` is rejected with `409 last_admin_protected`. Last-admin recovery is the `qp admin:grant` CLI command. _(both)_
- **ACL-10** — Audit log: every Group/Permission mutation and every user-group membership change is logged with the actor's user_id, action, target IDs, and timestamp. _(both)_

## Rules / invariants

- Authorization is **permission-based**, not role-based. Code never checks Group name.
- _Superseded by T-002 (2026-05-29)_: no system groups, no auto-join. New signups belong to zero groups. **Default = open**: any authenticated user can reach any route that lacks `RequirePermission` middleware (catalog browsing, playlist/bookmark/category CRUD on their own data — gated by ownership at the data layer, not by permissions). Restricted operations (admin endpoints, future catalog ops) carry an explicit `RequirePermission("resource:action")`.
- Permission names are `resource:action`, lowercase, snake_case for multi-word resources (e.g. `group:create`, `audit_log:view`, `audio_file:upload`). No wildcards. Conventions enforced at seed.
- A user can hold every admin perm AND remain a normal user — they just have more Groups.
- A user with zero Groups has zero permissions; they still reach every unrestricted route (per the default-open rule).

## Acceptance

- A new signup has zero groups but can still reach unrestricted routes (catalog, own playlists/bookmarks).
- After `qp admin:grant`, the user can hit `/v1/admin/groups` and manage other users / groups / permissions.
- Adding a user to a group containing a perm immediately lets them use the gated endpoint on their next request (no re-login).
- Removing the user from the granting group immediately revokes that capability.
- An admin without `group:delete` cannot delete a Group (403 `insufficient_permissions`).
- An attempt to revoke the last user holding `group:update` returns 409 with code `last_admin_protected`.
- The audit log shows a row for every Group / Permission link / membership mutation with actor, action, target IDs, timestamp.

## Open questions

- Per-resource ownership scoping (e.g., `Playlist:update` only on Playlists where `owner_id = current_user`) — implementation pattern via policy functions; design lives in `docs/decisions/`.
- Permission grouping in the admin UI (collapse all `Playlist:*` permissions under a header) — UX polish detail.
- API key / service account for bulk content ingest — defer to v2.
