footage
=======

a logging deck for game footage. tag regions with hotkeys, batch-export clips,
search the catalog later.

what
----

- log with hotkeys. tap to open a region, tap the same key to close it. one
  open region per tag per video. different tags can have simultaneous regions
- tag presets, global. a destiny raid preset, a cyberpunk combat preset, a
  noita run preset. saved once, available everywhere
- batch export via ffmpeg stream copy. fast, lossless, approximate on
  keyframes. log first, select, export later
- searchable catalog. tag, date, game, duration, free-text notes
- local LLM pop-up on Ctrl+L. drives the UI through tool calls, not a chat
  sidebar
- audio description transcription with whisper (v0.4.0+). context-level, not
  word-for-word
- screenshot annotation via vision model on extracted frames (v0.4.0+)

why
---

hours of captures. good stuff buried in them. scrubbing blind is bad. mark
it while you watch.

not
---

- not a video editor. no timeline, no compositing, no transitions. footage
  remuxes clips out, open davinci if you want to actually edit
- not a media player. footage controls mpv externally. the video window is
  mpv's window
- not a streaming or capture tool. no obs, no broadcast
- not a collaboration tool. single-user, local-first

architecture
------------

footage is a wails app (go + webview) that drives bundled mpv over a JSON
IPC named pipe (`\\.\pipe\footage-mpv`). manifest is JSONL on disk in the
session folder. python backend (`footage.py`, v0.4.0+) runs whisper and
vision via an ndjson subprocess.

```
footage (GUI, wails)              mpv (external, bundled)
+--------------------+           +----------------------+
|  session           |   sync    |                      |
|  file list         |<--------->|  video playback      |
|                    |   (IPC)   |                      |
+--------------------+           +----------------------+
|  tag panel         |
|  (hotkeys)         |
+--------------------+
|  region list       |
+--------------------+
|  notes             |
+--------------------+
```

docs
----

- [design.md](DESIGN.md)         architecture, data model, GUI, decisions
- [goals.md](GOALS.md)           versioned roadmap, non-goals
- [plan.md](PLAN.md)             phased implementation breakdown
- [changelog.md](CHANGELOG.md)   version history
- [docs/schema.md](docs/schema.md)  manifest schema v1
- [claude.md](CLAUDE.md)         instructions for AI coding assistants

cassie
======

she/her. builds tools for her own game footage and screenshot piles.
footage is one of them.

links
-----

- omg          [starsetbyte.lol](https://starsetbyte.lol)
- github       [github.com/acgh213](https://github.com/acgh213)

keys
----

- pgp          `F33D 2208 15F3 C29C 9AA5 08A4 E330 0B05 35DD FAFC`
