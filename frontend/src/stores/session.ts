import { writable, derived } from 'svelte/store'
import type { session, region, preset as presetNS } from '../../wailsjs/go/models'

type Session = session.Session
type Entry = region.Entry
type Preset = presetNS.Preset

export const currentSession = writable<Session | null>(null)
export const entries = writable<Entry[]>([])
export const openTags = writable<string[]>([])
export const presets = writable<Preset[]>([])
export const activePreset = writable<Preset | null>(null)

// pendingNotes: entryID waiting for notes input after region close
export const pendingNotes = writable<string | null>(null)

export const activeFile = derived(currentSession, $s => {
  if (!$s || !$s.files || $s.files.length === 0) return ''
  const idx = $s.active_idx ?? 0
  return $s.files[idx] ?? ''
})

export const regionsByFile = derived(
  [entries, activeFile],
  ([$entries, $activeFile]) =>
    $entries.filter(e =>
      (e.region?.video_path ?? e.bookmark?.video_path) === $activeFile
    )
)
