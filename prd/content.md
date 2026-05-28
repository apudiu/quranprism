# PRD · Content (Surah, Ayah)

The canonical Quran text catalog: 114 Surahs, all Ayahs, and their structural metadata. Read-only for end users; managed by admin.

Status: draft
Related: [recitation.md](./recitation.md), [translation.md](./translation.md), [admin.md](./admin.md), [compliance.md](./compliance.md)

## Requirements

Stable IDs — never renumber. Tier tag in parentheses.

- **CON-1** — The catalog contains all 114 Surahs with stable numeric IDs (1–114) matching Mushaf order. _(both)_
- **CON-2** — Each Surah has the following metadata: Arabic name (in Arabic script), English transliteration, English meaning, Makki/Madani classification, revelation order (1–114, distinct from Mushaf order), total ayah count. _(both)_
- **CON-3** — Each Ayah has a stable numeric ID within its Surah, and the canonical Arabic text (Uthmani script, Hafs an Asim qira'at). _(both)_
- **CON-4** — The source of truth for Arabic text is the Quran.com curated Uthmani Hafs dataset. Ingest pipeline imports from there; no in-house Arabic transcription. _(both)_
- **CON-5** — Bismillah handling follows standard Mushaf convention: in Surah 1 (Al-Fatiha), Bismillah is Ayah 1 of the Surah; in Surahs 2–114 except Surah 9 (At-Tawbah), Bismillah is a prefix prepended at the Surah head but is NOT a numbered Ayah; Surah 9 has no Bismillah at all. _(both)_
- **CON-6** — Sajdah ayahs are flagged: the 14 standard sajdah ayahs are marked on the Ayah record so playback UI and reading UI can display a sajdah indicator. _(both)_
- **CON-7** — Each Ayah carries its Juz (1–30) and Hizb (1–60) membership for navigation purposes. _(both)_
- **CON-8** — End-user APIs expose Surah and Ayah read endpoints (list Surahs, get Surah, list Ayahs in Surah, get Ayah). Mutations are admin-only. _(both)_
- **CON-9** — Surah and Ayah IDs are referentially stable across the lifetime of the app. New ingest passes update text/metadata in place but never renumber. _(both)_

## Rules / invariants

- Arabic text is the inviolable source content. **Never** generate Arabic via TTS or model output; ingest only from Quran.com Uthmani Hafs.
- The 114 Surahs are fixed by religion; no admin-created Surahs.
- Surah 9 (At-Tawbah) must never display a Bismillah prefix — UI and API enforce this.
- Sajdah ayah list is fixed by religious tradition; not user-editable.
- Ingest reruns are idempotent: re-importing the same source produces no schema change beyond text or metadata corrections.

## Acceptance

- `GET /surahs` returns 114 entries with the seven metadata fields per CON-2.
- `GET /surahs/1/ayahs/1` returns Bismillah as the first ayah's Arabic text (per CON-5).
- `GET /surahs/9/ayahs/1` returns an ayah that does NOT begin with Bismillah text, and the Surah record indicates `bismillah_prefix: false`.
- `GET /surahs/2/ayahs/1` returns the first numbered ayah of Al-Baqarah; the Surah record indicates `bismillah_prefix: true` and `bismillah_ayah_number: null`.
- All 14 sajdah ayahs return `is_sajdah: true`; all others return false.
- Each Ayah includes its `juz` (1–30) and `hizb` (1–60) fields.
- Re-running the ingest job leaves the catalog byte-identical (no Surah or Ayah ID changes).

## Open questions

- Sajdah types (recommended vs obligatory) — store the type, or just a single flag? Defer to admin domain.
- Exposing Word-by-word (kalimah-level) metadata for highlighting during playback — v2.
