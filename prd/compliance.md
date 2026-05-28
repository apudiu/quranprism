# PRD · Compliance

Religious, legal, and ethical constraints. Many of these are inviolable and are also reflected in the **Hard constraints** section of the root `CLAUDE.md`.

Status: draft
Related: [content.md](./content.md), [translation.md](./translation.md), [recitation.md](./recitation.md), [playlists.md](./playlists.md), [privacy.md](./privacy.md)

## Requirements

### Inviolable rules (Hard constraints)

- **CMP-1** — **No ads, ever, anywhere in the app.** This includes marketing pages, profile pages, settings, and any UI surface. Revenue is solely from Paid-tier subscriptions. _(both)_
- **CMP-2** — **No AI-generated or TTS audio anywhere.** Every audio file is human-recorded and human-uploaded by admin. This applies equally to Arabic recitation and all translation narrations. _(both)_
- **CMP-3** — **Translator-purity within a (Surah, Language) cell of any Playlist.** Structurally enforced because each Reciter is bound to exactly one Translator (REC-2). _(both)_
- **CMP-4** — **Whole-Surah Reciter coverage**: a Reciter has audio for ALL Ayahs of any Surah they appear in (REC-6). Partial-Surah is rejected at publish. _(both)_
- **CMP-5** — **Arabic recitation is never time-stretched or pitch-shifted.** When playback-speed control is added in a future release, it applies only to non-Arabic audio tracks; Arabic always plays at 1.0×. This preserves Tajweed pacing integrity. _(both)_
- **CMP-6** — **Arabic Quran text source is fixed**: ingest only from the Quran.com curated Uthmani Hafs dataset. No editing of Arabic text by humans or machines once ingested. _(both)_

### Translation labeling

- **CMP-7** — Every UI surface displaying translation text or playing translation audio shows the Translator's name (e.g., "Saheeh International"). _(both)_
- **CMP-8** — The Translator's source attribution (per TRA-3) is accessible via a one-click info link from the labeled Translator name. _(both)_

### Audio licensing

- **CMP-9** — Each Reciter record carries a license attribute (free-form text or enum: public-domain, permitted-with-attribution, licensed, etc.) populated at creation time. Reciters cannot be published without a license value set. _(both)_
- **CMP-10** — The license value is exposed on the Reciter detail page so users can see provenance. _(both)_

### Surah-specific religious rules

- **CMP-11** — Bismillah handling per CON-5: Surah 1 starts with Bismillah as Ayah 1; Surahs 2–114 except Surah 9 prefix Bismillah; Surah 9 has no Bismillah. UI and API enforce this. _(both)_
- **CMP-12** — Sajdah ayahs are marked in UI (per CON-6) with a small icon so listeners reading along know they have reached a sajdah verse. _(both)_

### Disabled-account handling for Quran content

- **CMP-13** — Disabling a user (per ACC-7) never affects the Quran catalog (Surahs, Ayahs, translation text, audio). The catalog is shared content and is never modified by user moderation. _(both)_

### Anti-abuse on sharing surfaces

- **CMP-14** — Comments on shared Playlists pass through a basic content filter (profanity / spam heuristics). Owner can override the filter to allow specific flagged comments. _(both)_
- **CMP-15** — Public profile and public-link share pages include a "Report this content" link visible to anyone viewing. _(Paid; pages only exist for Paid Owners)_

## Rules / invariants

- The Hard constraints section of `CLAUDE.md` mirrors CMP-1 through CMP-6 verbatim. Engineering reviews use the root document as the canonical reference.
- The compliance PRD itself is the source of truth for any soft constraint not promoted to root `CLAUDE.md`.
- A future tier or feature that conflicts with any CMP-1 to CMP-6 is rejected at design review.

## Acceptance

- A page-level audit (manual or automated) finds zero ad slots / ad scripts / ad SDK references in built artifacts.
- Audio upload endpoint rejects any file whose metadata claims AI-generated origin (rare in MP3 metadata, but if seen, hard-fail).
- A user attempting to disable Arabic at playback level (rather than at playlist level) finds no UI toggle. Playlists with Arabic disabled by the user at creation time still play translation audio without Arabic.
- A playback-speed control (when added) is greyed out / N/A for the Arabic track and active for non-Arabic tracks.
- The UI on any translation page shows the Translator name within 1 viewport-height of the translation text.
- A Reciter created without a license value cannot be published (422 error).
- Surah 9's first Ayah does NOT display a Bismillah; Surah 1's Ayah 1 IS Bismillah.

## Open questions

- Comment moderation toolset for Owners (block list, report-then-hide) — defer to v2.
- Per-region content restrictions (e.g., specific translations restricted in some jurisdictions) — out of scope for v1.
- DMCA / takedown process for user-submitted content — minimal in v1 since users only upload comments (not audio).
