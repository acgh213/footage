<script lang="ts">
  import { NudgeRegion, SetRegionTime, MergeRegions } from '../../wailsjs/go/main/App.js'
  import { createEventDispatcher } from 'svelte'
  import type { region as regionNS } from '../../wailsjs/go/models'

  type Region = regionNS.Region

  export let region: Region
  export let allRegions: Region[] = []

  const dispatch = createEventDispatcher()

  let editingField: 'start' | 'end' | null = null
  let editValue = ''

  function fmt(s: number): string {
    if (!s && s !== 0) return '--:--'
    const h = Math.floor(s / 3600)
    const m = Math.floor((s % 3600) / 60)
    const sec = (s % 60).toFixed(3)
    if (h > 0) return `${h}:${String(m).padStart(2,'0')}:${sec.padStart(6,'0')}`
    return `${String(m).padStart(2,'0')}:${sec.padStart(6,'0')}`
  }

  function parseTime(s: string): number | null {
    // Accept HH:MM:SS.mmm, MM:SS.mmm, or raw seconds
    const parts = s.trim().split(':')
    if (parts.length === 1) {
      const v = parseFloat(parts[0])
      return isNaN(v) ? null : v
    }
    if (parts.length === 2) {
      const m = parseInt(parts[0])
      const sec = parseFloat(parts[1])
      if (isNaN(m) || isNaN(sec)) return null
      return m * 60 + sec
    }
    if (parts.length === 3) {
      const h = parseInt(parts[0])
      const m = parseInt(parts[1])
      const sec = parseFloat(parts[2])
      if (isNaN(h) || isNaN(m) || isNaN(sec)) return null
      return h * 3600 + m * 60 + sec
    }
    return null
  }

  async function nudge(field: 'start' | 'end', delta: number) {
    try {
      await NudgeRegion(region.id, field, delta)
      dispatch('changed')
    } catch (e) {
      dispatch('error', String(e))
    }
  }

  function startEdit(field: 'start' | 'end') {
    editingField = field
    editValue = fmt(field === 'start' ? region.start_sec : region.end_sec ?? 0)
  }

  async function commitEdit() {
    if (!editingField) return
    const v = parseTime(editValue)
    if (v === null) {
      dispatch('error', 'invalid time format')
      editingField = null
      return
    }
    try {
      await SetRegionTime(region.id, editingField, v)
      dispatch('changed')
    } catch (e) {
      dispatch('error', String(e))
    }
    editingField = null
  }

  function cancelEdit() {
    editingField = null
  }

  // Candidates for merge: same tag, same video, different id
  $: mergeCandidates = allRegions.filter(
    r => r.id !== region.id && r.tag_key === region.tag_key && r.video_path === region.video_path
  )

  async function merge(otherId: string) {
    try {
      await MergeRegions(region.id, otherId)
      dispatch('changed')
    } catch (e) {
      dispatch('error', String(e))
    }
  }
</script>

<div class="editor">
  <div class="field-row">
    <span class="field-label">start</span>
    {#if editingField === 'start'}
      <input
        class="time-input"
        bind:value={editValue}
        on:keydown={e => e.key === 'Enter' ? commitEdit() : e.key === 'Escape' && cancelEdit()}
        on:blur={commitEdit}
        autofocus
      />
    {:else}
      <span class="time-val" on:click={() => startEdit('start')} role="button" tabindex="0" on:keydown={e => e.key === 'Enter' && startEdit('start')}>
        {fmt(region.start_sec)}
      </span>
    {/if}
    <div class="nudge-btns">
      <button on:click={() => nudge('start', -0.5)} title="-0.5s">-½</button>
      <button on:click={() => nudge('start', -0.1)} title="-0.1s">-</button>
      <button on:click={() => nudge('start', 0.1)} title="+0.1s">+</button>
      <button on:click={() => nudge('start', 0.5)} title="+0.5s">+½</button>
    </div>
  </div>

  <div class="field-row">
    <span class="field-label">end</span>
    {#if editingField === 'end'}
      <input
        class="time-input"
        bind:value={editValue}
        on:keydown={e => e.key === 'Enter' ? commitEdit() : e.key === 'Escape' && cancelEdit()}
        on:blur={commitEdit}
        autofocus
      />
    {:else}
      <span class="time-val" on:click={() => startEdit('end')} role="button" tabindex="0" on:keydown={e => e.key === 'Enter' && startEdit('end')}>
        {fmt(region.end_sec ?? 0)}
      </span>
    {/if}
    <div class="nudge-btns">
      <button on:click={() => nudge('end', -0.5)} title="-0.5s">-½</button>
      <button on:click={() => nudge('end', -0.1)} title="-0.1s">-</button>
      <button on:click={() => nudge('end', 0.1)} title="+0.1s">+</button>
      <button on:click={() => nudge('end', 0.5)} title="+0.5s">+½</button>
    </div>
  </div>

  {#if mergeCandidates.length > 0}
    <div class="merge-row">
      <span class="field-label">merge</span>
      <div class="merge-list">
        {#each mergeCandidates as other}
          <button class="merge-btn" on:click={() => merge(other.id)} title="merge with this region">
            {fmt(other.start_sec)}–{fmt(other.end_sec ?? 0)}
          </button>
        {/each}
      </div>
    </div>
  {/if}
</div>

<style>
  .editor {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 8px;
    background: var(--bg);
    border: 1px solid var(--border);
    border-top: none;
  }

  .field-row {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .field-label {
    color: var(--text-dim);
    font-size: 11px;
    width: 36px;
    flex-shrink: 0;
  }

  .time-val {
    color: var(--accent);
    font-size: 12px;
    min-width: 80px;
    cursor: pointer;
    border-bottom: 1px dashed var(--border);
    padding: 1px 2px;
  }

  .time-val:hover {
    border-color: var(--accent);
  }

  .time-input {
    width: 100px;
    font-size: 12px;
    padding: 2px 4px;
    color: var(--accent);
    background: var(--bg2);
    border: 1px solid var(--accent);
  }

  .nudge-btns {
    display: flex;
    gap: 2px;
  }

  .nudge-btns button {
    padding: 1px 5px;
    font-size: 11px;
  }

  .merge-row {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
  }

  .merge-list {
    display: flex;
    gap: 4px;
    flex-wrap: wrap;
  }

  .merge-btn {
    font-size: 11px;
    padding: 2px 6px;
    color: var(--text-dim);
    border-color: var(--border);
  }

  .merge-btn:hover {
    color: var(--accent);
    border-color: var(--accent);
  }
</style>
