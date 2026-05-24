# ✧ Footage — implementation plan ✧

## phase 0: scaffolding (v0.0.1)

the project exists and runs.

- [ ] **repo setup.** Go module init, project structure, dependency management
- [ ] **GUI scaffold.** application window with menu bar, basic layout panes (session, tags, regions, transport)
- [ ] **config loading.** `config.json` with mpv path, default preset, seek intervals, LLM settings
- [ ] **preset I/O.** read/write tag preset files to `%APPDATA%/Footage/presets/`
- [ ] **verify.** `go build` produces `footage.exe`. app launches with empty panes

## phase 1: logging core (v0.1.0)

watch, tag, annotate.

- [ ] **session creation.** new session dialog: name, add video files (browse or drag-drop)
- [ ] **session persistence.** save/load session state to `session.json`
- [ ] **mpv launch.** spawn mpv with named pipe IPC. start playback
- [ ] **transport control.** play/pause, seek relative (±5s, ±30s), seek absolute (click region in list)
- [ ] **playback speed.** 1x, 1.5x, 2x, 4x hotkeys or transport buttons
- [ ] **preset loading.** dropdown to select active preset. hotkeys displayed in tag panel
- [ ] **live tagging.** hotkey starts open region. hotkey again closes it with end time. timestamp captured from mpv IPC
- [ ] **region list.** scrollable, color-coded by tag, sorted by start time. click to seek
- [ ] **notes.** auto-focus notes field on region close. simple text input
- [ ] **bookmarks.** dedicated hotkey for one-tap bookmarks. no close required
- [ ] **manifest write.** append tagged regions to `manifest.jsonl` on close
- [ ] **multi-file session.** "next file" button loads next video in the session. manifest tracks which video each region belongs to
- [ ] **open regions preserved.** if a region is still open when switching files, it stays open and tagged to the current video

## phase 2: editing + export (v0.2.0)

refine the log, get clips out.

- [ ] **region editing.** click to adjust start/end times (text entry or nudge buttons). delete region with confirmation
- [ ] **merge regions.** select two adjacent regions with the same tag, merge into one
- [ ] **overlap display.** nested/overlapping regions shown with indentation in the list
- [ ] **batch export UI.** checkbox column in region list. select all / select by tag / select by search
- [ ] **remux export.** ffmpeg stream copy (`-c copy`) for selected regions. output to `exports/` with descriptive filenames
- [ ] **export progress.** progress bar during batch export. cancel available

## phase 3: Python backend (v0.2.0 cont.)

heavy lifting moves to a subprocess.

- [ ] **footage.py backend.** NDJSON protocol, same pattern as screenshot cataloger and chisel
- [ ] **whisper transcription.** select regions, queue for audio description. whisper runs locally. results write back to manifest
- [ ] **screenshot extraction.** ffmpeg frame extraction at start, middle, end of tagged regions
- [ ] **vision annotation.** feed frames through vision model. descriptions write back to manifest `screenshots` field

## phase 4: LLM query engine (v0.3.0)

talk to your footage.

- [ ] **LLM pop-up.** `Ctrl+L` opens pop-up overlay. query input, response area, dismiss on Escape
- [ ] **manifest search.** structured queries against the manifest. LLM returns filtered results
- [ ] **UI commands.** LLM can return commands: seek to timestamp, load file, filter region list by tag
- [ ] **contextual Q&A.** "what happened in this clip?" → LLM reads manifest entries for current video and answers
- [ ] **cross-session search.** search across all sessions. "find every firefight from 2023"

## phase 5: polish (v1.0.0)

- [ ] **global search UI.** search bar with results from all sessions
- [ ] **statistics panel.** tag frequency, total logged time, session summaries
- [ ] **auto-advance.** option to auto-load next file when current video ends
- [ ] **keyframe strip.** visual reference strip from the video timeline
- [ ] **export formats.** CSV export, JSON export, EDL export (v1.1)

## dependencies

**GUI application:**
- Go (language)
- GUI framework TBD (see DESIGN.md questions)

**External:**
- mpv (primary video player, JSON IPC)
- ffmpeg (remux export, frame extraction)
- Python 3.10+ (backend: whisper, vision, LLM)

**Python packages:**
- `openai` (LLM API client)
- `faster-whisper` (local transcription)
- `Pillow` (image handling for vision pipeline)

## windows notes

- mpv named pipe on Windows: `\\.\pipe\footage-mpv`
- preset storage: `%APPDATA%/Footage/presets/`
- session storage: user-chosen directory
- ffmpeg must be in PATH or configured in settings
- all file paths use Go's `filepath` package, never hardcoded separators
