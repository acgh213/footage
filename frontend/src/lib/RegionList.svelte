<script lang="ts">
  import { SeekToEntry, DeleteRegion, ExportRegion } from '../../wailsjs/go/main/App.js'
  import { regionsByFile, entries } from '../stores/session'
  import type { InProgressRegion } from '../stores/session'
  import { timePos, mpvRunning } from '../stores/playback'
  import { createEventDispatcher } from 'svelte'
  import type { region as regionNS } from '../../wailsjs/go/models'
  import RegionEditor from './RegionEditor.svelte'
  import ConfirmDelete from './ConfirmDelete.svelte'

  type Entry = regionNS.Entry
  type Region = regionNS.Region

  export let inProgress: InProgressRegion[] = []

  const dispatch = createEventDispatcher()

  let expandedId: string | null = null
  let confirmDeleteId: string | null = null

  // All closed regions across session (for merge candidates)
  $: allRegions = $entries.filter(e => !!e.region).map(e => e.region!) as Region[]

  function fmt(s: number): string {
    if (s === undefined || s === null) return '--:--'
    const h = Math.floor(s / 3600)
    const m = Math.floor((s % 3600) / 60)
    const sec = (s % 60).toFixed(1)
    if (h > 0) return `${h}:${String(m).padStart(2,'0')}:${sec.padStart(4,'0')}`
    return `${String(m).padStart(2,'0')}:${sec.padStart(4,'0')}`
  }

  function duration(r: Region): string {
    if (!r.end_sec) return '…'
    const d = r.end_sec - r.start_sec
    return d.toFixed(1) + 's'
  }

  function entryId(e: Entry): string {
    return e.region?.id ?? e.bookmark?.id ?? ''
  }

  function toggleExpand(id: string) {
    expandedId = expandedId === id ? null : id
  }

  async function seek(id: string) {
    if (!$mpvRunning) return
    try {
      await SeekToEntry(id)
    } catch (ex) {
      dispatch('error', String(ex))
    }
  }

  function requestDelete(id: string) {
    confirmDeleteId = id
  }

  async function confirmDelete() {
    if (!confirmDeleteId) return
    try {
      await DeleteRegion(confirmDeleteId)
      if (expandedId === confirmDeleteId) expandedId = null
      dispatch('refresh')
    } catch (ex) {
      dispatch('error', String(ex))
    }
    confirmDeleteId = null
  }

  async function exportClip(id: string) {
    try {
      const path = await ExportRegion(id)
      dispatch('status', 'exported → ' + path)
    } catch (ex) {
      dispatch('error', String(ex))
    }
  }
</script>

{#if confirmDeleteId}
  <ConfirmDelete
    on:confirm={confirmDelete}
    on:cancel={() => confirmDeleteId = null}
  />
{/if}

<div class="region-list">
  <div class="list-header">
    <span class="panel-title">regions ({$regionsByFile.length + inProgress.length})</span>
    {#if $regionsByFile.length > 0}
      <button class="batch-btn" on:click={() => dispatch('openBatch')} title="batch export">
        batch export
      </button>
    {/if}
  </div>

  {#if $regionsByFile.length === 0 && inProgress.length === 0}
    <div class="empty">no regions yet — use hotkeys to tag</div>
  {:else}
    <div class="list">
      {#each inProgress as ip}
        <div class="row ip-row" style="--tag-color: {ip.tag_color || '#a8c8e8'};">
          <span class="color-bar"></span>
          <span class="time">{fmt(ip.start_sec)}</span>
          <span class="sep">→</span>
          <span class="time ip-now">{fmt($timePos)}</span>
          <span class="dur ip-dur">({Math.max(0, $timePos - ip.start_sec).toFixed(1)}s)</span>
          <span class="tag-label">{ip.tag_label || ip.tag_key}</span>
          <span class="ip-indicator">● rec</span>
        </div>
      {/each}

      {#each $regionsByFile as e}
        {#if e.region}
          {@const r = e.region}
          {@const expanded = expandedId === r.id}
          <div
            class="row region-row"
            class:expanded
            role="button"
            tabindex="0"
            on:click={() => { seek(r.id); toggleExpand(r.id) }}
            on:keydown={ev => ev.key === 'Enter' && seek(r.id)}
            style="--tag-color: {r.tag_color || '#a8c8e8'};"
          >
            <span class="color-bar"></span>
            <span class="time">{fmt(r.start_sec)}</span>
            <span class="sep">→</span>
            <span class="time">{fmt(r.end_sec)}</span>
            <span class="dur">({duration(r)})</span>
            <span class="tag-label">{r.tag_label || r.tag_key}</span>
            {#if r.notes}
              <span class="notes" title={r.notes}>{r.notes}</span>
            {/if}
            <div class="actions">
              <button on:click|stopPropagation={() => exportClip(r.id)} title="export clip">↗</button>
              <button on:click|stopPropagation={() => requestDelete(r.id)} title="delete" class="del">×</button>
            </div>
          </div>
          {#if expanded}
            <RegionEditor
              region={r}
              {allRegions}
              on:changed={() => { dispatch('refresh'); expandedId = null }}
              on:error={ev => dispatch('error', ev.detail)}
            />
          {/if}
        {:else if e.bookmark}
          <div
            class="row bookmark-row"
            role="button"
            tabindex="0"
            on:click={() => seek(entryId(e))}
            on:keydown={ev => ev.key === 'Enter' && seek(entryId(e))}
          >
            <span class="bookmark-mark">◆</span>
            <span class="time">{fmt(e.bookmark.time_sec)}</span>
            {#if e.bookmark.notes}
              <span class="notes" title={e.bookmark.notes}>{e.bookmark.notes}</span>
            {/if}
            <div class="actions">
              <button on:click|stopPropagation={() => requestDelete(entryId(e))} title="delete" class="del">×</button>
            </div>
          </div>
        {/if}
      {/each}
    </div>
  {/if}
</div>

<style>
  .region-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
    overflow-y: auto;
    flex: 1;
    min-height: 0;
  }

  .list-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-shrink: 0;
  }

  .panel-title {
    color: var(--text-dim);
    font-size: 11px;
    letter-spacing: 0.05em;
  }

  .batch-btn {
    font-size: 11px;
    padding: 2px 6px;
    color: var(--accent);
    border-color: var(--accent);
    opacity: 0.7;
  }

  .batch-btn:hover {
    opacity: 1;
  }

  .list {
    display: flex;
    flex-direction: column;
  }

  .row {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 4px 4px 4px 0;
    border-bottom: 1px solid var(--border);
    cursor: pointer;
    font-size: 12px;
  }

  .row:hover {
    background: var(--bg2);
  }

  .row.expanded {
    background: var(--bg2);
    border-bottom-color: transparent;
  }

  .color-bar {
    width: 3px;
    align-self: stretch;
    background: var(--tag-color, var(--accent));
    flex-shrink: 0;
  }

  .time {
    color: var(--accent);
    font-variant-numeric: tabular-nums;
    min-width: 52px;
  }

  .sep {
    color: var(--text-dim);
  }

  .dur {
    color: var(--text-dim);
    font-size: 11px;
  }

  .tag-label {
    background: color-mix(in srgb, var(--tag-color, var(--accent)) 15%, transparent);
    border: 1px solid color-mix(in srgb, var(--tag-color, var(--accent)) 40%, transparent);
    color: var(--tag-color, var(--accent));
    padding: 0 5px;
    font-size: 11px;
    white-space: nowrap;
  }

  .notes {
    color: var(--text-dim);
    font-size: 11px;
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .bookmark-mark {
    color: var(--accent);
    font-size: 10px;
    margin-left: 3px;
  }

  .bookmark-row {
    opacity: 0.8;
  }

  .actions {
    margin-left: auto;
    display: flex;
    gap: 2px;
    flex-shrink: 0;
  }

  .actions button {
    padding: 0 5px;
    opacity: 0.4;
    font-size: 12px;
  }

  .actions button:hover {
    opacity: 1;
  }

  .del:hover {
    color: var(--red) !important;
  }

  .empty {
    color: var(--text-dim);
    font-size: 12px;
    padding: 8px 0;
  }

  .ip-row {
    opacity: 0.9;
    border-bottom-color: var(--tag-color, var(--accent));
  }

  .ip-now {
    color: var(--text-dim);
  }

  .ip-dur {
    color: var(--tag-color, var(--accent));
  }

  .ip-indicator {
    font-size: 10px;
    color: var(--tag-color, var(--accent));
    margin-left: auto;
    animation: pulse 1s ease-in-out infinite;
    flex-shrink: 0;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.3; }
  }
</style>
