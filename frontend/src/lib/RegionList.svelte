<script lang="ts">
  import { SeekToEntry, DeleteRegion, ExportRegion } from '../../wailsjs/go/main/App.js'
  import { regionsByFile } from '../stores/session'
  import type { InProgressRegion } from '../stores/session'
  import { timePos, mpvRunning } from '../stores/playback'
  import { createEventDispatcher } from 'svelte'
  import type { region as regionNS } from '../../wailsjs/go/models'

  type Entry = regionNS.Entry

  export let inProgress: InProgressRegion[] = []

  const dispatch = createEventDispatcher()

  function fmt(s: number): string {
    if (s === undefined || s === null) return '--:--'
    const h = Math.floor(s / 3600)
    const m = Math.floor((s % 3600) / 60)
    const sec = (s % 60).toFixed(1)
    if (h > 0) return `${h}:${String(m).padStart(2,'0')}:${sec.padStart(4,'0')}`
    return `${String(m).padStart(2,'0')}:${sec.padStart(4,'0')}`
  }

  function duration(r: regionNS.Region): string {
    if (!r.end_sec) return '…'
    const d = r.end_sec - r.start_sec
    return d.toFixed(1) + 's'
  }

  function entryId(e: Entry): string {
    return e.region?.id ?? e.bookmark?.id ?? ''
  }

  async function seek(id: string) {
    if (!$mpvRunning) return
    try {
      await SeekToEntry(id)
    } catch (ex) {
      dispatch('error', String(ex))
    }
  }

  async function remove(id: string) {
    try {
      await DeleteRegion(id)
      dispatch('refresh')
    } catch (ex) {
      dispatch('error', String(ex))
    }
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

<div class="region-list">
  <div class="panel-title">regions ({$regionsByFile.length + inProgress.length})</div>

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
          <div
            class="row region-row"
            role="button"
            tabindex="0"
            on:click={() => seek(entryId(e))}
            on:keydown={ev => ev.key === 'Enter' && seek(entryId(e))}
            style="--tag-color: {e.region.tag_color || '#a8c8e8'};"
          >
            <span class="color-bar"></span>
            <span class="time">{fmt(e.region.start_sec)}</span>
            <span class="sep">→</span>
            <span class="time">{fmt(e.region.end_sec)}</span>
            <span class="dur">({duration(e.region)})</span>
            <span class="tag-label">{e.region.tag_label || e.region.tag_key}</span>
            {#if e.region.notes}
              <span class="notes" title={e.region.notes}>{e.region.notes}</span>
            {/if}
            <div class="actions">
              <button on:click|stopPropagation={() => exportClip(entryId(e))} title="export clip">↗</button>
              <button on:click|stopPropagation={() => remove(entryId(e))} title="delete" class="del">×</button>
            </div>
          </div>
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
              <button on:click|stopPropagation={() => remove(entryId(e))} title="delete" class="del">×</button>
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

  .panel-title {
    color: var(--text-dim);
    font-size: 11px;
    letter-spacing: 0.05em;
    flex-shrink: 0;
  }

  .list {
    display: flex;
    flex-direction: column;
    gap: 2px;
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
