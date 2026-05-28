# PRD · Playlists

User-created Quran playlists with multi-language stacks and per-Surah reciter selection.

Status: draft
Related: [content.md](./content.md), [recitation.md](./recitation.md), [translation.md](./translation.md), [playback.md](./playback.md), [sharing.md](./sharing.md), [collaboration.md](./collaboration.md)

## Requirements

### Composition

- **PL-1** — A Playlist belongs to exactly one Owner (the User who created it). _(both)_
- **PL-2** — A Playlist has a fixed `granularity`: either `surah` or `ayah`. Set at creation; immutable. _(both)_
- **PL-3** — A surah-level Playlist contains an ordered list of items, each item being either a single Surah ID or a contiguous Surah range. Items may be mixed in one list (e.g., `1, 7–10, 36, 67–114`). _(both)_
- **PL-4** — An ayah-level Playlist contains an ordered list of items, each item being either a single Ayah (Surah:Ayah) or an Ayah range (Surah:Ayah – Surah:Ayah). Ranges MAY span across Surahs (e.g., `2:250–3:10`). _(both)_
- **PL-5** — Item order in the saved list defines playback order; reordering items via the editor reorders playback. _(both)_
- **PL-6** — A Playlist has a `name` (≤ 120 chars), optional `description` (≤ 500 chars), `created_at`, `updated_at`. _(both)_

### Language stack

- **PL-7** — A Playlist has a `language_stack`: an ordered list of enabled Languages. Arabic, when enabled, is always at the top of the stack (UI prevents reordering Arabic below other languages). _(both)_
- **PL-8** — Other Languages in the stack can be reordered via drag-and-drop. Order determines per-ayah audio playback order. _(both)_
- **PL-9** — Arabic MAY be disabled (omitted from the stack). When disabled, the listener hears translation audio only for each ayah. _(both)_

### Per-Surah Reciter selection

- **PL-10** — For every (Surah-touched-by-the-playlist × enabled Language), the Playlist stores a Reciter selection. _(both)_
- **PL-11** — **Smart-default propagation**: the user's first selection of a Reciter for a Language (made in any of the playlist's Surahs) auto-fills that Reciter for all OTHER Surahs in the playlist where the Reciter has audio coverage. _(both)_
- **PL-12** — Surahs where the auto-fill cannot place a Reciter (no coverage from the chosen Reciter) are visually highlighted in the editor so the user can pick a different Reciter for those. _(both)_
- **PL-13** — Per-Surah manual override is allowed. After the smart-default seed, any single-Surah Reciter change stays local to that Surah; it does not re-propagate. _(both)_
- **PL-14** — **Reciter summary list at the top of the editor**: shows distinct Reciters currently in use across the Playlist, grouped by Language (format: `[Language] – [Reciter Name]`). Multiple entries for one Language indicate mixed Reciters across Surahs. _(both)_
- **PL-15** — **Click-to-sync** from the summary list: clicking an entry applies that Reciter to every Surah in the Playlist where the Reciter has coverage for the Language. Surahs without coverage are skipped (still highlighted). _(both)_
- **PL-16** — A Playlist cannot be saved unless every (Surah × enabled Language) cell has a Reciter selected. The save action returns a validation error listing the uncovered cells. _(both)_

### Per-language volume mix

- **PL-17** — A Playlist stores a per-Language `volume_pct` (0–100, default 100) — the relative weight applied to each Language's audio track during playback. _(both)_
- **PL-18** — Volume mix changes are persisted on the Playlist; the volume_pct is not a global user preference. Each new Playlist starts with all enabled Languages at 100. _(both)_
- **PL-19** — Effective output during playback = `playlist.volume_pct[lang] × global_player_volume`. The global player volume is described in [playback.md](./playback.md). _(both)_

## Rules / invariants

- Granularity (`surah` vs `ayah`) is fixed at creation; cannot be switched.
- Arabic, when enabled, is always position 0 in `language_stack`.
- Owner is required and immutable (ownership transfer is a separate explicit action).
- Reciter selection is per (Surah × Language). A single Surah may carry different Reciters for different Languages (Arabic from Reciter A, Bengali from Reciter B — that's normal).
- Translator-purity is automatically enforced because each Reciter is bound to one Translator (see [recitation.md](./recitation.md)).
- Validation on save is exhaustive over (Surah × enabled Language); no partial saves.

## Acceptance

- Creating a `surah` playlist and passing an item like `Surah 2 Ayah 50` is rejected with a granularity mismatch error.
- Creating an `ayah` playlist with item `2:250–3:10` is accepted; the playback queue spans into Surah 3.
- Saving a playlist with Arabic enabled and reordering Arabic to position 1 is rejected by the server.
- Picking Reciter X for Bengali in Surah 1 of a playlist with Surahs 1–10 auto-fills Reciter X for Surahs 1–10 where X has coverage; uncovered Surahs are returned with a `needs_reciter` flag.
- After the auto-fill seed, changing Reciter X → Y in Surah 5 does NOT change Surahs 1–4 or 6–10.
- Clicking a summary-list entry "Bengali – Reciter X" applies X to all Surahs in the playlist where X has Bengali coverage.
- Saving with one uncovered cell returns 422 listing the (Surah, Language) pairs missing a Reciter.

## Open questions

- Playlist max item count — defer; soft cap can be added if needed (target audience won't hit hard limits).
- Playlist cloning / templates — defer to v2.
- Volume mix presets (e.g., "Arabic dominant", "Equal") — UX polish; defer.
