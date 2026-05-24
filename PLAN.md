# ✧ Footage — implementation plan ✧

## phase 0: scaffolding (v0.0.1)

the project exists and runs. before building the logger, validate the scariest seam.

- [ ] **repo setup.** Go module init, Wails project structure, dependency management
- [ ] **mpv IPC validation.** Wails window opens. loads presets. user picks a video file. bundled mpv launches. "Get Time" button prints current mpv timestamp via named pipe
- [ ] **config loading.** `config.json` with default preset, seek intervals, LLM settings
- [ ] **preset I/O.** read/write tag preset files to `%APPDATA%/Footage/presets/`
- [ ] **verify.** `go build` produces `footage.exe`. mpv IPC round-trip works

## phase 1: logging core (v0.1.0)

watch, tag, annotate, extract the first clip.

- [ ] **session creation.** new session dialog: name, add video files (browse or drag-drop)
- [ ] **session persistence.** save/load session state to `session.json`
- [ ] **mpv launch.** spawn bundled mpv with named pipe IPC. start playback
- [ ] **transport control.** play/pause, seek relative (±5s, ±30s), seek absolute (click region in list)
- [ ] **playback speed.** 1x, 1.5x, 2x, 4x hotkeys or transport buttons
- [ ] **preset loading.** dropdown to select active preset. hotkeys displayed in tag panel
- [ ] **live tagging.** hotkey starts open region. hotkey again closes it. one open region per tag per video. timestamp captured from mpv IPC. seconds stored as truth
- [ ] **bookmarks.** dedicated hotkey for `kind: "bookmark"`. one tap, no close. distinct from regions
- [ ] **region list.** scrollable, color-coded by tag, sorted by start time. click to seek
- [ ] **notes.** auto-focus notes field on region close. simple text input
- [ ] **manifest write.** append regions and bookmarks to `manifest.jsonl` per [docs/schema.md](docs/schema.md). stable ULIDs, seconds as source of truth
- [ ] **multi-file session.** "next file" button loads next video in the session
- [ ] **open regions preserved.** if a region is still open when switching files, it stays open and tagged to the current video
- [ ] **single-region export.** right-click a region → export clip via ffmpeg stream copy. approximate keyframe accuracy. output to `exports/` with descriptive filename

## phase 2: editing + batch export (v0.2.0)

refine the log, get multiple clips out.

- [ ] **region editing.** adjust start/end times (nudge or type). delete region with confirmation
- [ ] **merge regions.** select two adjacent regions with the same tag, merge into one
- [ ] **overlap display.** nested/overlapping regions shown with indentation in the list
- [ ] **batch export UI.** checkbox column in region list. select all / select by tag / select by search
- [ ] **remux export.** ffmpeg stream copy (`-c copy`) for selected regions. output to `exports/` with descriptive filenames
- [ ] **export progress.** progress bar during batch export. cancel available

## phase 3: LLM query engine (v0.3.0)

talk to your footage. no Python backend — Go calls the LLM API directly.

- [ ] **LLM pop-up.** `Ctrl+L` opens pop-up overlay. query input, response area, dismiss on Escape
- [ ] **manifest search.** structured queries against the manifest. LLM returns filtered results
- [ ] **UI commands.** LLM can return commands: seek to timestamp, load file, filter region list by tag
- [ ] **contextual Q&A.** "what happened in this clip?" → LLM reads manifest entries for current video and answers
- [ ] **provider config.** OpenAI-compatible endpoint, configurable model, local-first

## phase 4: Python backend (v0.4.0)

heavy lifting moves to a subprocess.

- [ ] **footage.py backend.** NDJSON protocol, same pattern as screenshot cataloger and chisel
- [ ] **whisper transcription.** select regions, queue for audio description. whisper runs locally. results write back to manifest
- [ ] **screenshot extraction.** ffmpeg frame extraction at start, middle, end of tagged regions
- [ ] **vision annotation.** feed frames through vision model. descriptions write back to manifest

## phase 5: polish (v1.0.0)

- [ ] **global search UI.** search bar with results from all sessions
- [ ] **statistics panel.** tag frequency, total logged time, session summaries
- [ ] **auto-advance.** option to auto-load next file when current video ends
- [ ] **keyframe strip.** visual reference strip from the video timeline
- [ ] **export formats.** CSV export, JSON export

## dependencies

**GUI application:**
- Go 1.21+ (language)
- [Wails](https://wails.io/) (GUI framework — Go backend + webview frontend, native-feeling Windows apps)
- mpv (bundled, JSON IPC for transport control)

**External:**
- ffmpeg (remux export, frame extraction)
- Python 3.10+ (backend: whisper, vision — v0.4.0+)

**Python packages (v0.4.0+):**
- `openai` (LLM API client — also used directly from Go in v0.3.0)
- `faster-whisper` (local transcription)
- `Pillow` (image handling for vision pipeline)

## windows notes

- mpv named pipe on Windows: `\\.\pipe\footage-mpv`
- preset storage: `%APPDATA%/Footage/presets/`
- session storage: user-chosen directory (sessions persist between launches)
- ffmpeg must be in PATH or configured in settings
- all file paths use Go's `filepath` package, never hardcoded separators
