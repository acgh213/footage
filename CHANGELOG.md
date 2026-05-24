changelog
=========

all notable changes to footage. format based on
[keep a changelog](https://keepachangelog.com/en/1.1.0/). footage uses
[semver](https://semver.org/).

unreleased
----------

### added

- project vision, design document, goals, and implementation plan
- manifest schema with stable ULIDs and seconds-as-truth
  ([docs/schema.md](docs/schema.md))
- repo scaffolded at
  [github.com/acgh213/footage](https://github.com/acgh213/footage)

### changed

- docs rewritten in plain lowercase aesthetic. no decorative glyphs,
  no marketing copy. README adds a personal block
- PLAN.md expanded with file-level breakdown, per-file subtasks, and
  post-1.0 phases (sync, auto-tag, EDL, diarization)

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

0.0.0 — 2026-05-24
------------------

### added

- initial repo creation
- README with project vision
- DESIGN.md with architecture, data model, player integration, GUI
  layout, hotkey precision
- GOALS.md with versioned roadmap and non-goals
- PLAN.md with phased implementation details
- docs/schema.md with manifest schema v1
