# PRD · Sharing

Playlist sharing modes, comment-style suggestions, and access-request flow. **Only Playlists are shareable.** Bookmarks are personal-only.

Status: draft
Related: [playlists.md](./playlists.md), [collaboration.md](./collaboration.md), [notifications.md](./notifications.md), [bookmarks.md](./bookmarks.md)

## Requirements

### Sharing modes

- **SH-1** — A Playlist supports up to three concurrent sharing modes; the Owner controls each independently. _(both)_
  - **Direct user-to-user share**: Owner adds specific Users (by username or email) to a Playlist's recipient list. Only those Users can view. _(Free + Paid)_
  - **Public link share**: Owner generates a public URL. Anyone with the URL can view. _(Paid only)_
  - **Public profile**: shared Playlists may be marked discoverable from the Owner's public profile. _(Paid only)_
- **SH-2** — Direct user-to-user share recipients are notified via in-app and email (see [notifications.md](./notifications.md)). _(both)_
- **SH-3** — Public link tokens are unguessable (≥ 128 bits of entropy). Owner can revoke a public link, which invalidates it permanently. _(Paid)_
- **SH-4** — Owner can remove individual direct-share recipients at any time. _(both)_

### Access permission

- **SH-5** — Recipients have **read-only** access: they may listen to the Playlist using its current items, Reciter selections, and volume mix. They may NOT edit the Playlist. _(both)_
- **SH-6** — Recipients have their own per-user resume position on the shared Playlist (independent of the Owner's progress). _(both)_
- **SH-7** — Recipients may leave Comments on the Playlist (see suggestion lifecycle below). _(both)_
- **SH-8** — Recipients may NOT re-share the Playlist via the platform's share UI. _(both)_

### Suggestion lifecycle — comment-style

- **SH-9** — A Recipient with view access may post a Comment on a shared Playlist. Comments are free-form markdown text up to 2 KB. Recipients use comments to request changes (e.g., "please use Reciter Y for Surah 2"). _(both)_
- **SH-10** — The Owner sees comments in the Playlist editor and is notified per [notifications.md](./notifications.md). _(both)_
- **SH-11** — The Owner may edit the Playlist freely in response (or ignore the comment). There is no structured "change object" or diff-apply mechanism in v1. _(both)_
- **SH-12** — The Owner may mark a Comment as **resolved** or **dismissed**. The Comment author is notified of the state change. _(both)_
- **SH-13** — A Comment author may edit or delete their own Comment until the Owner marks it resolved. After resolved, comments are read-only history. _(both)_
- **SH-14** — Comments are visible to: the Owner, all current direct-share recipients, and anyone viewing via public link / public profile. The Comment author's username is shown alongside the comment. _(both)_

### Access requests (out-of-band URL sharing)

- **SH-15** — Personal URL sharing (e.g., Recipient copies the URL and sends it via WhatsApp to a non-recipient) is outside the platform's control. When a non-authorized User opens the URL, they see a **"Request access"** prompt. _(both)_
- **SH-16** — Submitting an access request notifies the Owner (in-app + email). _(both)_
- **SH-17** — The Owner can approve (granting direct-share access to that User) or deny the request from the notification or from a dedicated access-requests page. _(both)_
- **SH-18** — Denial is silent: the requester sees no specific denial notification, only an unchanged "no access" state. _(both)_
- **SH-19** — Public-link-shared Playlists skip the access-request flow: anyone with the URL has view access directly. _(Paid)_

### Public profile

- **SH-20** — A Paid User has a public profile page at a stable URL (`/u/<username>`). The page lists Playlists the User has chosen to mark as discoverable. _(Paid)_
- **SH-21** — A Playlist is discoverable from the public profile only when the Owner has explicitly enabled it. Discoverability is independent of the public-link toggle (but practically often co-enabled). _(Paid)_
- **SH-22** — Free Users have no public profile page; their `/u/<username>` route returns 404. _(Free)_

## Rules / invariants

- A Playlist's Owner cannot be a recipient (Owner has implicit full access).
- Removing a recipient revokes their resume-position record on that Playlist (or marks it dormant — not deleted in v1, but inaccessible).
- A revoked public link can never be reactivated; the Owner must generate a new link with a new token.
- Bookmarks have no sharing endpoint at any tier.
- Direct-share recipient lists have no fixed limit in v1; soft limit may be added later for abuse prevention.

## Acceptance

- `POST /playlists/:id/share/users {emails:["alice@…"]}` adds Alice; Alice's `/me/shared` returns the playlist.
- `POST /playlists/:id/share/link` returns a URL like `/p/<token>` only if the Owner has Paid tier; Free users get 403.
- `GET /p/<valid-token>` returns the Playlist read-only.
- `GET /playlists/:id` by a non-Owner non-recipient returns a "Request access" hint (403 with that error code), NOT 404.
- `POST /playlists/:id/access-requests` from a non-recipient notifies the Owner.
- Owner approves the access request → requester is now in the direct-share recipient list and can `GET /playlists/:id`.
- A Recipient `POST /playlists/:id/comments {body_md:"..."}` succeeds; Owner is notified.
- Comment author can `PATCH` their own comment until Owner marks `state: resolved`; after that, edit returns 422.
- `GET /u/<paid-user>` returns a profile page listing discoverable Playlists; `GET /u/<free-user>` returns 404.
- Public link `DELETE` invalidates the token immediately; subsequent `GET /p/<old-token>` returns 410 Gone.

## Open questions

- Comment thread replies (nested) — defer to v2; v1 is flat.
- Per-comment reactions (likes) — defer.
- Block list / mute (Owner blocks a specific User from commenting) — defer to v2.
