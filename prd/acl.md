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
- **ACL-6** — Seed groups for v1 (created at migration time, mutable post-deploy via admin UI):
  - **Default user** — every new signup is automatically added. Permissions: `Playlist:*` (own), `Bookmark:*` (own), `BookmarkCategory:*` (own), `Comment:create+delete` (own), `Catalog:view` (read-only: languages, translators, reciters, surahs, ayahs).
  - **Super Admin** — full permission set. Manually granted by another Super Admin or via migration seed for the bootstrap admin.
  - **Content Manager** — catalog management permissions: Language/Translator/Reciter create/update/disable, audio upload, translation text bulk import. No user-management permissions. _(both)_
- **ACL-7** — Admin UI for groups: a user with `Group:*` and `Permission:view` permissions can create/edit Groups, manage Permissions assigned to Groups, and manage user-to-group memberships. _(both)_
- **ACL-8** — A user may be added to multiple Groups; permissions stack additively. Removing a Group revokes only the permissions unique to that Group. _(both)_
- **ACL-9** — Group `Default user` cannot be deleted or have its permission set reduced below the minimum end-user surface (enforced in admin UI). _(both)_
- **ACL-10** — Audit log: every Group/Permission mutation and every user-group membership change is logged with the actor's user_id, action, target IDs, and timestamp. _(both)_

## Rules / invariants

- Authorization is **permission-based**, not role-based. Code never checks Group name.
- Every signup is added to `Default user` at registration time. Removing them from `Default user` is allowed (used for moderation) and effectively locks their personal data behind admin-only access.
- Permission `subject` and `name` columns are case-sensitive and use PascalCase resource names + snake_case actions (e.g., `Playlist:create`). Conventions enforced at seed.
- A user can be a Super Admin AND a normal user simultaneously — they just have more Groups.
- A user with zero Groups has zero permissions; they cannot read catalog, cannot create anything, cannot delete their own account (the deletion endpoint requires `Account:delete_own`).

## Acceptance

- A new signup is automatically in `Default user` and can create a playlist; cannot create a Language.
- Adding the user to `Content Manager` immediately lets them create a Language (permission re-check on next request; no need to log in again).
- Removing the user from `Content Manager` immediately revokes that capability.
- A user with no Groups can sign in but every personal-data endpoint returns 403.
- An admin without `Group:delete` cannot delete a Group.
- An attempt to delete the `Default user` Group returns 422 with an explanatory error.
- The audit log shows a row for every Group membership change with actor, target user, timestamp.

## Open questions

- Per-resource ownership scoping (e.g., `Playlist:update` only on Playlists where `owner_id = current_user`) — implementation pattern via policy functions; design lives in `docs/decisions/`.
- Permission grouping in the admin UI (collapse all `Playlist:*` permissions under a header) — UX polish detail.
- API key / service account for bulk content ingest — defer to v2.
