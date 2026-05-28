# PRD · Accounts

User authentication, profile, and account lifecycle.

Status: draft
Related: [acl.md](./acl.md), [privacy.md](./privacy.md), [notifications.md](./notifications.md)

## Requirements

- **ACC-1** — Sign-up: email + password. No OAuth, no magic link, no phone in v1. _(both)_
- **ACC-2** — Email verification is required at sign-up. The user receives a verification link; until clicked, the account is in `unverified` state with limited capabilities (can view content, cannot save personal data like playlists/bookmarks/notes). _(both)_
- **ACC-3** — Password reset via email link. Reset token is single-use, expires within 1 hour. _(both)_
- **ACC-4** — Passwords stored with bcrypt (cost ≥ 12). Minimum 8 characters at registration; no maximum. No breached-password check in v1. _(both)_
- **ACC-5** — Sessions use server-issued JWT access tokens (~15 min TTL) plus refresh tokens stored in HTTP-only cookies (~30 day TTL). Refresh is single-flight on the client. _(both)_
- **ACC-6** — User profile fields: display name, email (unique, lowercased on store), optional avatar URL, preferred UI language (default `en`, also `bn` in v1 per UI i18n). _(both)_
- **ACC-7** — Account disable (`is_disabled`) is a soft state: disabled accounts cannot sign in but their data is retained. Used by admins for moderation. _(both)_
- **ACC-8** — Account deletion request: user-initiated deletion is honored within 30 days (see [privacy.md](./privacy.md)). _(both)_
- **ACC-9** — Login lockout: after 10 consecutive failed sign-in attempts for the same email within 15 minutes, further attempts return 429 for 15 minutes from that email. _(both)_
- **ACC-10** — Email change: requires confirmation at both the old and new addresses; the change is committed only after both confirmations. _(both)_

## Rules / invariants

- Email is the unique login identifier. Lowercased + trimmed on store and on lookup.
- Passwords never logged (logs scrub the field). Password verification compares server-side via bcrypt.
- Refresh token cookies are `HttpOnly; Secure; SameSite=Lax` (or `Strict` if compatible with redirect flows).
- Verification and reset tokens are cryptographically random, single-use, persisted as a hash (never plaintext).
- An `unverified` account cannot create playlists, bookmarks, or comments. It CAN browse content (Surahs, Ayahs, listening).
- Disabled accounts retain their data; their playlists become inaccessible to others (treat as if owner = null for sharing purposes).

## Acceptance

- `POST /auth/register` with a valid email + password returns 201 and triggers a verification email.
- Following the verification link flips the account to `verified`; the user can now create playlists.
- `POST /auth/login` with wrong password 10 times in a row produces a 429 on the 11th attempt within 15 min.
- `POST /auth/password-reset/request` always returns 200 (don't leak whether the email exists). A reset email is sent only if the email exists.
- Refresh-token rotation: each `POST /auth/refresh` invalidates the old refresh and issues a new one.
- `DELETE /me` initiates the 30-day deletion countdown; user can cancel within the window by signing in.
- Disabled users get 403 on `POST /auth/login` with a specific error code (`ACCOUNT_DISABLED`).

## Open questions

- 2FA / TOTP — defer to v2.
- Social login (Google/Apple) — defer to v2; can be added without breaking ACC-1.
- Phone-based auth — out of scope for v1 and probably v2.
- Email-change cooldown (rate limit) — defer to engineering decision.
