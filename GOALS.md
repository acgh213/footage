# ✧ Footage — goals & roadmap ✧

## v0.1.0 — "prove the deck"

the goal: watch a video, tag regions with hotkeys, extract a clip. close the loop.

- [ ] **session management.** create a session, drag or browse to add video files, open an existing session. sessions persist — close Footage, reopen, pick up where you left off
- [ ] **mpv integration.** launch bundled mpv from Footage, sync playback position, transport controls (play, pause, seek)
- [ ] **tag presets.** define tag groups with hotkeys and colors. save/load presets from disk
- [ ] **live tagging.** tap hotkey to open a region, tap again to close. one open region per tag per video. regions appear in the list with start time, end time, tag label, color
- [ ] **bookmarks.** one-tap bookmark hotkey. `kind: "bookmark"`, not a region with missing end time
- [ ] **notes.** auto-focus notes field when closing a region. free-text annotation
- [ ] **region list.** scrollable list of tagged regions, sorted by time. click to jump video to that region
- [ ] **transport bar.** play/pause, seek back 5s, skip ahead 30s, variable speed (1x, 1.5x, 2x, 4x)
- [ ] **manifest output.** all tagged regions and bookmarks saved to `manifest.jsonl`. schema per [docs/schema.md](docs/schema.md): stable ULIDs, seconds as truth
- [ ] **multiple files per session.** queue up several videos in a session. finish one, move to the next
- [ ] **single-region export.** right-click a region → export clip via ffmpeg stream copy (fast, lossless, approximate keyframe accuracy)

### v0.1.0 validation

the first runnable app is even smaller than v0.1.0. before building the full logger, validate the scariest seam:

1. Wails window opens
2. Loads presets
3. User picks a video file
4. mpv launches
5. "Get Time" button prints current mpv timestamp

that's the "will this even work" check. everything stacks on top.

## v0.2.0 — "edit and batch"

refine the log, get multiple clips out.

- [ ] **region editing.** adjust start/end times after tagging (nudge or type). delete regions. merge adjacent regions with the same tag
- [ ] **overlap handling.** display overlapping regions in the list with indentation or nesting
- [ ] **batch export UI.** checkbox column in region list. select all / select by tag / select by search. export selected to `exports/`

## v0.3.0 — "talk to your footage"

the LLM becomes a query engine. no Python backend yet — Go calls the LLM API directly for manifest search.

- [ ] **LLM pop-up.** `Ctrl+L` opens a pop-up overlay. type a query, get a response, dismiss
- [ ] **manifest search.** "show me every wipe from the Oryx encounter" → filtered region list
- [ ] **video control.** "go to the first boss encounter" → LLM returns a seek command, Footage executes it
- [ ] **contextual answers.** "what was happening at 14:30?" → LLM reads the manifest and answers
- [ ] **provider config.** OpenAI-compatible endpoint, configurable model, local-first

## v0.4.0 — "deeper intelligence"

Python backend arrives for heavy lifting.

- [ ] **Python backend.** `footage.py` subprocess with NDJSON protocol — same pattern as screenshot cataloger and chisel
- [ ] **whisper integration.** select regions, queue audio description transcription. results write back to manifest
- [ ] **screenshot annotation.** extract frames from tagged regions, feed through vision model, write descriptions back to manifest

## v1.0.0 — "a real tool"

polish, export formats, and cross-session features.

- [ ] **export formats.** remux to `.mp4` / `.mkv`, export region list as CSV, export manifest as JSON
- [ ] **global search.** search across all sessions. "find every firefight from 2023"
- [ ] **statistics.** most-used tags, total logged time, average segment duration, busiest logging days
- [ ] **auto-advance.** option to auto-load the next file in the session when the current one finishes
- [ ] **keyframe strip.** visual strip of frames from the current video for scrubbing reference
- [ ] **preferences.** default speed, default preset, seek intervals, LLM settings

## v1.1+ — "stretch goals"

- [ ] **auto-tagging with vision.** batch process unlogged footage through vision model for initial region suggestions
- [ ] **speaker diarization.** identify and label different speakers in multi-speaker segments
- [ ] **game detection.** auto-detect which game a video is from based on HUD and visual patterns
- [ ] **tag statistics and heatmaps.** "this 2-hour session was 40% combat, 25% traversal, 15% menus"
- [ ] **export to edit decision list (EDL).** export region data in a format DaVinci Resolve / Premiere can import

## non-goals

- **not a video editor.** no timeline, no compositing, no transitions, no effects, no rendering. Footage extracts clips via remux — it doesn't edit them
- **not a media player replacement.** Footage doesn't play video — it controls mpv. the video window is mpv
- **not a streaming tool.** no broadcast, no OBS integration, no live capture
- **not a collaboration tool.** single-user, local-first. share the manifest if you want, but Footage doesn't sync
- **not a general media cataloger.** this is for game footage. it might work for other things, but that's not the design target
