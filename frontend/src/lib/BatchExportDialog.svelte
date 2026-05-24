<script lang="ts">
  import { BatchExport, BrowseForExportDir } from '../../wailsjs/go/main/App.js'
  import { entries, activeFile } from '../stores/session'
  import { createEventDispatcher } from 'svelte'
  import type { region as regionNS } from '../../wailsjs/go/models'
  import { EventsOn } from '../../wailsjs/runtime/runtime.js'

  type Region = regionNS.Region

  const dispatch = createEventDispatcher()

  let outDir = ''
  let running = false
  let done = false
  let successCount = 0
  let totalCount = 0

  type Progress = { region_id: string; out_path: string; err: string; done: boolean }
  let progress: Record<string, Progress> = {}

  // All closed regions in the session
  $: allRegions = $entries
    .filter(e => !!e.region)
    .map(e => e.region!) as Region[]

  let selected: Record<string, boolean> = {}

  $: {
    // Default-select all regions for the active file
    const newSel: Record<string, boolean> = {}
    for (const r of allRegions) {
      newSel[r.id] = selected[r.id] ?? (r.video_path === $activeFile)
    }
    selected = newSel
  }

  $: selectedIds = Object.entries(selected).filter(([,v]) => v).map(([k]) => k)

  function fmt(s: number): string {
    const m = Math.floor((s % 3600) / 60)
    const sec = (s % 60).toFixed(1)
    return `${String(m).padStart(2,'0')}:${sec.padStart(4,'0')}`
  }

  function selectAll() { selected = Object.fromEntries(allRegions.map(r => [r.id, true])) }
  function selectNone() { selected = Object.fromEntries(allRegions.map(r => [r.id, false])) }
  function selectByTag(tag: string) {
    selected = Object.fromEntries(allRegions.map(r => [r.id, r.tag_key === tag]))
  }

  $: tags = [...new Set(allRegions.map(r => r.tag_key).filter(Boolean))]

  async function browseDir() {
    try {
      const d = await BrowseForExportDir()
      if (d) outDir = d
    } catch {}
  }

  async function run() {
    if (selectedIds.length === 0) return
    running = true
    done = false
    successCount = 0
    totalCount = selectedIds.length
    progress = {}

    // Listen for progress events
    const unsubscribe = EventsOn('batch-progress', (p: Progress) => {
      progress = { ...progress, [p.region_id]: p }
    })

    try {
      const count = await BatchExport(selectedIds, outDir)
      successCount = count
    } catch (e) {
      dispatch('error', String(e))
    }
    running = false
    done = true
    // @ts-ignore — EventsOn returns an unsubscribe fn
    if (typeof unsubscribe === 'function') unsubscribe()
  }

  function basename(p: string): string {
    return p.replace(/.*[\\/]/, '')
  }
</script>

<div class="overlay" role="dialog" aria-modal="true">
  <div class="dialog">
    <div class="dialog-header">
      <span class="dialog-title">batch export</span>
      <button class="close-btn" on:click={() => dispatch('close')}>×</button>
    </div>

    <div class="controls">
      <div class="select-row">
        <span class="dim">select:</span>
        <button on:click={selectAll}>all</button>
        <button on:click={selectNone}>none</button>
        {#each tags as tag}
          <button on:click={() => selectByTag(tag)}>{tag}</button>
        {/each}
      </div>

      <div class="dir-row">
        <span class="dim">output</span>
        <span class="dir-val" class:placeholder={!outDir}>{outDir || 'auto (exports/ beside each video)'}</span>
        <button on:click={browseDir}>browse</button>
        {#if outDir}
          <button on:click={() => outDir = ''} class="dim">clear</button>
        {/if}
      </div>
    </div>

    <div class="region-list">
      {#each allRegions as r}
        <label class="region-row" class:selected={selected[r.id]}>
          <input type="checkbox" bind:checked={selected[r.id]} />
          <span class="tag-dot" style="background:{r.tag_color || '#a8c8e8'};"></span>
          <span class="time">{fmt(r.start_sec)}–{fmt(r.end_sec ?? 0)}</span>
          <span class="tag">{r.tag_label || r.tag_key}</span>
          <span class="file dim">{basename(r.video_path)}</span>
          {#if progress[r.id]}
            <span class="prog" class:err={!!progress[r.id].err}>
              {progress[r.id].err || '✓'}
            </span>
          {/if}
        </label>
      {/each}
    </div>

    <div class="footer">
      {#if done}
        <span class="result" class:ok={successCount === totalCount}>
          {successCount}/{totalCount} exported
        </span>
      {/if}
      <button
        class="run-btn"
        on:click={run}
        disabled={running || selectedIds.length === 0}
      >
        {running ? `exporting ${Object.keys(progress).length}/${totalCount}…` : `export ${selectedIds.length} clip${selectedIds.length !== 1 ? 's' : ''}`}
      </button>
      <button on:click={() => dispatch('close')} class="dim">close</button>
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: #00000077;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 50;
  }

  .dialog {
    background: var(--bg2);
    border: 1px solid var(--border);
    width: 600px;
    max-width: 95vw;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
  }

  .dialog-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 12px;
    border-bottom: 1px solid var(--border);
    flex-shrink: 0;
  }

  .dialog-title {
    font-size: 13px;
    color: var(--accent);
    letter-spacing: 0.04em;
  }

  .close-btn {
    padding: 0 6px;
    font-size: 16px;
    opacity: 0.5;
  }

  .controls {
    padding: 8px 12px;
    display: flex;
    flex-direction: column;
    gap: 6px;
    border-bottom: 1px solid var(--border);
    flex-shrink: 0;
  }

  .select-row, .dir-row {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
  }

  .dir-val {
    flex: 1;
    font-size: 11px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--green);
  }

  .dir-val.placeholder {
    color: var(--text-dim);
    font-style: italic;
  }

  .region-list {
    overflow-y: auto;
    flex: 1;
    padding: 4px 0;
  }

  .region-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 12px;
    cursor: pointer;
    font-size: 12px;
  }

  .region-row:hover {
    background: var(--bg);
  }

  .region-row.selected {
    background: color-mix(in srgb, var(--accent) 6%, transparent);
  }

  .tag-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .time {
    color: var(--accent);
    min-width: 110px;
    font-variant-numeric: tabular-nums;
  }

  .tag {
    min-width: 100px;
  }

  .file {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 11px;
  }

  .prog {
    font-size: 11px;
    color: var(--green);
  }

  .prog.err {
    color: var(--red);
  }

  .footer {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px;
    border-top: 1px solid var(--border);
    flex-shrink: 0;
  }

  .run-btn {
    color: var(--accent);
    border-color: var(--accent);
    padding: 4px 14px;
  }

  .result {
    font-size: 12px;
    color: var(--red);
    flex: 1;
  }

  .result.ok {
    color: var(--green);
  }

  .dim {
    color: var(--text-dim);
    font-size: 11px;
  }
</style>
