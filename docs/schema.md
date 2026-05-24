# manifest schema

the manifest is a JSONL file (one JSON object per line) in the session folder.

each line is a region, a bookmark, or a session metadata entry. the `kind` field distinguishes them.

## version

current schema version: **1**

## region

a tagged segment in a video.

```json
{
  "schema_version": 1,
  "kind": "region",
  "id": "region_01JM8XK4Y7Q0A1B2C3D4E5F6G7H8",
  "session_id": "session_01JM8XK4Y7Q0A1B2C3D4E5F6G7H8",
  "video_path": "D:/Captures/destiny2_raid_2021.mp4",
  "start_sec": 192.5,
  "end_sec": 347.2,
  "display_start": "00:03:12.500",
  "display_end": "00:05:47.200",
  "tag": "boss encounter",
  "preset": "destiny-raid",
  "notes": "first Oryx clear. wiped at final stand once. Octavio was screaming.",
  "created_at": "2026-06-01T14:22:00.500Z",
  "updated_at": "2026-06-01T14:22:00.500Z"
}
```

### fields

| field | type | required | description |
|-------|------|:--------:|-------------|
| `schema_version` | int | ✓ | manifest schema version (currently `1`) |
| `kind` | string | ✓ | `"region"` or `"bookmark"` |
| `id` | string | ✓ | stable ULID, globally unique within the session |
| `session_id` | string | ✓ | session ULID that owns this entry |
| `video_path` | string | ✓ | absolute path to the source video file |
| `start_sec` | float | ✓ | start time in seconds (source of truth) |
| `end_sec` | float | ✓ | end time in seconds (source of truth) |
| `display_start` | string | ✓ | human-readable timestamp, derived from `start_sec` |
| `display_end` | string | ✓ | human-readable timestamp, derived from `end_sec` |
| `tag` | string | ✓ | freeform tag label |
| `preset` | string | ✓ | name of the tag preset used |
| `notes` | string | | freeform notes |
| `created_at` | string | ✓ | ISO 8601 with milliseconds, UTC |
| `updated_at` | string | ✓ | ISO 8601 with milliseconds, UTC |

### rules

- **seconds are truth.** `start_sec` and `end_sec` are the authoritative time values. `display_start` and `display_end` are derived display sugar. mpv and ffmpeg consume seconds. never parse display timestamps for logic.
- **IDs are ULIDs.** use `01JM8XK4Y7Q...` format. sortable by creation time, globally unique. no auto-increment integers.
- **regions can overlap.** two regions in the same video with overlapping time ranges are valid. the UI handles nesting display.
- **one open region per tag per video.** during logging, only one region of a given tag can be open at a time in a single video. pressing the tag hotkey toggles that tag's open region.
- **`updated_at` updates on any edit.** adjusting times, editing notes, merging regions — all update `updated_at`.

## bookmark

a timestamp marker. no end time, no region.

```json
{
  "schema_version": 1,
  "kind": "bookmark",
  "id": "bookmark_01JM8XK4Y7Q0A1B2C3D4E5F6G7H8",
  "session_id": "session_01JM8XK4Y7Q0A1B2C3D4E5F6G7H8",
  "video_path": "D:/Captures/destiny2_raid_2021.mp4",
  "time_sec": 420.0,
  "display_time": "00:07:00.000",
  "notes": "check this moment — weird physics glitch?",
  "created_at": "2026-06-01T14:30:00.000Z",
  "updated_at": "2026-06-01T14:30:00.000Z"
}
```

### differences from regions

- **no `end_sec`** — bookmarks are a single point in time
- **no `tag`** — bookmarks aren't tagged. they're just "come back to this"
- **no `preset`** — bookmarks are preset-independent
- **`time_sec`** replaces `start_sec`/`end_sec`
- **`display_time`** replaces `display_start`/`display_end`

## file lifecycle

- **append-only during logging.** each region close or bookmark creation appends one line to `manifest.jsonl`. no in-place edits while logging
- **rewrite on edits.** adjusting times, deleting, merging — the entire file is rewritten atomically (write to temp → rename)
- **no event log in v0.1.** the manifest is current state. an `events.jsonl` audit log may be added later if revision history is needed
