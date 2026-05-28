# PRD · Translation

Languages and Translators that render Quranic Arabic into other tongues. Translation text is exposed to end users; audio for translations lives in [recitation.md](./recitation.md).

Status: draft
Related: [content.md](./content.md), [recitation.md](./recitation.md), [admin.md](./admin.md), [compliance.md](./compliance.md)

## Requirements

- **TRA-1** — A Language is identified by its ISO 639 code (e.g., `en`, `bn`, `ar`), native name (e.g., `বাংলা`), English name (e.g., `Bengali`), and direction (LTR or RTL). _(both)_
- **TRA-2** — Arabic is a Language with the reserved code `ar`. It is the source language and has NO Translators; admin UI prevents adding Translators under it. _(both)_
- **TRA-3** — A Translator is identified by name, belongs to exactly one non-Arabic Language, and carries an optional bio and source attribution (where the translation text originated, e.g., "Saheeh International, 1997"). _(both)_
- **TRA-4** — TranslationText is the per (Translator × Surah × Ayah) text. Stored as plain text (no markdown rendering, no HTML). One row per Ayah per Translator. _(both)_
- **TRA-5** — Translation text is sourced from Quran.com's translation dataset for v1 launch; license terms for each Translator are recorded in the Translator record. _(both)_
- **TRA-6** — Translator completeness: a Translator must have TranslationText for ALL Ayahs of the Quran before they can be marked `published` and exposed to end users. Partial Translators stay in `draft`. _(both)_
- **TRA-7** — End-user APIs expose published Languages and published Translators (and their per-ayah text). Drafts are admin-only. _(both)_
- **TRA-8** — Translator name is shown in UI alongside translation text and audio (e.g., "Saheeh International"), per [compliance.md](./compliance.md) labeling rules. _(both)_

## Rules / invariants

- **Translator purity (structural)**: a Reciter belongs to exactly one Translator (recorded in [recitation.md](./recitation.md)). Picking a Reciter for a playlist language slot implicitly picks the Translator. No mixing of Translators within a single Surah's audio for a language is possible by construction.
- Arabic Language must never have a Translator record.
- A Translator's Language is immutable after creation (the text is bound to a target language).
- Source attribution is required at publish time (license traceability).
- TranslationText is plain UTF-8 text only. No HTML, no markdown, no images.

## Acceptance

- `GET /languages` returns only published Languages plus Arabic.
- `POST /admin/translators` rejects a body with `language_id` referencing Arabic (`ar`).
- A Translator marked `published` is queryable from end-user APIs; a Translator in `draft` is not.
- Marking a Translator `published` fails if any Ayah is missing TranslationText (per CON-1 expects 6,236 ayahs total).
- `GET /translators/:id/text/:surah/:ayah` returns the plain-text translation; rendering as HTML is escaped.
- Translator name is included in the response for any audio file from a Reciter under that Translator.

## Open questions

- Multiple version snapshots per Translator (e.g., 1934 vs 1997 editions) — out of v1 scope; single version per Translator record for now.
- Markdown formatting in translations (footnotes, bold for emphasis) — defer.
