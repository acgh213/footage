# ✧ Footage — goals & roadmap ✧

## short-term (v0.1 — "log a session")

the goal: watch a video, tag regions with hotkeys, see them in a list. basic but usable.

- [ ] **session management.** create a session, drag or browse to add video files, open an existing session
- [ ] **mpv integration.** launch mpv from Footage, sync playback position, transport controls (play, pause, seek)
- [ ] **tag presets.** define tag groups with hotkeys and colors. save/load presets from disk
- [ ] **live tagging.** tap hotkey to start a region, tap again to close. regions appear in the list with start time, end time, tag label, color
- [ ] **notes.** auto-focus notes field when closing a region. free-text annotation
- [ ] **region list.** scrollable list of tagged regions, sorted by time. click to jump video to that region
- [ ] **transport bar.** play/pause, seek back 5s, skip ahead 30s, variable speed (1x, 1.5x, 2x, 4x)
- [ ] **bookmarks.** one-tap bookmark hotkey. no region to close. just marks a timestamp
- [ ] **manifest output.** all tagged regions saved to `manifest.jsonl` in the session folder
- [ ] **multiple files per session.** queue up several videos in a session. finish one, move to the next

## medium-term (v0.2 — "make it useful")

batch processing arrives. the manifest becomes searchable.

- [ ] **batch export.** check off regions in the list, click Export — remux selected segments to `exports/` using ffmpeg stream copy (no re-encode)
- [ ] **region editing.** adjust start/end times after tagging. delete regions. merge adjacent regions with the same tag
- [ ] **overlap handling.** display overlapping regions in the list with indentation or nesting
- [ ] **persistent sessions.** close Footage, reopen next week, the session is where you left it — same files, same manifest, in-progress regions preserved
- [ ] **Python backend.** `footage.py` subprocess for whisper, vision, and LLM operations
- [ ] **whisper integration.** select regions, queue audio description transcription. results write back to manifest
- [ ] **screenshot annotation.** extract frames from tagged regions, feed through vision model, write descriptions back to manifest

## medium-term (v0.3 — "talk to your footage")

the LLM becomes a query engine.

- [ ] **LLM pop-up.** `Ctrl+L` opens a pop-up overlay. type a query, get a response, dismiss
- [ ] **manifest search.** "show me every wipe from the Oryx encounter" → filtered region list
- [ ] **video control.** "go to the first boss encounter" → LLM returns a seek command, Footage executes it
- [ ] **contextual answers.** "what was happening at 14:30?" → LLM reads the manifest and answers
- [ ] **provider config.** same pattern as chisel — OpenAI-compatible endpoint, configurable model, local-first

## long-term (v1.0 — "a real tool")

polish, export formats, and cross-session features.

- [ ] **export formats.** remux to `.mp4` / `.mkv`, export region list as CSV, export manifest as JSON for external tools
- [ ] **global search.** search across all sessions, not just the current one. "find every firefight from 2023"
- [ ] **statistics.** most-used tags, total logged time, average segment duration, busiest logging days
- [ ] **auto-advance.** option to auto-load the next file in the session when the current one finishes
- [ ] **keyframe strip.** visual strip of frames from the current video for scrubbing reference
- [ ] **preferences.** default speed, default preset, seek intervals, mpv path, LLM settings

## long-term (v1.1+ — "stretch goals")

- [ ] **auto-tagging with vision.** batch process unlogged footage through vision model for initial region suggestions
- [ ] **speaker diarization.** identify and label different speakers in multi-speaker segments
- [ ] **game detection.** auto-detect which game a video is from based on HUD and visual patterns
- [ ] **tag statistics and heatmaps.** "this 2-hour session was 40% combat, 25% traversal, 15% menus"
- [ ] **export to edit decision list (EDL).** export region data in a format DaVinci Resolve / Premiere can import
- [ ] **VLC fallback.** full VLC HTTP API support for users without mpv

## non-goals

- **not a video editor.** no timeline, no compositing, no transitions, no effects, no rendering
- **not a media player replacement.** Footage doesn't play video — it controls a player. the video window is mpv or VLC
- **not a streaming tool.** no broadcast, no OBS integration, no live capture
- **not a collaboration tool.** single-user, local-first. share the manifest if you want, but Footage doesn't sync
- **not a general media cataloger.** this is for game footage. it might work for other things, but that's not the design target
