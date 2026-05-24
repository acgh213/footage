design
======

architecture
------------

```
+-----------------------------------------------------------------+
|                     footage (GUI application)                    |
|                                                                 |
|  +-------------+  +--------------+  +-----------------------+   |
|  |  session    |  |   tag panel  |  |     region list       |   |
|  |  (files)    |  |  (presets,   |  |  (timeline of marked  |   |
|  |             |  |   hotkeys)   |  |   segments, live)     |   |
|  +-------------+  +--------------+  +-----------------------+   |
|                                                                 |
|  +-----------------------------------------------------------+  |
|  |  transport bar  > || prev 5s  next 30s   1x 1.5x 2x 4x   |  |
|  +-----------------------------------------------------------+  |
|                                                                 |
|                         |                                       |
|              IPC (mpv JSON API)                                 |
|                         |                                       |
|              +----------------------+                           |
|              |  mpv                 |                           |
|              |  (external window,   |                           |
|              |   bundled)           |                           |
|              +----------------------+                           |
|                                                                 |
|  +----------------------------------------------------------+   |
|  |               python backend (footage.py)                |   |
|  |  +----------+  +------------+  +--------------------+   |   |
|  |  | whisper  |  |  vision    |  | LLM (query engine) |   |   |
|  |  | (audio)  |  | (frames)   |  |                    |   |   |
|  |  +----------+  +------------+  +--------------------+   |   |
|  +----------------------------------------------------------+   |
+-----------------------------------------------------------------+
```

the GUI handles the logging deck: session management, tag presets, hotkey
input, region editing, notes, transport control. the python backend
handles heavy lifting: whisper transcription, vision annotation, and
(deferred from v0.3.0) LLM-orchestrated queries. communication is ndjson
over subprocess — same pattern as the screenshot cataloger and chisel.

player integration
------------------

footage doesn't embed a video player. it controls one externally via IPC.

### mpv (primary, bundled, sole)

mpv exposes a JSON IPC API over a named pipe. footage sends commands and
reads playback position:

```
# Windows named pipe
\\.\pipe\footage-mpv

# Commands
{"command": ["set_property", "speed", 2.0]}
{"command": ["seek", -5, "relative"]}
{"command": ["get_property", "time-position"]}
```

benefits:

- open source, fast, handles any format
- scriptable IPC, no HTTP overhead
- bundled with footage, zero setup

reference: <https://mpv.io/manual/stable/#json-ipc>

data model
----------

### session

a session is a *focus set* — a group of video files you're working
through. files aren't moved or copied. the session just tracks which
files are in the set and what's been logged.

```
my-session/
+- session.json          # session metadata: name, created, files[]
+- manifest.jsonl        # tagged regions and bookmarks
+- exports/              # remuxed clips (created on batch export)
```

### manifest format

see [docs/schema.md](docs/schema.md) for the full schema. key points:

- stable ULIDs for regions, bookmarks, sessions, videos
- seconds are source of truth. `start_sec`/`end_sec` are authoritative.
  display timestamps are derived sugar
- append-only during logging, rewrite on edits. no event log in v0.1
- regions can overlap. flat storage, UI handles nesting display
- bookmarks are `kind: "bookmark"`. single timestamp, no end, no tag

### tag presets

presets are global, stored in `%APPDATA%/Footage/presets/`. each preset
is a JSON file:

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

`auto_notes` controls whether a notes field appears automatically when
closing a tag region.

GUI design
----------

### layout

```
+----------------------------------------------------------------+
|  footage  -  destiny2_raid_2021.mp4                    _ [] x  |
+------------+-------------------------+-------------------------+
|  SESSION   |                         |  REGIONS                |
|            |      TAG PRESET         |                         |
|  + raid    |  +-------------------+  |  00:03:12 - 00:05:47    |
|    > oryx  |  | destiny-raid      |  |  [1] boss encounter     |
|    > gate  |  |                   |  |  first oryx clear       |
|  + pvp     |  | 1  boss encounter |  |                         |
|    > ib    |  | 2  traversal      |  |  00:08:30 - 00:10:15    |
|    > tri   |  | 3  adds clear     |  |  [2] traversal          |
|            |  | 4  wipe           |  |                         |
|            |  | 5  loot / chest   |  |  00:12:00 - 00:14:45    |
|            |  | 6  menu / loadout |  |  [3] adds clear         |
|            |  | 7  cutscene       |  |                         |
|            |  | 0  bookmark       |  |  > 00:16:22 - ...       |
|            |  |                   |  |  [b] bookmark           |
|            |  +-------------------+  |                         |
|            |                         |                         |
|            |  NOTES                  |                         |
|            |  +-------------------+  |                         |
|            |  | (active or last)  |  |                         |
|            |  +-------------------+  |                         |
+------------+-------------------------+-------------------------+
|  >  ||   -5s  +30s    1x 1.5x 2x 4x   [============   ]  12:34 |
+----------------------------------------------------------------+
```

### interaction model

1. load a session — drag video files in, or open a folder. they appear
   in the session panel
2. pick a preset — dropdown of your saved tag groups. the hotkeys load
3. play — mpv launches. footage syncs playback position via IPC
4. tag as you watch — tap `1` when a boss encounter starts, tap `1`
   again when it ends. the region appears in the list
5. notes — when you close a region, the notes field focuses for quick
   annotation
6. jump around — transport bar controls mpv. seek back 5 seconds, skip
   ahead 30
7. variable speed — 1.5x, 2x, 4x for fast-forwarding through traversal
8. bookmarks — tap `0` for "come back to this." one tap, no region to
   close

### hotkey precision

- one open region per tag per video. pressing `1` opens a boss
  encounter region. pressing `1` again closes it. pressing `2` while
  `1` is open opens a traversal region — both can be open
  simultaneously, but only one of each tag type
- bookmarks are separate. the bookmark hotkey creates `kind: "bookmark"`
  entries. they are not regions with missing end times
- closing a region auto-focuses notes. when you close a region, the
  notes field receives focus

### LLM pop-up

`Ctrl+L` opens a pop-up overlay, not a sidebar. it's an interaction,
not a persistent chat. the LLM can:

- search the manifest ("show me every wipe from the oryx encounter")
- jump the video ("go to the first boss encounter in this clip")
- answer questions ("what was i doing at 14:30 in this video?")
- drive the UI — the LLM returns commands, not just text

the pop-up closes when you dismiss it. it doesn't linger. you summon
it for a task, get the result, move on.

backend processing
------------------

### audio description transcription

run whisper locally on tagged regions. not full dialogue transcription —
context-level:

- "two speakers, one giving instructions, gunfire in background"
- speaker identification for multi-speaker segments
- key phrases extracted, not full text

runs as a batch job after logging — select regions, queue whisper, get
descriptions back into the manifest.

### screenshot annotation

extract frames from tagged regions (start, middle, end + key frames)
and feed them through the same vision pipeline as the screenshot
cataloger. annotated frames link back to their parent region in the
manifest.

### LLM query engine

the python backend loads the manifest and provides a query interface.
the local LLM (same infrastructure as chisel) can:

- execute structured searches against the manifest
- generate human-readable answers from manifest data
- return UI commands (seek, load, filter)

note: in v0.3.0 the LLM is wired directly from go via HTTP — no python
required. the python backend in v0.4.0 absorbs the LLM path when
whisper/vision arrive, so all heavy work lives in one process.

decisions
---------

- GUI, not TUI. windows state sync between a TUI and external video
  player is unreliable. a GUI can IPC-sync with mpv cleanly
- mpv as sole player. JSON IPC is fast, scriptable, and bundled. no
  VLC fallback — one player, one integration path
- tag presets are global. saved in AppData, available to all sessions.
  no per-project preset management
- regions are flat, overlap is allowed. the UI handles nesting display.
  the manifest doesn't enforce hierarchy
- batch export, not live clipping. logging and exporting are separate
  passes. log first, export later
- LLM is a pop-up, not a sidebar. it's summoned for a task, not a
  persistent conversation partner
- JSONL manifest. same proven pattern. append-only during logging,
  rewritten on edits
- multiple files per session. pull files into a focus set. they don't
  move. work through them in order
- go + wails for the GUI. keeps us in the go ecosystem. produces
  native-feeling windows apps via webview. same language as the TUI
  projects, lower cognitive overhead
- mpv bundled. footage ships with mpv included. zero setup — open the
  app, start logging. larger download, but "install mpv first" is
  friction we don't want
- sessions persist with explicit new. closing footage and reopening
  restores the last session — same files, same manifest, same state.
  "new session" starts fresh. both paths are a single click
- full region editing. after tagging, you can adjust start/end times
  (nudge or type), delete regions, and merge adjacent regions with the
  same tag. the log is mutable
- one open region per tag per video. pressing a tag hotkey toggles
  that tag's open region. different tags can have simultaneous open
  regions. bookmarks are `kind: "bookmark"`, not regions with missing
  end times
- ffmpeg stream copy is approximate. remux export is fast and
  lossless, but cuts may land on keyframes depending on codec.
  documented tradeoff. accurate re-encode mode later
- python absorbs LLM in v0.4.0. go calls the LLM directly in v0.3.0
  (no python needed for query-only). once whisper/vision arrive, all
  heavy work consolidates into the python subprocess
