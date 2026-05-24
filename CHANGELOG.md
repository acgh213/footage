changelog
=========

all notable changes to footage. format based on
[keep a changelog](https://keepachangelog.com/en/1.1.0/). footage uses
[semver](https://semver.org/).

versioning
----------

- v0.1.0 — logging core. session, mpv sync, live tagging, bookmarks,
  single-region export
- v0.2.0 — editing + batch export. region editing, merge, batch UI
- v0.3.0 — LLM query. pop-up overlay, manifest search, video control
- v0.4.0 — python backend. whisper audio description, vision
  screenshot annotation, LLM consolidated into python
- v1.0.0 — polish. global search, statistics, export formats,
  keyframe strip, auto-advance
- v1.1.0 — sync. cross-machine manifest merge, content-hash file
  rebase, tombstones
- v1.2.0 — auto-tagging. vision-proposed regions with human accept
- v1.3.0 — EDL + DAW handoff. CMX 3600, fcpxml, otio export
- v1.4.0 — speaker diarization + game detection

0.1.0 — 2026-05-24
------------------

### added

- session management with persistence — last session restores on launch
- region logging with JSONL manifest (append-only, atomic rewrite on edits)
- stable ULIDs for all region and bookmark IDs
- one-open-region-per-tag-per-video toggle logic
- bookmark support (key 0, single-point, no end time)
- live in-progress region display — row appears in list immediately
  when a tag opens, with pulsing rec indicator and live elapsed time
- recording banner showing all open tags with elapsed time
- notes field auto-focuses after closing a region (enter to save, esc
  to skip). tag hotkeys fire through the notes field without interruption
- tag hotkeys always-on — fire regardless of focus, preventDefault
  prevents chars from landing in any active text input
- transport controls — ±5s, ±30s, play/pause, 1×/1.5×/2×/4× speed
- preset selector — switch between presets while a session is open
- session file panel — add videos, click to open in mpv, remove from session
- ffmpeg stream-copy export — single region to exports/ beside the source
  video, descriptive filename (stem_HH-MM-SS_tag.mp4)
- 250ms time-pos polling while mpv is running
- mpv named-pipe IPC over \\.\pipe\footage-mpv (async goroutine reader,
  request_id dispatch, reconnect-safe)
- mpv auto-discovery: PATH, cmd.exe where, WinGet packages, WinGet links,
  fixed install paths, bundled assets/mpv/mpv.exe
- manual mpv path override via file picker, saved to config.json
- built-in presets: destiny-raid and default (written on first launch)
- Wails v2 app skeleton — Go backend, Svelte + TypeScript frontend,
  WebView2 renderer

### fixed

- mpv not found when GUI process PATH differs from console PATH
- time-pos returning "property not found" — async IPC reader eliminates
  response routing races; 300ms stabilisation delay after pipe connect
- notes textarea not receiving focus — reactive fired before {#if} block
  mounted; fixed with afterUpdate pattern

0.0.0 — 2026-05-24
------------------

### added

- initial repo creation
- README with project vision and personal block
- DESIGN.md with architecture, data model, player integration, GUI
  layout, hotkey precision, decisions
- GOALS.md with versioned roadmap and non-goals
- PLAN.md with phased implementation breakdown and post-1.0 phases
- docs/schema.md with manifest schema v1
