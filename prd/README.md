# Product Requirements

What quranprism must do and **why** — the product/business source of truth.

- **HOW** we build it → `docs/decisions/` (engineering ADRs).
- **What's being worked on now** → `tasks/`.

## Vision

quranprism plays Quran recitation interleaved with translations in multiple languages, in the order and combination the listener chooses. One source text refracted into every language a listener wants — verse by verse.

## Tiers (canonical gating table)

The single source for what each plan includes. Each domain file marks its requirements `(Free)`, `(Paid)`, or `(both)`.

| Capability | Free | Paid |
|---|---|---|
| Personal playlists (unlimited) | ✓ | ✓ |
| Personal bookmarks (unlimited) | ✓ | ✓ |
| Bookmark categories + notes | ✓ | ✓ |
| Listening (Arabic + translations, all reciters) | ✓ | ✓ |
| Direct user-to-user playlist share | ✓ | ✓ |
| Public link share (playlists) | — | ✓ |
| Public profile (discoverable shared items) | — | ✓ |
| Collaborative playlist (multi-editor + sync listening) | — | ✓ |
| Offline downloads | — | — (v2) |

**No ads on either tier.** Pricing TBD pre-launch.

## Domains

| Domain | File | IDs | Scope |
|---|---|---|---|
| Content (Surah, Ayah) | [content.md](./content.md) | `CON-*` | Canonical Quran text catalog: Surahs, Ayahs, Bismillah/Sajdah/Juz markers |
| Translation | [translation.md](./translation.md) | `TRA-*` | Languages, Translators, translation text |
| Recitation | [recitation.md](./recitation.md) | `REC-*` | Reciters, per-ayah audio files, coverage rules |
| Accounts | [accounts.md](./accounts.md) | `ACC-*` | Email+password auth, profile, sessions, GDPR |
| ACL | [acl.md](./acl.md) | `ACL-*` | Permission-Group-User authorization model |
| Playlists | [playlists.md](./playlists.md) | `PL-*` | Composition, language stack, per-Surah reciter selection, volume mix |
| Playback | [playback.md](./playback.md) | `PB-*` | Sequential per-ayah, seek modes, repeat, multi-device sync, player volume |
| Bookmarks | [bookmarks.md](./bookmarks.md) | `BM-*` | Ayah bookmarks, multi-category tagging, markdown notes, personal-only |
| Sharing | [sharing.md](./sharing.md) | `SH-*` | Playlist sharing (user-to-user / public link / profile), comments, access requests |
| Collaboration | [collaboration.md](./collaboration.md) | `COL-*` | Paid: multi-editor playlists + sync listening sessions |
| Notifications | [notifications.md](./notifications.md) | `NOT-*` | In-app + email, coalesced events |
| Admin | [admin.md](./admin.md) | `ADM-*` | Catalog ops, audio upload pipeline, audit log |
| Compliance | [compliance.md](./compliance.md) | `CMP-*` | No ads, translator-purity, no AI/TTS, Quran-content integrity rules |
| Privacy | [privacy.md](./privacy.md) | `PRV-*` | GDPR export/delete, password reset, data handling |

## Conventions

- Requirement IDs are **stable** — never renumber. Cite from code comments, task files, and ADRs (e.g. "implements `PL-3`").
- To retire a requirement: strike it (`~~PL-9~~ deprecated`) briefly, then delete.
- Keep each file **minimal** — requirements, not prose. Prune the obsolete.
- Product/business rules live here, not in code or `CLAUDE.md`. Engineering choices go to `docs/decisions/`.
- Add a domain by copying [`_TEMPLATE.md`](./_TEMPLATE.md).
