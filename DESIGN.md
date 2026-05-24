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
│              IPC (mpv JSON API / VLC HTTP)                      │
│                         │                                       │
│              ┌──────────────────────┐                           │
│              │  mpv / VLC           │                           │
│              │  (external window)   │                           │
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
- Can be bundled with Footage or pointed at an existing install

### VLC (fallback)

VLC's HTTP API on `localhost:8080` for basic transport control. Higher latency, less reliable for frame-accurate seeking. Fallback only.

## data model

### session

A session is a *focus set* — a group of video files you're working through. Files aren't moved or copied. The session just tracks which files are in the set and what's been logged.

```
my-session/
├── session.json          # session metadata: name, created, files[]
├── manifest.jsonl        # tagged regions, one JSON object per line
└── exports/              # remuxed clips (created on batch export)
```

### manifest format (JSONL)

```json
{
  "video": "D:/Captures/destiny2_raid_2021.mp4",
  "region": {"start": "00:03:12.500", "end": "00:05:47.200"},
  "tag": "boss encounter",
  "preset": "destiny-raid",
  "notes": "first Oryx clear. wiped at final stand once. Octavio was screaming.",
  "logged_at": "2026-06-01T14:22:00",
  "transcript": null,
  "screenshots": [],
  "duration_sec": 154.7
}
```

A region without an end time is an *open region* — still being tagged. Tags are freeform strings. The `preset` field tracks which preset group the tag belongs to.

Regions can overlap. Nested regions are stored flat in the manifest and the UI handles the nesting display. A QTE inside a boss encounter is two separate entries with overlapping time ranges.

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
- **mpv as primary player.** JSON IPC is fast, scriptable, and mpv handles every video format. VLC as HTTP fallback
- **Tag presets are global.** Saved in AppData, available to all sessions. No per-project preset management
- **Regions are flat, overlap is allowed.** The UI handles nesting display. The manifest doesn't enforce hierarchy
- **Batch export, not live clipping.** Logging and exporting are separate passes. Log first, export later
- **LLM is a pop-up, not a sidebar.** It's summoned for a task, not a persistent conversation partner
- **JSONL manifest.** Same proven pattern. Append-only during logging, rewritten on edits
- **Multiple files per session.** Pull files into a focus set. They don't move. You work through them in order

## questions for Cassie

1. **GUI framework.** Go + Wails keeps us in the Go ecosystem and produces a native-feeling Windows app. Python + PySide/PyQt is heavier but more flexible. Go + Fyne is pure Go but looks less polished. Preference?

2. **mpv bundling.** Should Footage ship with mpv bundled (zero setup, larger download), or require mpv in PATH (lighter, one setup step)?

3. **Session scope.** Does a session span one sitting (review clips, close app, done) or persist across sessions (reopen next week, pick up where you left off)?

4. **Region editing.** After tagging, can you adjust start/end times? Delete regions? Merge adjacent regions with the same tag? Or is the log immutable once closed?
