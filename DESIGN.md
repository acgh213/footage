# ✧ Footage — design document ✧

## architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Footage (GUI application)                    │
│                                                                 │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────────────┐  │
│  │  session    │  │   tag panel  │  │     region list       │  │
│  │  (files)    │  │  (presets,   │  │  (timeline of marked  │  │
│  │             │  │   hotkeys)   │  │   segments, live)     │  │
│  └─────────────┘  └──────────────┘  └───────────────────────┘  │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  transport bar  ▶ ⏸ ⏮ 5s ⏭ 30s  1x 1.5x 2x 4x            ││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                 │
│                         │                                       │
│              IPC (mpv JSON API)                                │
│                         │                                       │
│              ┌──────────────────────┐                           │
│              │  mpv                 │                           │
│              │  (external window,   │                           │
│              │   bundled)           │                           │
│              └──────────────────────┘                           │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │               Python backend (footage.py)                 │   │
│  │  ┌──────────┐  ┌────────────┐  ┌────────────────────┐   │   │
│  │  │ whisper  │  │  vision    │  │  LLM (query engine) │   │   │
│  │  │ (audio)  │  │ (frames)   │  │                     │   │   │
│  │  └──────────┘  └────────────┘  └────────────────────┘   │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

**GUI** handles the logging deck: session management, tag presets, hotkey input, region editing, notes, transport control. **Python backend** handles heavy lifting: whisper transcription, vision model annotation, LLM queries. Communication via NDJSON subprocess — same pattern as the screenshot cataloger and chisel.

## player integration

Footage doesn't embed a video player. It controls one externally via IPC.

### mpv (primary)

mpv exposes a JSON IPC API over a named pipe or socket. Footage sends commands and reads playback position:

```
# Windows named pipe
\\.\pipe\footage-mpv

# Commands
{"command": ["set_property", "speed", 2.0]}
{"command": ["seek", -5, "relative"]}
{"command": ["get_property", "time-position"]}
```

Benefits:
- Open source, fast, handles any format
- Scriptable IPC, no HTTP overhead
- Bundled with Footage — zero setup

## data model

### session

A session is a *focus set* — a group of video files you're working through. Files aren't moved or copied. The session just tracks which files are in the set and what's been logged.

```
my-session/
├── session.json          # session metadata: name, created, files[]
├── manifest.jsonl        # tagged regions and bookmarks
└── exports/              # remuxed clips (created on batch export)
```

### manifest format

See **[docs/schema.md](docs/schema.md)** for the full manifest schema. Key points:

- **Stable ULIDs** for regions, bookmarks, sessions, and videos
- **Seconds as source of truth** — `start_sec` and `end_sec` are authoritative. Display timestamps are derived sugar
- **Append-only during logging, rewritten on edits** — no event log in v0.1
- **Regions overlap.** Flat storage, UI handles nesting display
- **Bookmarks are `kind: "bookmark"`** — a single timestamp, no region, no tag

### tag presets

Presets are global, stored in `%APPDATA%/Footage/presets/`. Each preset is a JSON file:

```json
{
  "name": "destiny-raid",
  "tags": [
    {"key": "1", "label": "boss encounter", "color": "#e8919e"},
    {"key": "2", "label": "traversal", "color": "#a8c8e8"},
    {"key": "3", "label": "adds clear", "color": "#8cc4a0"},
    {"key": "4", "label": "wipe", "color": "#f0a68c"},
    {"key": "5", "label": "loot / chest", "color": "#dbb87c"},
    {"key": "6", "label": "menu / loadout", "color": "#b8a0d4"},
    {"key": "7", "label": "cutscene", "color": "#c47a66"},
    {"key": "0", "label": "bookmark", "color": "#8a7d8a"}
  ],
  "auto_notes": true
}
```

`auto_notes` controls whether a notes field appears automatically when closing a tag region.

## GUI design

### layout

```
┌────────────────────────────────────────────────────────────────┐
│  Footage — destiny2_raid_2021.mp4                    — □ ✕     │
├────────────┬─────────────────────────┬─────────────────────────┤
│  SESSION   │                         │  REGIONS                │
│            │      TAG PRESET         │                         │
│  📁 raid   │  ┌───────────────────┐  │  00:03:12 — 00:05:47    │
│   🎬 oryx │  │ desting-raid      │  │  🔴 boss encounter       │
│   🎬 gate │  │                   │  │  first Oryx clear        │
│  📁 pvp   │  │ 1  boss encounter │  │                         │
│   🎬 ib   │  │ 2  traversal      │  │  00:08:30 — 00:10:15    │
│   🎬 tri │  │ 3  adds clear     │  │  🔵 traversal             │
│            │  │ 4  wipe           │  │                         │
│            │  │ 5  loot / chest   │  │  00:12:00 — 00:14:45    │
│            │  │ 6  menu / loadout │  │  🟢 adds clear            │
│            │  │ 7  cutscene       │  │                         │
│            │  │ 0  bookmark       │  │  ▶ 00:16:22 — ...       │
│            │  │                   │  │  🟡 bookmark             │
│            │  └───────────────────┘  │                         │
│            │                         │                         │
│            │  NOTES                  │                         │
│            │  ┌───────────────────┐  │                         │
│            │  │ (active or last)  │  │                         │
│            │  │                   │  │                         │
│            │  └───────────────────┘  │                         │
├────────────┴─────────────────────────┴─────────────────────────┤
│  ▶ ⏸  ⏮5s  ⏭30s   1x 1.5x 2x 4x   ████████████░░░  12:34     │
└────────────────────────────────────────────────────────────────┘
```

### interaction model

1. **Load a session** — drag video files in, or open a folder. They appear in the session panel
2. **Pick a preset** — dropdown of your saved tag groups. The hotkeys load
3. **Play** — mpv launches and starts playing. Footage syncs playback position via IPC
4. **Tag as you watch** — tap `1` when a boss encounter starts, tap `1` again when it ends. The region appears in the list
5. **Notes** — when you close a tag region, the notes field focuses for quick annotation
6. **Jump around** — transport bar controls mpv. Seek back 5 seconds, skip ahead 30
7. **Variable speed** — 1.5x, 2x, 4x for fast-forwarding through traversal sections
8. **Bookmarks** — tap `0` for "come back to this." One tap, no region to close

### hotkey precision

- **one open region per tag per video.** pressing `1` opens a boss encounter region. pressing `1` again closes it. pressing `2` while `1` is open opens a traversal region — both can be open simultaneously, but only one of each tag type
- **bookmarks are separate.** the bookmark hotkey creates `kind: "bookmark"` entries. they are not regions with missing end times
- **closing a region auto-focuses notes.** when you close a tag region, the notes field receives focus for quick annotation

### LLM pop-up

`Ctrl+L` opens a pop-up overlay, not a sidebar. It's an interaction, not a persistent chat. The LLM can:

- Search the manifest ("show me every wipe from the Oryx encounter")
- Jump the video ("go to the first boss encounter in this clip")
- Answer questions ("what was I doing at 14:30 in this video?")
- Drive the UI — the LLM returns commands, not just text

The pop-up closes when you dismiss it. It doesn't linger. You summon it for a task, get the result, move on.

## backend processing

### audio description transcription

Run whisper locally on tagged regions. Not full dialogue transcription — context-level:
- "Two speakers, one is giving instructions, gunfire in background"
- Speaker identification for multi-speaker segments
- Key phrases extracted, not full text

This runs as a batch job after logging — select regions, queue whisper, get descriptions back into the manifest.

### screenshot annotation

Extract frames from tagged regions (start, middle, end + key frames) and feed them through the same vision model pipeline as the screenshot cataloger. Annotated frames link back to their parent region in the manifest.

### LLM query engine

The Python backend loads the manifest and provides a query interface. The LLM (local, same infrastructure as chisel) can:
- Execute structured searches against the manifest
- Generate human-readable answers from manifest data
- Return UI commands (seek, load, filter)

## decisions

- **GUI, not TUI.** Windows state sync between TUI and external video player is unreliable. A GUI can embed or IPC-sync with mpv cleanly
- **mpv as sole player.** JSON IPC is fast, scriptable, and bundled. No VLC fallback — one player, one integration path
- **Tag presets are global.** Saved in AppData, available to all sessions. No per-project preset management
- **Regions are flat, overlap is allowed.** The UI handles nesting display. The manifest doesn't enforce hierarchy
- **Batch export, not live clipping.** Logging and exporting are separate passes. Log first, export later
- **LLM is a pop-up, not a sidebar.** It's summoned for a task, not a persistent conversation partner
- **JSONL manifest.** Same proven pattern. Append-only during logging, rewritten on edits
- **Multiple files per session.** Pull files into a focus set. They don't move. You work through them in order
- **Go + Wails for the GUI.** Keeps us in the Go ecosystem. Produces native-feeling Windows apps via webview. Same language as the TUI projects, lower cognitive overhead
- **mpv bundled.** Footage ships with mpv included. Zero setup — open the app, start logging. Larger download, but "install mpv first" is friction we don't want
- **Sessions persist with explicit new.** Closing Footage and reopening restores the last session — same files, same manifest, same state. "New session" starts fresh. Both paths are a single click
- **Full region editing.** After tagging, you can adjust start/end times (nudge or type), delete regions, and merge adjacent regions with the same tag. The log is mutable
- **One open region per tag per video.** Pressing a tag hotkey toggles that tag's open region. Different tags can have simultaneous open regions. Bookmarks are `kind: "bookmark"`, not regions with missing end times
- **ffmpeg stream copy is approximate.** Remux export is fast and lossless, but cuts may land on keyframes depending on codec. Documented tradeoff. Accurate re-encode mode later
