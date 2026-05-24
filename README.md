# Footage

a logging deck for game footage. watch, tag, annotate, catalog.

you have hours of game captures. you know there's good stuff in there — firefights, clutch moments, cutscenes, dumb glitches — but finding anything later means scrubbing through a timeline blind. Footage is the tool you use to mark what matters while you watch, so you can find it again later.

## what it does

- **log footage with hotkeys.** assign tags to regions as you watch — tap once to start, tap again to stop. firefight. boss encounter. menu. traversal. cutscene. you define the tags
- **tag presets.** save tag groups for your games — a Destiny raid preset, a Cyberpunk combat preset, a Noita run preset. available everywhere, no per-project setup
- **batch export.** after logging, check off the segments you want to extract and remux them to files. no re-encoding, no quality loss
- **searchable catalog.** query your footage by tag, date, game, duration, or free-text notes. type "oryx wipe" and get every segment you marked
- **local LLM integration.** a pop-up overlay that can search the manifest, answer questions about your footage, and drive the UI — not a chat sidebar, a tool that executes commands
- **audio description transcription.** run whisper locally on tagged segments for context-level transcription — enough to know what's being said in a clip, not a full transcript
- **screenshot annotation.** grab frames from tagged regions and feed them through a vision model for visual search

## philosophy

- **not a video editor.** no timeline, no compositing, no effects, no transitions. Footage exports clips via remux (stream copy, zero quality loss) — but it doesn't edit them. open DaVinci if you want to do more than extract segments
- **local-first.** your footage, your tags, your hardware. whisper runs locally. LLM queries run against your own models
- **JSONL manifest.** same battle-tested pattern as the screenshot cataloger. your data is plain text on disk, not locked in a database
- **control surface.** Footage runs alongside your video player — it's a deck, not a viewer. the video plays externally, the logging happens here

## architecture (planned)

```
Footage (GUI)                     mpv/VLC (external player)
┌─────────────────────┐           ┌──────────────────────┐
│  ┌───────────────┐  │   sync    │                      │
│  │  session      │  │◄────────►│  video playback      │
│  │  file list    │  │  (IPC)   │                      │
│  │               │  │           └──────────────────────┘
│  ├───────────────┤  │
│  │  tag panel    │  │
│  │  (hotkeys)    │  │
│  │               │  │
│  ├───────────────┤  │
│  │  region list  │  │
│  │  (log)        │  │
│  │               │  │
│  ├───────────────┤  │
│  │  notes        │  │
│  │               │  │
│  └───────────────┘  │
└─────────────────────┘
```

## naming

named in the tradition of Snow Leopard era Apple: one word, a noun, the thing you work with. you open Footage, and you log your footage.

---

*built for people who have hours of game captures and want to actually find the good parts.*
