# PRD · Notifications

In-app + email notifications, coalesced per-entity. Mirrors the notification model in `~/Projects/wickmeet/apps/api/src/modules/notification/`.

Status: draft
Related: [sharing.md](./sharing.md), [collaboration.md](./collaboration.md), [accounts.md](./accounts.md)

## Requirements

### Channels

- **NOT-1** — v1 ships two notification channels:
  - **In-app notification center**: a bell icon with unread count + a notifications inbox page. _(both)_
  - **Email** via AWS SES. _(both)_
- **NOT-2** — Web push / mobile push deferred to v2. _(both)_

### Notification model

- **NOT-3** — A Notification row has: `id`, `recipient_user_id`, `type` (enum), `entity_type` (string), `entity_id` (foreign id), `actor_user_id` (nullable; null for system notifications), `actor_count` (integer ≥ 1), `actors` (JSONB array of capped recent actor snapshots), `data` (JSONB free-form payload), `seen_at` (nullable), `read_at` (nullable), `created_at`, `updated_at`. _(both)_
- **NOT-4** — **Coalescing**: the first Notification for a given `(recipient_user_id, type, entity_type, entity_id)` is INSERTed. Every subsequent event for the same key UPDATEs that row: `actor_count` increments, the new actor is prepended to `actors` (capped at 5 most-recent), `updated_at` bumps, `read_at` resets to NULL. _(both)_
- **NOT-5** — `seen_at` is set when the recipient opens the notification center (bell popover or inbox page). `read_at` is set when the recipient explicitly marks read or clicks through to the linked entity. _(both)_

### Listing / inbox

- **NOT-6** — `GET /me/notifications` lists notifications ordered by `updated_at` DESC. Cursor-based pagination (opaque base64url cursor). Page size 1–100, default 20. _(both)_
- **NOT-7** — Filter `?filter=unread` returns only rows with `read_at IS NULL`. Default (omitted) returns everything. _(both)_
- **NOT-8** — Unread count: `GET /me/notifications/unread-count` returns the count of `read_at IS NULL` rows for the bell badge. Capped at 99+ display in UI. _(both)_
- **NOT-9** — `POST /me/notifications/:id/read` and `POST /me/notifications/mark-all-read` mutate state. _(both)_

### Event types (v1)

- **NOT-10** — Events that produce notifications in v1:
  - `playlist_shared_to_you` — emitted when a User is added as a direct-share recipient of a Playlist.
  - `playlist_access_requested` — emitted to the Playlist Owner when a non-recipient requests access.
  - `playlist_commented` — emitted to the Playlist Owner when a Recipient leaves a comment.
  - `comment_resolved` — emitted to the Comment author when the Owner marks their comment resolved or dismissed.
  - `collab_added` — emitted when a User is invited to a Collaborative Playlist.
  - `collab_role_changed` — emitted when an Admin is promoted, demoted, or removed.
  - `collab_session_started` — emitted to Participants when an Owner/Admin starts a sync session. _(both)_

### Email delivery

- **NOT-11** — Each notification type has a corresponding email template. Email is sent via AWS SES. _(both)_
- **NOT-12** — Per-user email preferences: a settings page lets the User toggle email-on/off per event type. Defaults: all email-on at sign-up. _(both)_
- **NOT-13** — Email is throttled: at most one email per `(recipient, type, entity)` per hour. Subsequent events of the same kind to the same entity within the throttle window are coalesced in-app only (NOT-4 applies; the email is suppressed). _(both)_
- **NOT-14** — Email contains a one-click unsubscribe link for that event type. Clicking unsubscribe sets the per-user, per-type email preference to off. _(both)_

## Rules / invariants

- Coalescing key is `(recipient_user_id, type, entity_type, entity_id)`. Notifications NOT NULL on `entity_type` + `entity_id` — Postgres treats NULLs as distinct, which would silently let duplicates through.
- `actors` JSONB array is bounded at 5; older actors are dropped from the snapshot but `actor_count` keeps the total.
- A system event (no actor) has `actor_user_id = NULL` and a special `data` field describing the system action.
- Self-notifications are not produced: the recipient should never equal the actor; if they do, the event is dropped silently.
- Emails NEVER include the user's password or any auth token in the body.
- Notifications survive playlist deletion: `entity_id` may dangle; the UI handles missing-entity gracefully.

## Acceptance

- Alice shares Playlist P with Bob → Bob's bell badge shows +1 unread, and Bob receives an email (subject like "Alice shared a playlist with you").
- Alice shares Playlist P with Bob then Carol within an hour: Bob has a single notification with `actor_count = 2`, `actors = ["Alice", "Carol"]`. Carol gets her own first-time notification.
- Bob opens the bell popover → all visible notifications have `seen_at` set; `read_at` is still null until Bob clicks through.
- Bob clicks the notification → `read_at` is set, bell badge decrements.
- Bob disables email for `playlist_shared_to_you` in settings → Carol shares another playlist with Bob; Bob's in-app notification fires but no email is sent.
- Throttle: Carol shares 10 playlists with Bob in 10 minutes → Bob receives one email, with the 10 events coalesced in-app over time.
- Unsubscribe link in an email: clicking turns off Bob's `playlist_shared_to_you` email preference; no further emails of that type.

## Open questions

- Web push (after notification permission grant) — defer to v2.
- Mobile push (iOS / Android) — depends on native apps which are post-v1.
- Digest emails (daily summary) — defer to v2.
- Notification preferences default state at sign-up — confirmed all-on; can revisit if email complaints accumulate.
