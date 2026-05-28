# PRD · Recitation

Reciters and the audio files they produce. Audio is human-recorded and admin-uploaded; per (Reciter × Surah × Ayah) one file each.

Status: draft
Related: [content.md](./content.md), [translation.md](./translation.md), [playback.md](./playback.md), [admin.md](./admin.md), [compliance.md](./compliance.md)

## Requirements

- **REC-1** — A Reciter is identified by name and bound to exactly one Language. _(both)_
- **REC-2** — If the Reciter's Language is non-Arabic, the Reciter is ALSO bound to exactly one Translator (whose Language matches the Reciter's). The same human person narrating two Translators' texts is modeled as two distinct Reciter records. _(both)_
- **REC-3** — If the Reciter's Language is Arabic, the Reciter has NO Translator. Arabic Reciters recite the source Arabic text directly. _(both)_
- **REC-4** — An AudioFile is per (Reciter × Surah × Ayah): one file per ayah of one Reciter. Stored object key pattern: `audio/{reciter_id}/{surah}/{ayah}.{ext}`. _(both)_
- **REC-5** — At upload time, the audio file's duration (in milliseconds) is captured and stored on the AudioFile record. Duration metadata is the source-of-truth for the playback logical timeline; the player never probes the file. _(both)_
- **REC-6** — **Whole-Surah coverage rule**: a Reciter must have AudioFile rows for ALL Ayahs of any Surah they appear in. Partial-Surah coverage is rejected at publish time. _(both)_
- **REC-7** — A Reciter MAY cover only some Surahs (e.g., Surahs 1–10 only). Surahs they have not recorded are not exposed for selection where that Reciter's coverage is required. _(both)_
- **REC-8** — Reciter `published` state requires at least one fully-covered Surah. Published Reciters with their covered Surah list are exposed to end users. _(both)_
- **REC-9** — Reciters expose, per Surah, an ordered list of Ayah AudioFiles with durations and URLs. Clients use this to build the per-ayah playback queue and the logical seek timeline. _(both)_
- **REC-10** — Re-recording a single Ayah replaces the corresponding AudioFile (object key reuse, duration updated). No history retention in v1. _(both)_
- **REC-11** — Audio format: MP3, mono or stereo, recommended bitrate 64–128 kbps. Maximum file size 5 MB per ayah enforced at upload. _(both)_

## Rules / invariants

- **No AI-generated audio. No TTS.** All audio is human-recorded and human-uploaded. Inviolable.
- A Reciter belongs to exactly one Translator (for non-Arabic) or to Arabic directly (no Translator).
- Translator binding is immutable after creation: changing it would silently violate translator-purity for any playlist using this Reciter.
- A Reciter's Language is immutable after creation.
- Whole-Surah coverage is enforced before any user-facing exposure of a Reciter for that Surah.
- AudioFile duration is required and must be > 0 ms.

## Acceptance

- `POST /admin/reciters` with `language_id = 'ar'` and `translator_id != null` is rejected.
- `POST /admin/reciters` with `language_id = 'bn'` and `translator_id = null` is rejected.
- Uploading the 5th Ayah of Surah 1 for Reciter X with duration_ms = 0 is rejected.
- A Reciter with audio for Ayahs 1–6 of Surah 1 (where Surah 1 has 7 ayahs) cannot mark Surah 1 as covered.
- `GET /reciters/:id/surahs` returns only Surahs with full Ayah coverage.
- `GET /reciters/:id/surahs/:surah/ayahs` returns ordered Ayah AudioFile entries with durations and URLs.
- Replacing an Ayah's audio file updates `duration_ms` on the record and overwrites the object at the same URL.

## Open questions

- Qira'at metadata on Arabic Reciters (Hafs, Warsh, Doori, etc.) — useful for future qira'at filter; defer to v2.
- Multiple recordings per (Reciter, Surah, Ayah) for quality variants — defer; one canonical recording per (Reciter, Surah, Ayah) in v1.
- Streaming protocol (range requests, signed URLs) — engineering decision in `docs/decisions/`.
