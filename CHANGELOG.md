# changelog

all notable changes to Footage will be documented in this file.

format based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Footage uses [Semantic Versioning](https://semver.org/).

---

## [unreleased]

### added
- project vision, design document, goals, and implementation plan
- manifest schema with stable ULIDs and seconds-as-truth ([docs/schema.md](docs/schema.md))
- repo scaffolded at [github.com/acgh213/footage](https://github.com/acgh213/footage)

---

## versioning convention

| version | what ships |
|---------|-----------|
| 0.1.0 | logging core — session, mpv sync, live tagging, bookmarks, single-region export |
| 0.2.0 | editing + batch export — region editing, merge, batch export UI |
| 0.3.0 | LLM query — pop-up overlay, manifest search, video control commands |
| 0.4.0 | Python backend — whisper audio description, vision screenshot annotation |
| 1.0.0 | polish — global search, statistics, export formats, keyframe strip |

---

## [0.0.0] — 2026-05-24

### added
- initial repo creation
- README with project vision
- DESIGN.md with architecture, data model, player integration, GUI layout, hotkey precision
- GOALS.md with versioned roadmap and non-goals
- PLAN.md with phased implementation details
- docs/schema.md with manifest schema v1
