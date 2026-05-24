# CLAUDE.md

this file provides guidance to Claude Code (and other AI coding assistants) when working in this repository.

## project overview

Footage is a logging deck for game footage — a GUI application that controls an external video player (mpv) and lets you tag, annotate, and catalog segments of video with hotkeys. think "sports analyst replay station" for your game captures.

- **repo:** github.com/acgh213/footage
- **language:** Go (GUI) + Python (backend — whisper, vision, LLM)
- **player:** mpv (primary, JSON IPC) with VLC fallback

## architecture

```
Footage GUI (Go)
  ├── session panel (file list)
  ├── tag panel (preset hotkeys)
  ├── region list (tagged segments)
  ├── transport bar (playback control)
  └── notes (per-region annotation)
        │
        │ IPC (mpv JSON API over named pipe)
        │
    mpv (external, video playback)
        │
        │ NDJSON subprocess
        │
Python backend (footage.py)
  ├── whisper (audio description)
  ├── vision model (screenshot annotation)
  └── LLM (query engine, manifest search, UI commands)
```

Footage does not embed a video player. It controls mpv externally. The GUI is the control surface — you watch the video in mpv's window, and you tag/edit/navigate from Footage.

## key design decisions

1. **GUI, not TUI.** Windows state sync between a TUI and external video player is unreliable. A GUI can IPC-sync with mpv cleanly.
2. **mpv as primary player.** JSON IPC over named pipe (`\\.\pipe\footage-mpv`). VLC HTTP API as fallback.
3. **Tag presets are global.** stored in `%APPDATA%/Footage/presets/`. available to all sessions. no per-project management.
4. **Regions are flat, overlap is allowed.** the manifest stores regions as independent entries. the UI handles nesting display.
5. **Batch export, not live clipping.** logging and export are separate passes. log first, select regions, export later.
6. **LLM is a pop-up, not a sidebar.** `Ctrl+L` opens an overlay. the LLM returns commands that drive the UI, not just text. dismiss when done.
7. **JSONL manifest.** append-only during logging, rewritten on edits. same proven pattern as the screenshot cataloger and chisel.
8. **Multiple files per session.** pull files into a focus set. they don't move. work through them sequentially.
9. **Audio description, not full transcription.** whisper runs for context-level audio description — enough to know what's being said, not a word-for-word transcript.

## project structure (planned)

```
footage/
├── README.md
├── DESIGN.md
├── GOALS.md
├── PLAN.md
├── CHANGELOG.md
├── CLAUDE.md              ← this file
├── footage.py             # Python backend (whisper, vision, LLM)
├── cmd/
│   └── footage/
│       └── main.go        # entry point
├── internal/
│   ├── app/               # application window, layout
│   ├── session/           # session management, manifest I/O
│   ├── preset/            # tag preset loading/saving
│   ├── player/            # mpv IPC client, transport control
│   ├── region/            # region CRUD, overlap handling
│   ├── export/            # ffmpeg remux batch export
│   ├── llm/               # LLM pop-up, query engine
│   └── config/            # settings, preferences
└── assets/
    └── presets/            # default presets (shipped with app)
```

## data model

### session on disk

```
my-session/
├── session.json           # session metadata: name, created, files[]
├── manifest.jsonl         # tagged regions
└── exports/               # remuxed clips
```

### manifest entry

```json
{
  "video": "D:/Captures/destiny2_raid_2021.mp4",
  "region": {"start": "00:03:12.500", "end": "00:05:47.200"},
  "tag": "boss encounter",
  "preset": "destiny-raid",
  "notes": "first Oryx clear. wiped at final stand once.",
  "logged_at": "2026-06-01T14:22:00",
  "transcript": null,
  "screenshots": [],
  "duration_sec": 154.7
}
```

### tag preset

```json
{
  "name": "destiny-raid",
  "tags": [
    {"key": "1", "label": "boss encounter", "color": "#e8919e"},
    {"key": "2", "label": "traversal", "color": "#a8c8e8"}
  ],
  "auto_notes": true
}
```

## mpv IPC protocol

```
# Windows named pipe
\\.\pipe\footage-mpv

# Commands (JSON)
{"command": ["set_property", "speed", 2.0]}
{"command": ["seek", -5, "relative"]}
{"command": ["get_property", "time-position"]}

# Response
{"data": 192.5, "error": "success"}
```

reference: https://mpv.io/manual/stable/#json-ipc

## ndjson protocol (Go ↔ Python)

same pattern as chisel and screenshot cataloger:

**request:**
```json
{"op": "transcribe", "video": "path/to/video.mp4", "region": {"start": "00:03:12", "end": "00:05:47"}}
```

**response:**
```json
{"op": "transcribe", "result": "Two speakers, combat dialogue, gunfire background", "status": "ok"}
```

## windows compatibility

- mpv named pipe: `\\.\pipe\footage-mpv` (Windows named pipe syntax)
- preset storage: `%APPDATA%/Footage/presets/` (use `os.UserConfigDir()`)
- session storage: user-chosen directory (not forced to AppData)
- all file paths use `filepath` package, never hardcoded separators
- ffmpeg must be in PATH or configured in settings
- `os/exec` for spawning mpv and ffmpeg subprocesses

## phase implementation order

follow PLAN.md. phases are ordered by dependency:

0. scaffolding (project structure, config, preset I/O)
1. logging core (session, mpv sync, tagging, region list, notes, bookmarks)
2. editing + export (region editing, batch remux export)
3. Python backend (whisper, vision annotation)
4. LLM query engine (pop-up, manifest search, UI commands)
5. polish (global search, statistics, export formats)

do not skip ahead. each phase depends on the one before it.

## naming conventions

- all markdown is lowercase with hyphens
- Go packages are lowercase, single word where possible
- JSONL is the manifest format — append-only during logging, rewritten on edits
- timestamps in manifest are ISO 8601 with milliseconds: `2026-06-01T14:22:00.500`

## references

- [DESIGN.md](DESIGN.md) — architecture, data model, GUI layout, player integration, decisions
- [GOALS.md](GOALS.md) — short-term through long-term feature roadmap, non-goals
- [PLAN.md](PLAN.md) — phased implementation breakdown with tasks
- [CHANGELOG.md](CHANGELOG.md) — version history
