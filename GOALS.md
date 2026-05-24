goals
=====

versioned roadmap. each version closes a loop. don't ship a version that
doesn't.

v0.1.0 — prove the deck
-----------------------

watch a video, tag regions with hotkeys, extract a clip. that's the loop.

- [ ] session management. create a session, drag or browse to add video
      files, open an existing session. sessions persist — close footage,
      reopen, pick up where you left off
- [ ] mpv integration. launch bundled mpv from footage, sync playback
      position, transport controls (play, pause, seek)
- [ ] tag presets. define tag groups with hotkeys and colors. save/load
      from disk
- [ ] live tagging. tap a hotkey to open a region, tap again to close.
      one open region per tag per video. regions appear in the list with
      start, end, label, color
- [ ] bookmarks. one-tap bookmark hotkey. `kind: "bookmark"`, not a
      region with a missing end time
- [ ] notes. auto-focus notes field on region close. free-text annotation
- [ ] region list. scrollable, sorted by time, click to seek
- [ ] transport bar. play/pause, seek back 5s, skip ahead 30s, variable
      speed (1x, 1.5x, 2x, 4x)
- [ ] manifest output. all regions and bookmarks saved to
      `manifest.jsonl` per [docs/schema.md](docs/schema.md): stable ULIDs,
      seconds as truth
- [ ] multiple files per session. queue several videos. finish one, move
      to the next
- [ ] single-region export. right-click → export via ffmpeg stream copy.
      fast, lossless, approximate on keyframes

### v0.1.0 validation

before the full logger, validate the scariest seam:

1. wails window opens
2. loads presets
3. user picks a video file
4. mpv launches
5. "get time" button prints current mpv timestamp

that's the "will this even work" check. everything stacks on top.

v0.2.0 — edit and batch
-----------------------

refine the log, get multiple clips out.

- [ ] region editing. adjust start/end (nudge or type). delete. merge
      adjacent regions with the same tag
- [ ] overlap handling. display overlapping regions with indentation
- [ ] batch export UI. checkbox column. select all / by tag / by search.
      export selected to `exports/`

v0.3.0 — talk to your footage
-----------------------------

LLM as a query engine. no python yet — go calls the LLM API directly.

- [ ] LLM pop-up. `Ctrl+L` opens an overlay. type a query, get a
      response, dismiss
- [ ] manifest search. "show me every wipe from the oryx encounter" →
      filtered region list
- [ ] video control. "go to the first boss encounter" → LLM returns a
      seek command, footage executes
- [ ] contextual answers. "what was happening at 14:30?" → LLM reads the
      manifest and answers
- [ ] provider config. openai-compatible endpoint, configurable model,
      local-first

v0.4.0 — deeper intelligence
----------------------------

python backend arrives for heavy lifting.

- [ ] python backend. `footage.py` subprocess, ndjson protocol — same
      pattern as the screenshot cataloger and chisel
- [ ] whisper integration. select regions, queue audio description.
      results write back to manifest
- [ ] screenshot annotation. extract frames from tagged regions, feed
      through vision model, write descriptions back to manifest

v1.0.0 — a real tool
--------------------

polish, export formats, cross-session features.

- [ ] export formats. remux to mp4/mkv, csv export of region list, json
      export of manifest
- [ ] global search. across all sessions. "find every firefight from
      2023"
- [ ] statistics. most-used tags, total logged time, average segment
      duration, busiest logging days
- [ ] auto-advance. option to auto-load the next file when the current
      finishes
- [ ] keyframe strip. visual frame strip for scrubbing reference
- [ ] preferences. default speed, default preset, seek intervals, LLM
      settings

v1.1.0 — sync and multi-machine
-------------------------------

same captures on multiple machines, manifests stay in sync. single-user
across machines, not multi-user collab.

- [ ] content-hash file tracking. videos identified by hash + path, so a
      path can rebase per machine
- [ ] manifest merge. last-write-wins per region by `updated_at`.
      deletions as tombstones
- [ ] sync-folder watcher. dropbox/syncthing-style folder, footage
      notices manifest changes and reloads

v1.2.0 — auto-tagging
---------------------

vision model proposes initial regions for unlogged footage. nothing
commits without a human accept.

- [ ] frame sampler. every N seconds, classify against current preset
- [ ] proposal grouper. merge contiguous same-tag classifications into
      proposed regions
- [ ] proposals panel. accept / edit / reject. accepted ones become
      real regions

v1.3.0 — EDL + DAW handoff
--------------------------

footage extracts. for a real edit, hand off to a real NLE.

- [ ] CMX 3600 EDL export
- [ ] final cut pro XML export (imports into resolve and premiere)
- [ ] opentimelineio export (broader compat)

v1.4.0 — speaker diarization + game detection
---------------------------------------------

smarter metadata.

- [ ] speaker diarization. pyannote-style speaker labels in
      multi-speaker segments
- [ ] game detection. HUD/visual classifier per game, tags the video's
      metadata automatically
- [ ] confidence fields on auto-written values. UI shows uncertainty

stretch
-------

- [ ] tag heatmaps. "this 2-hour session was 40% combat, 25%
      traversal, 15% menus"
- [ ] keyboard-only workflow. no mouse for the full log → edit →
      export loop

non-goals
---------

- not a video editor. no timeline, no compositing, no transitions, no
  effects, no rendering. footage remuxes clips out — it doesn't edit
- not a media player replacement. footage doesn't play video — it
  controls mpv
- not a streaming tool. no broadcast, no obs integration, no live
  capture
- not a collaboration tool. single-user, local-first. share the
  manifest if you want, but footage doesn't sync (across machines, yes
  — across users, no)
- not a general media cataloger. this is for game footage. it might
  work for other things, but that's not the target
