# PRD · Bookmarks

Personal Ayah bookmarks with multi-category tagging and markdown notes. **Strictly personal — no sharing, no public visibility.**

Status: draft
Related: [content.md](./content.md), [accounts.md](./accounts.md), [privacy.md](./privacy.md)

## Requirements

- **BM-1** — A Bookmark belongs to exactly one Owner (the User who created it). Bookmarks are not shareable. _(both)_
- **BM-2** — A Bookmark points to a single (Surah, Ayah). No range bookmarks. _(both)_
- **BM-3** — A User may create multiple Bookmarks on the same (Surah, Ayah). Each has independent categories and note. _(both)_
- **BM-4** — A Bookmark carries an optional markdown note (`note_md`), maximum 2 KB source. Server validates length. _(both)_
- **BM-5** — A Bookmark may have zero or more BookmarkCategory tags (multi-tag, like). Categories are User-scoped (see BM-7). _(both)_
- **BM-6** — Bookmark fields: `id`, `user_id`, `surah`, `ayah`, `note_md`, `created_at`, `updated_at`. _(both)_

### BookmarkCategory

- **BM-7** — A BookmarkCategory is owned by a User. Other Users never see one User's categories. _(both)_
- **BM-8** — Category fields: `id`, `user_id`, `name` (≤ 60 chars, unique per user, case-insensitive), `created_at`. _(both)_
- **BM-9** — Categories are flat (no hierarchy). _(both)_
- **BM-10** — A Bookmark-to-Category linkage is many-to-many via `bookmark_category` (M:N), scoped within the same User. _(both)_
- **BM-11** — Deleting a Category removes all its links to Bookmarks; the Bookmarks themselves remain. _(both)_

### Listing / filtering

- **BM-12** — List `/me/bookmarks` returns the User's bookmarks ordered by `created_at` DESC by default; pagination via cursor. _(both)_
- **BM-13** — Filter by category (`?category=<id>`) returns Bookmarks tagged with that Category. _(both)_
- **BM-14** — Filter by ayah (`?surah=<n>&ayah=<n>`) returns all Bookmarks the User has on that exact (Surah, Ayah). _(both)_

### Note rendering

- **BM-15** — The note is stored as markdown source. Rendering (server- or client-side) sanitizes output: no `<script>`, no iframes, no inline event handlers; safe HTML allowed (formatting, links, lists). _(both)_

## Rules / invariants

- Bookmarks are **not shareable** under any tier. The Bookmark API has no share endpoint.
- The BookmarkCategory API has no share endpoint.
- A Bookmark's `user_id` is immutable. Reassigning ownership is not supported.
- Note size is server-enforced at 2,048 bytes of UTF-8 source.
- Category name uniqueness is enforced per User (`UNIQUE (user_id, lower(name))`).
- A Bookmark's tagged Categories must all belong to the same Owner as the Bookmark.

## Acceptance

- `POST /me/bookmarks {surah:1, ayah:1, note_md:"...", category_ids:[]}` creates a bookmark.
- A second `POST` with the same surah+ayah by the same user creates a SECOND bookmark (allowed by BM-3).
- `POST /me/bookmarks` with `note_md` > 2 KB returns 422.
- `POST /me/bookmark-categories {name:"Patience"}` and then again returns 409 conflict (case-insensitive uniqueness).
- `POST /me/bookmarks` with a `category_ids` belonging to another User returns 422 (cross-user reference).
- Deleting a Category sets `category_ids = []` for all linked Bookmarks; the Bookmarks remain.
- `GET /me/bookmarks?category=<id>` returns only Bookmarks tagged with that Category.
- Bookmark API contains no share or visibility endpoint.

## Open questions

- Bookmark search by note text — out of v1 (notes are not indexed).
- Bookmark export (CSV / JSON) — defer to v2; covered partially by [privacy.md](./privacy.md) GDPR export.
- Visual indicator on the Ayah view when bookmarks exist for the current user — UX polish.
