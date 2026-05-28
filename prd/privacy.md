# PRD · Privacy

GDPR-equivalent export and deletion, password handling, notes storage, minimum data collection.

Status: draft
Related: [accounts.md](./accounts.md), [bookmarks.md](./bookmarks.md), [acl.md](./acl.md), [compliance.md](./compliance.md)

## Requirements

### Data export

- **PRV-1** — A signed-in User may request a data export at any time. The export is delivered as a downloadable archive (JSON + media manifests) within 30 days of request (target: under 24 hours in practice). _(both)_
- **PRV-2** — Export contents include: profile (name, email, preferences), all Playlists owned by the user (items, language stack, Reciter selections, volume mix, names, descriptions), all Bookmarks (with notes and category links), all BookmarkCategories, all Comments authored, all received Notifications (without joining other users' PII beyond actor usernames). _(both)_
- **PRV-3** — Export does NOT include: the Quran catalog itself (shared content), other users' data referenced from comments / shares (only the User's own author rows). _(both)_

### Account deletion

- **PRV-4** — A signed-in User may request account deletion. The account enters a 30-day grace period during which signing in cancels the deletion. _(both)_
- **PRV-5** — After 30 days, the User's personal data is hard-deleted: Playlists, Bookmarks, Categories, Notes, Comments authored, Notifications received, profile record, password hash. The User row may be retained as a tombstone for foreign-key integrity (e.g., `users.email = NULL`, `users.deleted_at = NOW()`). _(both)_
- **PRV-6** — Shared Playlists with that User as Owner: hard-deleted along with the user's data. Recipients lose access; this is the documented consequence of deletion. _(both)_
- **PRV-7** — Comments authored by the deleted User on others' Playlists: the comment row remains but the author identity is anonymized (`author = "[deleted]"`). _(both)_
- **PRV-8** — Audit log entries with `actor_user_id = deleted_user` retain the row but anonymize the actor display name. _(both)_

### Password and session handling

- **PRV-9** — Passwords stored as bcrypt hashes (per ACC-4). Plaintext passwords are never logged, never returned in responses, never stored. _(both)_
- **PRV-10** — Logs scrub fields: `password`, `password_confirmation`, `current_password`, `new_password`, `token`, `refresh_token`, `cookie`, `Authorization` header. _(both)_
- **PRV-11** — Reset / verification tokens are stored as hashes (SHA-256) of the underlying random value; the plaintext token is only ever in the URL emailed to the user. _(both)_
- **PRV-12** — Session cookies are `HttpOnly; Secure; SameSite=Lax` (see ACC-5). _(both)_

### Notes storage

- **PRV-13** — Bookmark notes are stored as plaintext in Postgres in v1. Standard at-rest encryption is provided by the database/disk layer (RDS-managed); no application-layer encryption. _(both)_
- **PRV-14** — Notes are not indexed for search in v1; they are not included in product analytics. _(both)_

### Minimum data collection

- **PRV-15** — The system collects only data needed to operate the product: email, display name, password hash, content the user creates (Playlists, Bookmarks, etc.), session metadata, notification preferences, optional avatar URL. No tracking pixels, no marketing analytics SDKs. _(both)_
- **PRV-16** — IP addresses appear in webserver access logs and are retained per the infrastructure log retention policy (target: 30 days). They are not joined to user records for marketing. _(both)_

### Privacy notice

- **PRV-17** — A public Privacy Policy page exists at `/privacy` (mirrors `docs/privacy-policy.md`) describing the above in user-readable terms. _(both)_
- **PRV-18** — Signup forces explicit acceptance of the Privacy Policy and Terms of Service via a required checkbox. _(both)_

## Rules / invariants

- Export and deletion are User-initiated; admins do not delete users without an explicit verified request (account moderation uses `is_disabled`, not delete).
- A deletion request in progress is reversible up to the moment hard-delete runs.
- Once hard-delete runs, the data is gone (no restoration from backups for individual user data; we don't operate a per-user restore service in v1).
- Logs scrubbing is enforced at the logger level, not at the call site.
- No third-party analytics SDK is added to the codebase in v1.

## Acceptance

- `POST /me/data-export` returns 202 and produces a downloadable archive within 24 hours (link sent by email).
- The export archive contains the user's Playlists, Bookmarks, Categories, Notes; does NOT contain the Quran catalog or other users' data.
- `DELETE /me` starts the 30-day grace period; logging in within the window cancels the deletion.
- After 30 days, a scheduled job hard-deletes the user's personal data; the `users` row is anonymized.
- A Comment authored by a deleted user shows as `"[deleted]"` author name on the Playlist where it was posted.
- `grep -r "password" apps/api/internal/.../logs.*` produces no candidate log line emitting a password.
- `/privacy` page returns 200 and matches `docs/privacy-policy.md` content.

## Open questions

- Per-user data retention preferences (auto-delete activity older than N months) — defer to v2.
- Cross-region data residency — out of scope for v1.
- Cookie consent banner — only required if we ever add tracking; v1 has no tracking, so no banner needed in v1.
