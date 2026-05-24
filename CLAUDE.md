# CLAUDE.md

this file provides guidance to Claude Code (and other AI coding assistants) when working in this repository.

## project overview

Footage is a logging deck for game footage — a GUI application that controls an external video player (mpv, bundled) and lets you tag, annotate, and catalog segments of video with hotkeys. think "sports analyst replay station" for your game captures.

- **repo:** github.com/acgh213/footage
- **language:** Go (GUI, Wails) + Python (backend — whisper, vision, v0.4.0+)
- **player:** mpv (bundled, JSON IPC over named pipe)

## architecture

```
Footage (GUI, Wails)                 mpv (external, bundled)
┌─────────────────────┐           ┌──────────────────────┐
│  ┌───────────────┐  │   sync    │                      │
│  │  session      │  │◄────────►│  video playback      │
│  │  file list    │  │  (IPC)   │                      │
│  │               │  │           └──────────────────────┘
│  ├───────────────┤  │
│  │  tag panel    │  │
│  │  (hotkeys)    │  │
│  ├───────────────┤  │
│  │  region list  │  │
│  ├───────────────┤  │
│  │  notes        │  │
│  └───────────────┘  │
└─────────────────────┘
```

Footage does not embed a video player. It controls mpv externally. The GUI is the control surface — you watch the video in mpv's window, and you tag/edit/navigate from Footage.

## key design decisions

1. **GUI, not TUI.** Windows state sync between a TUI and external video player is unreliable. A GUI can IPC-sync with mpv cleanly. Built with Wails (Go + webview).
2. **mpv as sole player, bundled.** JSON IPC over named pipe (`\\.\pipe\footage-mpv`). No VLC fallback — one player, one integration path. Zero setup.
3. **Tag presets are global.** stored in `%APPDATA%/Footage/presets/`. available to all sessions.
4. **Regions are flat, overlap is allowed.** the manifest stores regions as independent entries. the UI handles nesting display.
5. **One open region per tag per video.** pressing a tag hotkey toggles that tag's open region. different tags can have simultaneous open regions. bookmarks are `kind: "bookmark"`, not regions with missing end times.
6. **Batch export, not live clipping.** logging and export are separate passes. log first, select regions, export later. ffmpeg stream copy is fast and lossless but approximate on keyframes.
7. **LLM is a pop-up, not a sidebar.** `Ctrl+L` opens an overlay. the LLM returns commands that drive the UI, not just text. summoned for a task, not a persistent conversation. direct Go HTTP calls to LLM API in v0.3.0 — no Python backend needed for queries.
8. **JSONL manifest.** append-only during logging, rewritten on edits. stable ULIDs, seconds as source of truth. see [docs/schema.md](docs/schema.md).
9. **Multiple files per session.** pull files into a focus set. they don't move. work through them sequentially.
10. **Audio description, not full transcription.** whisper runs for context-level audio description — enough to know what's being said, not a word-for-word transcript. v0.4.0+.
11. **Sessions persist.** close and reopen picks up where you left off. "new session" starts fresh.

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

see **[docs/schema.md](docs/schema.md)** for the full manifest schema. key rules:

- **Stable ULIDs** for all IDs (`region_01J...`, `session_01J...`, `bookmark_01J...`)
- **Seconds as source of truth** — `start_sec` and `end_sec` are authoritative. `display_start` and `display_end` are derived sugar
- **One open region per tag per video** — toggling behavior
- **Bookmarks are `kind: "bookmark"`** — `time_sec`, no end time, no tag
- **Regions can overlap** — flat storage, UI handles nesting
- **Append-only during logging, rewrite on edits** — write to temp file then rename

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

0. scaffolding (project structure, config, preset I/O, mpv IPC validation)
1. logging core (session, mpv sync, tagging, bookmarks, single-region export)
2. editing + batch export (region editing, batch remux)
3. LLM query engine (pop-up, manifest search, UI commands — Go direct HTTP, no Python)
4. Python backend (whisper, vision annotation)
5. polish (global search, statistics, export formats)

do not skip ahead. each phase depends on the one before it.

## naming conventions

- all markdown is lowercase with hyphens
- Go packages are lowercase, single word where possible
- JSONL is the manifest format — append-only during logging, rewritten on edits
- timestamps in manifest are ISO 8601 with milliseconds: `2026-06-01T14:22:00.500`

## references

- [DESIGN.md](DESIGN.md) — architecture, data model, GUI layout, player integration, hotkey precision, decisions
- [GOALS.md](GOALS.md) — versioned feature roadmap, non-goals
- [PLAN.md](PLAN.md) — phased implementation breakdown with tasks
- [CHANGELOG.md](CHANGELOG.md) — version history
- [docs/schema.md](docs/schema.md) — manifest schema v1 (regions, bookmarks, ULIDs, seconds-as-truth)
