# PRD · Playback

Per-ayah sequential playback across enabled Languages, seek modes, repeat modes, progress, multi-device sync, and player-level volume.

Status: draft
Related: [playlists.md](./playlists.md), [recitation.md](./recitation.md), [accounts.md](./accounts.md), [collaboration.md](./collaboration.md)

## Requirements

### Sequence model

- **PB-1** — Playback iterates the Playlist's item list in order. For each Ayah covered by an item, the player plays the Ayah's AudioFile in each enabled Language in `language_stack` order before advancing to the next Ayah. _(both)_
- **PB-2** — Arabic, when enabled, plays first for each Ayah (PL-7). _(both)_
- **PB-3** — Per-Ayah audio source is `AudioFile(reciter_id_for(surah, language), surah, ayah)` (PL-10). _(both)_
- **PB-4** — The per-Language volume_pct from the Playlist (PL-17) scales the language's track gain. _(both)_

### Seek modes

- **PB-5** — Two seek modes are supported, selectable via a UI toggle near the seek bar:
  - **Surah-seek (big seek)**: seek bar maps to the Playlist's full logical timeline (sum of all Ayah AudioFile durations across enabled Languages and items). Dragging lands on the nearest Ayah boundary. Small playback pauses during the jump are acceptable. _(both)_
  - **Within-ayah seek**: seek bar maps to the currently playing Ayah's currently-playing Language audio file. Dragging scrubs continuously to any millisecond offset inside that file. _(both)_
- **PB-6** — The Surah-seek logical timeline is computed from per-Ayah `duration_ms` metadata (REC-5); no file probing required. _(both)_
- **PB-7** — Switching between seek modes does not affect playback state (current Ayah, current Language, current offset). _(both)_

### Repeat modes

- **PB-8** — Repeat-current-ayah: when enabled, after the last enabled Language finishes for the current Ayah, playback loops back to the first enabled Language of the SAME Ayah indefinitely (until user disables or hits next). _(both)_
- **PB-9** — Repeat-playlist: when enabled, after the last Ayah of the last item finishes, playback restarts from the first Ayah of the first item. _(both)_
- **PB-10** — Repeat-current-ayah and Repeat-playlist are mutually exclusive: enabling one disables the other (single repeat-mode selector with three states: Off / Repeat Ayah / Repeat Playlist). _(both)_

### Progress

- **PB-11** — Resume position per (User, Playlist) is `(item_index, surah, ayah, within_ayah_offset_ms)`. _(both)_
- **PB-12** — On "resume where you left off", playback starts from that position using the Playlist's currently saved language stack, Reciter selections, and volume mix. No frozen UI state. _(both)_
- **PB-13** — Progress updates are written at most once per 5 seconds while playback is active to limit DB churn. Pause writes immediately. _(both)_

### Multi-device

- **PB-14** — A User may play a Playlist on multiple devices simultaneously. No active-session lock. _(both)_
- **PB-15** — Progress writes are **last-write-wins**: the most recent write wins; older writes are dropped. _(both)_
- **PB-16** — A device that is not actively playing does NOT write progress; it shows the synced position from the server. _(both)_

### Player volume

- **PB-17** — The global player volume (0–100, default 100) is a device-local persistent setting (like YouTube). Stored in browser localStorage / native preferences, not in the server-side User record. _(both)_
- **PB-18** — Effective output for any track = `playlist.volume_pct[lang] × global_player_volume`. _(both)_
- **PB-19** — The global player volume is NOT synced across devices. Each device remembers its own. _(both)_

## Rules / invariants

- Playback always advances by full Ayah audio files; no mid-Ayah skipping to next Ayah (except via user-initiated seek).
- The per-Ayah Language order is fully determined by `playlist.language_stack`.
- The two seek modes are independent UI affordances; they share the same underlying playback state.
- Progress writes never race-overwrite a newer write: server validates `updated_at` ≤ now.
- Multi-device writes do NOT push notifications to inactive devices; inactive devices fetch on demand.

## Acceptance

- A playlist with Arabic + Bengali enabled and items `[Surah 1]` plays Ayah 1 Arabic, then Ayah 1 Bengali, then Ayah 2 Arabic, then Ayah 2 Bengali, etc.
- Disabling Arabic in the language_stack and playing the playlist plays only Bengali audio per Ayah.
- Toggling repeat to "Repeat Ayah" mid-playback causes the current Ayah to loop forever until disabled.
- Toggling repeat to "Repeat Playlist" makes the playlist restart from item 1 Ayah 1 after the last item's last Ayah finishes.
- Dragging the seek bar in Surah-seek mode lands precisely on the start of the target Ayah; within-ayah offset is reset to 0 for the first Language of that Ayah.
- Dragging the seek bar in within-ayah seek mode jumps the current Ayah's current Language file to the chosen ms offset; the next Languages and Ayahs continue as normal.
- On phone playback, the laptop opens the playlist and shows the latest position; laptop starts playing from there; phone position is now superseded by laptop progress writes.
- Setting the global player volume to 70% persists; the next playback starts at 70%.

## Open questions

- Background playback on mobile / lock-screen controls — engineering decision, not a PRD-level toggle.
- Buffering policy / preload-next-Ayah — implementation detail in `docs/decisions/`.
- Crossfade between Ayahs / Languages — out of v1.
