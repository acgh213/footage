plan
====

phased implementation. each phase depends on the one before. don't skip
ahead. files listed per phase are the concrete shape; subtasks under each
file are what actually has to happen in code.

dependencies
------------

GUI:

- go 1.21+
- [wails](https://wails.io/) (go backend + webview frontend)
- mpv (bundled, JSON IPC for transport)

external:

- ffmpeg (remux export, frame extraction)
- python 3.10+ (backend: whisper, vision — v0.4.0+)

go modules (added per phase):

- `github.com/oklog/ulid/v2` (ULIDs, v0.1.0)
- `github.com/Microsoft/go-winio` (named pipe on windows, v0.0.1)
- `github.com/wailsapp/wails/v2` (GUI, v0.0.1)

python packages (v0.4.0+):

- `faster-whisper` (local transcription)
- `Pillow` (image handling)
- `openai` (LLM client — also used directly from go in v0.3.0)

windows notes
-------------

- mpv named pipe: `\\.\pipe\footage-mpv`
- preset storage: `%APPDATA%/Footage/presets/`
- session storage: user-chosen directory (sessions persist between launches)
- ffmpeg must be in PATH or configured in settings
- all file paths use go's `filepath` package, never hardcoded separators

phase 0 — scaffolding (v0.0.1)
------------------------------

the project exists, runs, and proves the scariest seam (mpv IPC). before
building the full logger, the smoke window has to work end-to-end.

### files

```
go.mod                                  module github.com/acgh213/footage
cmd/footage/main.go                     wails app entrypoint, lifecycle
internal/config/config.go               read/write %APPDATA%/Footage/config.json
internal/config/defaults.go             default seek intervals, default preset
internal/preset/preset.go               load/save preset JSON, list presets
internal/preset/builtin.go              ship destiny-raid preset as default
internal/player/mpv.go                  spawn mpv subprocess, attach to pipe
internal/player/ipc.go                  named-pipe client: request, response, observe
internal/player/commands.go             typed wrappers: GetTimePos, Seek, SetSpeed, Pause
frontend/wails.json                     wails config
frontend/src/App.svelte                 phase-0 smoke window
frontend/src/lib/Smoke.svelte           file picker + "get time" button
assets/mpv/                             bundled mpv (gitignored, fetched at build)
build/windows/installer.nsi             eventual installer config (stub for now)
```

### subtasks

`internal/player/mpv.go`:

- locate bundled mpv at `assets/mpv/mpv.exe` (windows)
- spawn with flags: `--input-ipc-server=\\.\pipe\footage-mpv --idle`
- poll until pipe is writable (timeout: 3s)
- expose `Start(videoPath string) error` and `Stop() error`
- on Stop: send `quit` over IPC, wait, then kill if still alive

`internal/player/ipc.go`:

- dial named pipe (windows: `go-winio` package)
- monotonic `request_id` counter
- background goroutine reading newline-delimited JSON responses
- channels keyed by `request_id` for response routing
- separate channel for `observe_property` events

`internal/player/commands.go`:

- `GetTimePos() (float64, error)` → `get_property time-position`
- `Seek(delta float64, relative bool) error`
- `SetSpeed(s float64) error`
- `Pause() error` / `Play() error` → `set_property pause true|false`

`internal/preset/preset.go`:

- `PresetDir() string` → `os.UserConfigDir()/Footage/presets`
- `Load(name string) (*Preset, error)`
- `Save(p *Preset) error` (write to temp, atomic rename)
- `List() ([]string, error)`

`internal/config/config.go`:

- `Load() (*Config, error)` — read or create with defaults
- `Save(c *Config) error` — atomic write
- struct: `LastSession`, `DefaultPreset`, `SeekShort`, `SeekLong`, `LLM`

### verify

- `go build ./cmd/footage` produces `footage.exe`
- launch app, pick a video, mpv window opens
- "get time" button prints current `time-pos` to the UI
- scrub in mpv, click again — new pos reflected
- kill the app cleanly, no orphaned mpv process

phase 1 — logging core (v0.1.0)
-------------------------------

watch, tag, annotate, extract the first clip. closes the loop.

### files

```
internal/session/session.go             Session struct, Files[], metadata
internal/session/store.go               save/load session.json
internal/session/persistence.go         "last session" pointer, restore on launch
internal/region/region.go               Region, Bookmark, OpenRegion types
internal/region/store.go                manifest.jsonl writer (append), reader
internal/region/ulid.go                 ULID generation
internal/region/toggle.go               one-open-region-per-tag-per-video toggling
internal/export/remux.go                ffmpeg stream copy
internal/export/filename.go             descriptive filename builder
frontend/src/lib/SessionPanel.svelte    file list, drag-drop, switch active
frontend/src/lib/TagPanel.svelte        preset selector, hotkey grid
frontend/src/lib/RegionList.svelte      scrollable, color-coded, click to seek
frontend/src/lib/NotesField.svelte      auto-focus on region close
frontend/src/lib/Transport.svelte       play/pause, ±5s, ±30s, speed buttons
frontend/src/stores/session.ts          svelte store mirroring backend session
frontend/src/stores/playback.ts         time-pos, speed, paused (observed)
```

### subtasks

`internal/session/session.go`:

- struct: `ID` (ULID), `Name`, `Dir`, `Files []SessionFile`, `ActiveFile`,
  `CreatedAt`, `UpdatedAt`
- `New(name, dir string) (*Session, error)`
- `AddFile(path string) error` (dedupe by absolute path)
- `RemoveFile(path string)` (no manifest deletion — just unlist)

`internal/session/persistence.go`:

- on app start: read `config.LastSession`, load it if present
- on session change: update `config.LastSession` and `Save()`
- "new session" button clears the pointer first, then starts fresh

`internal/region/region.go`:

- types matching [docs/schema.md](docs/schema.md)
- `Region`: `ID`, `SessionID`, `VideoPath`, `StartSec`, `EndSec`,
  `Tag`, `Preset`, `Notes`, `CreatedAt`, `UpdatedAt`
- `Bookmark`: same minus `EndSec`, `Tag`, `Preset`; plus `TimeSec`
- `OpenRegion`: in-memory only, becomes a `Region` when closed
- `DisplayStart()/DisplayEnd()` derive from seconds, do not store

`internal/region/store.go`:

- `WriteRegion(r Region) error` — append one line, fsync optional
- `WriteBookmark(b Bookmark) error` — same
- `ReadAll(path string) ([]Entry, error)` — parse jsonl, skip blank
  lines, error on malformed lines (with line number)
- `Rewrite(path string, entries []Entry) error` — write temp, fsync,
  atomic rename

`internal/region/toggle.go`:

- state: `map[videoPath]map[tagKey]*OpenRegion`
- `Toggle(videoPath, tagKey string, now float64) (*Region, bool)` —
  returns closed region + true, or nil + false if a new one opened
- `Bookmark(videoPath string, now float64) Bookmark`
- `SwitchVideo(newPath string)` — preserves open regions; subsequent
  toggles for those tags resolve against their original video
- `OpenRegions()` for UI to show which tags are "lit"

`internal/export/remux.go`:

- args: `ffmpeg -ss <start> -to <end> -i <video> -c copy -avoid_negative_ts make_zero <out>`
- run with stderr captured; parse `time=` lines for progress
- return `(cancel func(), done chan error)`

`internal/export/filename.go`:

- pattern: `{video_stem}_{HH-MM-SS}_{tag}_{ulid_suffix}.{ext}`
- sanitize tag for filesystem (spaces → `-`, drop slashes)

frontend stores:

- `playback.ts` polls mpv every 100ms via `GetTimePos`, or uses
  observe_property if available. update reactively
- `session.ts` round-trips to backend on every change

### verify

- new session, add 2 video files, file list shows both
- close app, reopen — last session restored, region list intact
- play a 30s video, tap `1` at 5s, tap `1` at 10s — region appears
- tap `1` at 12s — opens a new region (the prior one is closed)
- tap `2` while `1` is open — both open simultaneously
- tap `0` — bookmark appears, no end time
- close a region — notes field focused, type a note, manifest updates
  on next write
- right-click region → export → mp4 lands in `exports/`, opens in any
  player

phase 2 — editing + batch export (v0.2.0)
-----------------------------------------

refine the log, get multiple clips out.

### files

```
internal/region/edit.go                 nudge/set start/end, delete, merge
internal/region/overlap.go              compute nesting tree (display-only)
internal/region/search.go               filter regions by tag/text/duration
internal/export/batch.go                queue, progress, cancel, per-region status
frontend/src/lib/RegionEditor.svelte    inline timestamp editing
frontend/src/lib/BatchExportDialog.svelte  checkbox column, filters, run
frontend/src/lib/ConfirmDelete.svelte   reusable confirm modal
```

### subtasks

`internal/region/edit.go`:

- `Nudge(id string, field string, delta float64) error` — field is
  `"start"` or `"end"`
- `Set(id string, field string, value float64) error`
- `Delete(id string) error` — backend just deletes; UI does the confirm
- `Merge(idA, idB string) error` — same tag, same video, gap ≤ tolerance
  (default 500ms). produces one region with the earliest start and the
  latest end
- all four trigger `region.store.Rewrite`

`internal/region/overlap.go`:

- `Tree(regions []Region) []Node` — group by video, sort by start,
  build containment tree (region A contains region B if
  A.start ≤ B.start and B.end ≤ A.end)
- pure display helper, no schema change

`internal/region/search.go`:

- `Filter(regions []Region, q Query) []Region`
- `Query`: `Tags []string`, `Video string`, `ContainsText string`,
  `MinDuration`, `MaxDuration`

`internal/export/batch.go`:

- `Run(ctx context.Context, regions []Region, outDir string, onProgress func(id string, pct float64)) error`
- one ffmpeg at a time (ffmpeg parallelizes internally; queuing avoids
  contention)
- per-region status broadcast to the UI

### verify

- select region, click +5s on end, time updates, manifest rewritten
- delete region with confirm dialog — region disappears, manifest
  shrinks
- merge two adjacent same-tag regions — one region remains, span is
  union
- select 3 regions, batch export, progress bar fills, 3 files in
  `exports/`
- cancel mid-export — running ffmpeg dies, partial files cleaned up

phase 3 — LLM query engine (v0.3.0)
-----------------------------------

talk to your footage. no python yet — go calls the LLM API directly
over HTTP.

### files

```
internal/llm/client.go                  openai-compatible HTTP client
internal/llm/prompt.go                  system prompt + manifest packing
internal/llm/tools.go                   tool schemas
internal/llm/dispatch.go                parse tool calls, route to UI actions
internal/llm/provider.go                provider config (endpoint, model, key)
frontend/src/lib/LLMPopup.svelte        Ctrl+L overlay
frontend/src/lib/LLMResult.svelte       answer text + applied-action chips
```

### subtasks

`internal/llm/client.go`:

- standard openai chat completions request shape
- support tool calls
- timeout (default 30s)
- stream support optional in v0.3.0

`internal/llm/tools.go`:

- `seek(seconds float64)` — jump current video to absolute time
- `load_file(video_path string)` — switch active video
- `filter_regions(tag?, video?, contains_text?)` — narrow the region list
- `answer(text string)` — terminal action; the prose to display

`internal/llm/prompt.go`:

- compact representation of current session manifest
- format per region: `[id] <video> <start>-<end> <tag>: <notes>`
- window: most recent N regions (default 200), or all for current video
- system prompt: brief, tool-use-first, lowercase

`internal/llm/dispatch.go`:

- parse the tool call from the response
- route to: `player.Seek`, `session.LoadFile`, `regionList.SetFilter`,
  or display answer text
- a response may chain: filter + answer

`internal/llm/provider.go`:

- read `config.LLM`: endpoint, model, api_key (or env var name)
- default endpoint placeholder — local-first preference

### verify

- `Ctrl+L` opens popup
- "show me every wipe" → region list filters to tag=wipe
- "go to the first boss encounter" → mpv seeks to that region's start
- "what was happening at 14:30?" → LLM reads manifest, returns prose
- `Escape` dismisses the popup, no state retained

phase 4 — python backend (v0.4.0)
---------------------------------

heavy lifting moves to a long-lived subprocess. the LLM path also
migrates here so all model interaction lives in one place.

### files

```
footage.py                              ndjson main loop
py/ops/__init__.py
py/ops/transcribe.py                    faster-whisper wrapper
py/ops/vision.py                        vision model call
py/ops/extract_frames.py                ffmpeg frame extraction
py/ops/llm.py                           LLM (absorbed from internal/llm)
py/util/ndjson.py                       stdin/stdout helpers
internal/backend/client.go              spawn footage.py, ndjson stdin/stdout
internal/backend/ops.go                 typed RPC wrappers
internal/backend/restart.go             health check + restart on crash
```

### subtasks

`internal/backend/client.go`:

- spawn `footage.py` as long-lived subprocess
- newline-delimited JSON request/response
- `request_id` correlation
- restart on crash (with exponential backoff capped at 30s)
- shut down cleanly on app exit

`internal/backend/ops.go`:

- `Transcribe(region Region) (string, error)`
- `Vision(region Region, frames []string) ([]string, error)`
- `LLM(request LLMRequest) (LLMResponse, error)` — replaces direct
  HTTP from v0.3.0

`footage.py`:

- read stdin line-by-line, parse `op`, dispatch
- write `{"op":..., "request_id":..., "result":..., "status":"ok"}` or
  `"error"` with `"message"`
- lazy model loading: whisper loads on first transcribe op, vision on
  first vision op. keep loaded across requests
- graceful shutdown on `op:"shutdown"` or EOF on stdin

`py/ops/transcribe.py`:

- input: video path + start_sec + end_sec
- extract audio with ffmpeg to a temp wav (or use whisper's video
  decoder directly if faster)
- run faster-whisper in `transcribe` or `summarize` mode (the latter
  via a downstream LLM call to condense the transcript into a
  description)
- return: `{"description": str, "speakers": int, "keyphrases": [str]}`

`py/ops/vision.py`:

- input: list of frame paths
- batch call to the vision model
- return: one description per frame, plus a region-level summary

`py/ops/extract_frames.py`:

- input: video + region
- output: 5 frames — start, end, midpoint, plus 2 keyframes inside
  the region (use ffmpeg `select='eq(pict_type,I)'`)
- written to a per-region cache dir

### verify

- select 3 regions, queue transcribe, descriptions appear in manifest
  within reasonable time
- vision pass: extracted frames in cache, descriptions on regions
- LLM via python: same behaviors as phase 3, no direct HTTP from go
- crash the python process (`taskkill /F /IM python.exe`) — backend
  restarts automatically, queued ops requeue

phase 5 — polish (v1.0.0)
-------------------------

cross-session features and UX shine.

### files

```
internal/search/global.go               cross-session search index
internal/search/index.go                lightweight index (in-memory + persisted)
internal/stats/stats.go                 tag frequency, total time, busiest days
internal/export/formats.go              csv export, json export of manifest
internal/playback/auto_advance.go       hook for "video ended" event
frontend/src/lib/KeyframeStrip.svelte   timeline thumbnail strip
frontend/src/lib/Stats.svelte           stats panel
frontend/src/lib/Preferences.svelte     settings page
frontend/src/lib/GlobalSearch.svelte    cross-session search UI
```

### subtasks

`internal/search/global.go`:

- enumerate sessions in known session roots
- build a per-session region index lazily on first access
- search by: tag, free-text in notes, date range, duration range
- results return `(SessionID, RegionID)` pairs for jump-to

`internal/stats/stats.go`:

- tag frequency: count per tag
- total logged time: sum of region durations
- average segment duration per tag
- busiest logging days: regions created per day (top N)

`internal/export/formats.go`:

- `ExportCSV(regions []Region, w io.Writer) error` — columns: id,
  video, start, end, duration, tag, notes
- `ExportJSON(regions []Region, w io.Writer) error` — selected slice
  of the manifest

`internal/playback/auto_advance.go`:

- observe mpv `eof-reached` property
- if enabled in preferences: load next file in session

`frontend/src/lib/KeyframeStrip.svelte`:

- on video load: extract N keyframes (default 50) via ffmpeg
- render as a horizontal strip beneath transport
- click a frame: seek to that timestamp

### verify

- global search finds regions across multiple sessions
- stats panel shows reasonable numbers
- csv export opens in excel
- keyframe strip renders for an open video, click seeks correctly
- auto-advance enabled: video ends, next file loads automatically

phase 6 — sync and multi-machine (v1.1.0)
-----------------------------------------

same captures on multiple machines, manifests stay in sync. single-user
across machines, not multi-user collab.

### files

```
internal/sync/manifest_merge.go         3-way merge for manifest.jsonl
internal/sync/files_state.go            track file hashes, rebase paths
internal/sync/folder.go                 watch a sync folder for changes
internal/sync/tombstone.go              soft-delete entries with deleted_at
docs/schema.md (extend to v2)           add file_hash, deleted_at fields
frontend/src/lib/SyncStatus.svelte      indicator + conflict resolution UI
```

### subtasks

`internal/sync/files_state.go`:

- on add: compute sha256 (or blake3) of first/last MB plus full size
  for a cheap identity hash
- store hash alongside path in `session.json`
- on load: if path missing, scan known sync roots for a file matching
  hash; rebase silently if found

`internal/sync/manifest_merge.go`:

- 3-way merge using last-known + local + remote
- per-region rule: latest `updated_at` wins
- deletions: kept as tombstones with `deleted_at`. tombstones expire
  after N days (default 90)
- conflicts (both sides modified same region after last sync): surface
  to UI, let the user pick

`internal/sync/folder.go`:

- file-watch the configured sync folder (`fsnotify`)
- on manifest change: re-merge into in-memory state, broadcast to UI
- debounce: 500ms after the last write

schema v2 (docs/schema.md):

- add `file_hash` to session.json entries
- add `deleted_at` (optional) to region/bookmark entries

### verify

- log on machine A, sync folder (dropbox/syncthing), open on machine
  B — regions appear, video path rebased
- edit same region on both, last edit wins
- delete on one side — entry tombstoned, gone from UI everywhere

phase 7 — auto-tagging (v1.2.0)
-------------------------------

vision model proposes initial regions for unlogged footage. nothing
commits without a human accept.

### files

```
py/ops/auto_tag.py                      sample frames, classify, propose
internal/region/proposals.go            accept/reject proposed regions
internal/region/proposals_store.go      store proposals separately from real regions
frontend/src/lib/ProposalsPanel.svelte  list with accept/edit/reject
frontend/src/lib/ProposalConfig.svelte  sample interval, confidence threshold
```

### subtasks

`py/ops/auto_tag.py`:

- input: video path, preset (tags + their descriptions), sample
  interval (default 5s), confidence threshold (default 0.6)
- sample frames at interval
- per frame: vision model classifies against preset tags
- merge contiguous same-tag classifications into proposed regions
- output: list of `(tag, start_sec, end_sec, confidence)`

`internal/region/proposals_store.go`:

- separate jsonl: `proposals.jsonl` in session folder
- proposals are not real regions until accepted
- accepted proposal → `region.WriteRegion` + remove from proposals
- rejected proposal → just remove

`frontend/src/lib/ProposalsPanel.svelte`:

- list of proposals with confidence, suggested tag, time range
- per-row buttons: accept, edit (opens RegionEditor), reject
- "accept all above threshold X" bulk action

### verify

- run auto-tag on a sample video → proposals appear
- accept one, reject another — accepted ones become real regions,
  rejected ones disappear
- adjusting confidence threshold re-filters proposals live

phase 8 — EDL + DAW handoff (v1.3.0)
------------------------------------

footage extracts. for a real edit, hand off to a real NLE.

### files

```
internal/export/edl.go                  CMX 3600 EDL writer
internal/export/fcpxml.go               final cut pro XML
internal/export/otio.go                 opentimelineio (optional)
frontend/src/lib/ExportFormatMenu.svelte  format picker
```

### subtasks

`internal/export/edl.go`:

- CMX 3600 format — text, line-oriented, ancient but universally
  supported
- one event per selected region
- reel = video filename, source in/out from `start_sec`/`end_sec`
- record in/out concatenated (events stacked end-to-end)

`internal/export/fcpxml.go`:

- fcpxml 1.10+ (resolve and premiere both import current versions)
- `<asset>` per source video, `<clip>` per region, single sequence
- preserve tag as marker name + color

`internal/export/otio.go`:

- use `opentimelineio` schema (json)
- broadest compat: many NLEs and tools read OTIO now

### verify

- export selected regions as EDL, open in resolve — regions appear as
  clips on a timeline
- same for fcpxml in resolve and premiere
- otio round-trips through otioview

phase 9 — speaker diarization + game detection (v1.4.0)
-------------------------------------------------------

smarter metadata, written back to regions with confidence.

### files

```
py/ops/diarize.py                       speaker labels via pyannote
py/ops/detect_game.py                   HUD/visual classifier
internal/region/auto_fields.go          write speaker, game, confidence
docs/schema.md (extend to v3)           add speaker, game, confidence fields
frontend/src/lib/AutoFieldsBadge.svelte  show inferred fields with confidence
```

### subtasks

`py/ops/diarize.py`:

- input: video + region range
- run pyannote (or whisperx for whisper-aligned diarization)
- output: `[(speaker_label, start_sec, end_sec)]`
- multiple speakers in one region → list of labels with time windows

`py/ops/detect_game.py`:

- HUD detection: sample 3 frames per video, classify against a list
  of known games (start with destiny 2, cyberpunk, noita, then
  expand)
- output: `{"game": "destiny-2", "confidence": 0.92}` or null

`internal/region/auto_fields.go`:

- when an auto field is written: include `confidence` and `source`
  (`"vision"`, `"diarize"`, `"manual"`)
- UI shows auto fields with a confidence badge; manual override
  always wins

schema v3:

- region: optional `speaker` (string or list), `game` (string),
  `confidence` (float per auto field), `source` (per auto field)

### verify

- diarize a multi-speaker clip — speakers labeled in manifest
- detect-game pass — video metadata gets a game tag
- manual override: edit `game` field, source flips to `manual`,
  confidence cleared

stretch (no version yet)
------------------------

- tag heatmaps panel
- mobile companion app (read-only viewer for the manifest)
- keyboard-only full workflow (log → edit → export with zero mouse)
- "explain this clip" — pipeline that runs transcribe + vision +
  describe on a region and writes a one-paragraph summary
