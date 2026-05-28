# PRD · Admin

Catalog management, audio upload pipeline, translation text ingest, and audit log. Permission-gated via [acl.md](./acl.md).

Status: draft
Related: [content.md](./content.md), [translation.md](./translation.md), [recitation.md](./recitation.md), [acl.md](./acl.md)

## Requirements

### Catalog ops — Languages

- **ADM-1** — Create / update / disable Languages: name (native + English), ISO 639 code, direction (LTR / RTL). Arabic (`ar`) is seeded at migration time and cannot be deleted. Permissions: `Language:create`, `Language:update`, `Language:disable`. _(both)_
- **ADM-2** — Disabling a Language hides it from end users (Playlists referring to it continue to function for those that already include it; new selections of that Language are blocked). _(both)_

### Catalog ops — Translators

- **ADM-3** — Create / update / disable Translators: name, language_id (non-Arabic only), optional bio, source attribution. Permissions: `Translator:create`, `Translator:update`, `Translator:disable`. _(both)_
- **ADM-4** — Translator state: `draft` → `published`. Only Translators with TranslationText for ALL Ayahs (6,236 total) may be published. The publish action runs a coverage check. _(both)_
- **ADM-5** — Bulk import of TranslationText from a CSV / JSON file uploaded by the admin. The importer is idempotent: re-uploading the same source updates rows in place. Permissions: `Translator:import_text`. _(both)_

### Catalog ops — Reciters

- **ADM-6** — Create / update / disable Reciters: name, language_id, translator_id (required for non-Arabic, forbidden for Arabic). Permissions: `Reciter:create`, `Reciter:update`, `Reciter:disable`. _(both)_
- **ADM-7** — Reciter state: `draft` → `published`. Publish requires at least one Surah with complete coverage (REC-6 enforces full-Surah coverage). _(both)_

### Audio upload pipeline

- **ADM-8** — Per-ayah audio upload endpoint accepts an MP3 file plus metadata `(reciter_id, surah, ayah)`. Server validates size (≤ 5 MB), MIME type, and decodes briefly to extract `duration_ms`. Permission: `AudioFile:upload`. _(both)_
- **ADM-9** — Bulk upload for a full Surah: admin uploads a zip of `ayah-001.mp3, ayah-002.mp3, …` named in canonical order; server validates names match the Surah's Ayah count and processes each as ADM-8. _(both)_
- **ADM-10** — Audio file storage: object key `audio/{reciter_id}/{surah}/{ayah}.mp3`. Re-upload of the same (reciter, surah, ayah) replaces the existing object and updates `duration_ms`. No history. _(both)_
- **ADM-11** — A Reciter's Surah coverage transitions to `complete` automatically when all Ayahs of that Surah have uploaded files. _(both)_

### Audit log

- **ADM-12** — Every admin mutation writes an Audit Log row: `actor_user_id`, `action` (string, e.g. `Reciter.create`), `subject_type` + `subject_id`, `changes` (JSONB diff), `timestamp`. Append-only; no deletion. _(both)_
- **ADM-13** — Audit log is browsable by Super Admins. Permission: `AuditLog:view`. _(both)_
- **ADM-14** — Audit log entries for sensitive subjects (Groups, Permissions, User memberships) are also written by the ACL flows ([acl.md](./acl.md) ACL-10). The schema is shared. _(both)_

### Admin UI permissions

- **ADM-15** — The admin app (`apps/admin`) shows nav entries based on the user's effective permissions. A user without any admin permission sees nothing in the admin app and is redirected to the user web app. _(both)_

## Rules / invariants

- All admin mutations require both authentication AND the specific permission for the action. No "is_admin" boolean shortcut.
- Catalog entities (Language, Translator, Reciter) are soft-disabled, not deleted. Deletion would cascade to AudioFile / TranslationText; we don't lose source content.
- Re-uploads of audio replace in place — no history retention in v1. Restoring an older recording requires re-uploading from a backup.
- Coverage transitions (`incomplete` → `complete`) for a Reciter's Surah are computed by the server on each upload/delete event, never trusted from clients.
- Audit log is append-only at the DB level (no UPDATE / DELETE grants on the table).

## Acceptance

- A user without `Reciter:create` calling `POST /admin/reciters` gets 403.
- A user with `Reciter:create` posting valid body succeeds; the Reciter is in `draft` state.
- Publishing a Reciter who covers Surahs 1–10 fully succeeds; publishing one with only Ayahs 1–6 of Surah 1 fails with a coverage error.
- Uploading audio for (Reciter X, Surah 1, Ayah 1) creates an AudioFile row with `duration_ms` extracted from the file.
- Re-uploading the same (Reciter X, Surah 1, Ayah 1) overwrites the object and updates `duration_ms`.
- Uploading a 6 MB MP3 returns 422 with a size error.
- Uploading audio for an invalid Ayah number for the Surah is rejected (e.g., Surah 1 only has 7 ayahs; uploading for Ayah 8 fails).
- Audit log shows a row for each create/update/disable; the `changes` JSONB contains the field diff.
- A Super Admin browsing `/admin/audit-log` sees recent admin actions across the catalog.

## Open questions

- Multi-file batch upload UX for translation text (CSV layout) — defer to engineering decision.
- Audio waveform preview in the admin UI before publish — UX polish; defer.
- Soft-delete vs soft-disable on Translator — current decision is disable only.
- Bulk import audit-log granularity (one row per item or one row per batch) — defer to engineering decision.
