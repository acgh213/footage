# changelog

all notable changes to Footage will be documented in this file.

format based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Footage uses [Semantic Versioning](https://semver.org/).

---

## [unreleased]

### added
- project vision, design document, goals, and implementation plan
- repo scaffolded at [github.com/acgh213/footage](https://github.com/acgh213/footage)

---

## versioning convention

| phase | version | what ships |
|-------|---------|-----------|
| scaffolding | 0.0.1 | project creation, config, preset I/O |
| logging core | 0.1.0 | session, mpv sync, live tagging, region list, notes |
| editing + export | 0.2.0 | region editing, batch remux export |
| Python backend | 0.2.0 | whisper transcription, vision annotation |
| LLM query | 0.3.0 | pop-up query engine, manifest search, UI commands |
| polish | 1.0.0 | global search, statistics, export formats |

---

## [0.0.0] — 2026-05-24

### added
- initial repo creation
- README with project vision
- DESIGN.md with architecture, data model, player integration, GUI layout
- GOALS.md with short/medium/long-term roadmap and non-goals
- PLAN.md with phased implementation details
