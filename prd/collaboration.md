# PRD · Collaboration (Paid)

Multi-editor playlists with transient sync-listening sessions. **Paid-tier only.**

Status: draft
Related: [playlists.md](./playlists.md), [playback.md](./playback.md), [sharing.md](./sharing.md), [notifications.md](./notifications.md)

## Requirements

### Enabling collaboration

- **COL-1** — A Playlist may be promoted to a Collaborative Playlist by its Owner. This requires the Owner has Paid tier. Demotion back to non-collaborative is allowed and removes all non-Owner roles. _(Paid)_
- **COL-2** — Promoting to Collaborative does not change the Playlist's `id`, items, language stack, or Reciter selections. _(Paid)_

### Roles

- **COL-3** — A Collaborative Playlist has three participant role levels:
  - **Owner** — exactly one, immutable except via explicit ownership transfer (not in v1).
  - **Admin Participant** — promoted by Owner from any Participant; can edit the Playlist and control sync sessions; multiple allowed.
  - **Participant** — default for invitees; can listen along but cannot edit and cannot control sync session playback. _(Paid)_
- **COL-4** — The Owner invites Participants by username or email. Newly added Participants enter as role `Participant`. _(Paid)_
- **COL-5** — The Owner promotes a Participant to Admin via an action in the participant management UI. Owner demotes Admin → Participant the same way. _(Paid)_
- **COL-6** — Owner can remove any Participant or Admin from the Playlist at any time. _(Paid)_

### Edit surface

- **COL-7** — Owner and Admins share the same edit surface: add/remove/reorder Playlist items, change per-(Surah × Language) Reciter selection, adjust per-Language volume mix, edit Playlist name and description. _(Paid)_
- **COL-8** — Participant role cannot edit any field. Attempts return 403. _(Paid)_
- **COL-9** — Edit conflicts use last-write-wins on a per-field basis: the most recent write to `(field, key)` wins. No optimistic locking errors surfaced to UI in v1; concurrent edits may overwrite each other silently. _(Paid)_

### Sync sessions

- **COL-10** — A Sync Session is an ephemeral object representing a "let's listen together" state on a Collaborative Playlist. At most one active session per Playlist at a time. _(Paid)_
- **COL-11** — Sessions are started by an Owner or Admin. Participants and other Admins may join an active session; joining brings them to the session's current playback position. _(Paid)_
- **COL-12** — During an active session:
  - Only Owner and Admins may control playback (play, pause, seek, skip to next item, change language stack mid-session). Control actions broadcast in real time to all session members. _(Paid)_
  - Participants are passive listeners — their player UI shows playback controls as disabled. _(Paid)_
- **COL-13** — A session ends when the host stops it explicitly OR when the last connected member disconnects OR after 10 minutes of all-paused inactivity. _(Paid)_
- **COL-14** — Sync sessions DO NOT mutate any participant's solo-listening resume position (per PB-11). Leaving the session restores the participant to their personal resume position. _(Paid)_
- **COL-15** — Real-time transport for sync sessions is implementation-defined (WebSocket / SSE) and lives in `docs/decisions/`. Not a PRD concern. _(Paid)_

### Notifications

- **COL-16** — Participants are notified (in-app + email) when added to a Collaborative Playlist, when promoted to Admin, when demoted, when removed. See [notifications.md](./notifications.md). _(Paid)_

## Rules / invariants

- Collaborative Playlists exist only at Paid tier. Downgrading the Owner to Free demotes the Playlist (removes all non-Owner roles, ends any active session). Items, Reciter selections, etc., remain.
- Participants and Admins must themselves have an account; collaboration requires authentication for all members (no anonymous join).
- A Playlist's Owner is also implicitly a Participant for session purposes; no separate record needed.
- A Participant cannot also be an Admin or Owner simultaneously — roles are mutually exclusive per Playlist per User.
- Edits during an active sync session take effect immediately and are visible to session members; the session's playback queue updates accordingly.

## Acceptance

- A Free Owner attempting to enable Collaboration on their Playlist gets 403.
- A Paid Owner enables Collaboration; the Playlist is now mutable by all Admins.
- Inviting Bob as a Participant notifies Bob; Bob can view + listen but cannot edit.
- Promoting Bob to Admin lets Bob change a Reciter on Surah 1 — the change is reflected for everyone.
- During a sync session controlled by the Owner, a Participant attempting to pause via API gets 403; the playback continues for all members.
- Two Admins each renaming the Playlist within the same second: the last write wins; both Admins see the final value on their next refresh.
- Ending the sync session restores each participant's UI to their solo resume position; their solo position was not modified during the session.
- Owner Paid → Free downgrade demotes the Playlist back to non-Collaborative; previously invited Participants lose access (notified once).

## Open questions

- Ownership transfer — defer to v2; deletion-and-recreation suffices for v1.
- Optimistic-locking UI ("someone else edited this — refresh") — defer; LWW is acceptable for v1.
- Recorded sync sessions (replay later) — out of scope.
- Voice chat / text chat during a session — out of scope.
