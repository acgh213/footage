<script lang="ts">
  import {
    GetSession, GetPresets, GetMPVPath, BrowseForMPV,
    GetTimePos, StopPlayer, GetOpenTags, GetRegions,
    GetInProgressRegions, OpenFile
  } from '../wailsjs/go/main/App.js'
  import type { preset as presetNS } from '../wailsjs/go/models'
  import SessionPanel from './lib/SessionPanel.svelte'
  import TagPanel from './lib/TagPanel.svelte'
  import RegionList from './lib/RegionList.svelte'
  import NotesField from './lib/NotesField.svelte'
  import Transport from './lib/Transport.svelte'
  import BatchExportDialog from './lib/BatchExportDialog.svelte'
  import {
    currentSession, presets, activePreset, entries, openTags, inProgressRegions
  } from './stores/session'
  import {
    timePos, mpvRunning, paused, speed, startPolling, stopPolling
  } from './stores/playback'

  type Preset = presetNS.Preset

  let status = ''
  let statusOk = true
  let mpvPath = ''
  let showBatch = false

  function setStatus(msg: string, ok: boolean) {
    status = msg
    statusOk = ok
  }

  async function init() {
    try {
      mpvPath = await GetMPVPath()
    } catch {}
    try {
      const s = await GetSession()
      currentSession.set(s)
    } catch {}
    try {
      const ps = await GetPresets()
      presets.set(ps)
      if (ps.length > 0 && !$activePreset) {
        activePreset.set(ps[0])
      }
    } catch {}
    await refreshEntries()
    await refreshOpenTags()
  }

  async function refreshEntries() {
    try {
      const e = await GetRegions()
      entries.set(e ?? [])
    } catch {}
  }

  async function refreshOpenTags() {
    try {
      const t = await GetOpenTags()
      openTags.set(t ?? [])
      const ip = await GetInProgressRegions()
      inProgressRegions.set(ip ?? [])
    } catch {}
  }

  function fmt(s: number): string {
    const h = Math.floor(s / 3600)
    const m = Math.floor((s % 3600) / 60)
    const sec = (s % 60).toFixed(1)
    if (h > 0) return `${h}:${String(m).padStart(2,'0')}:${sec.padStart(4,'0')}`
    return `${String(m).padStart(2,'0')}:${sec.padStart(4,'0')}`
  }

  async function browseMPV() {
    try {
      const p = await BrowseForMPV()
      if (p) {
        mpvPath = p
        setStatus('mpv path set: ' + p, true)
      }
    } catch (e) {
      setStatus(String(e), false)
    }
  }

  function onSessionOpened() {
    mpvRunning.set(true)
    paused.set(false)
    speed.set(1)
    startPolling(async () => {
      const t = await GetTimePos()
      timePos.set(t)
      return t
    })
    refreshEntries()
    refreshOpenTags()
    setStatus('mpv running', true)
  }

  function onStop() {
    stopPolling()
    mpvRunning.set(false)
    timePos.set(0)
    openTags.set([])
    setStatus('mpv stopped', true)
  }

  async function stop() {
    try {
      await StopPlayer()
      onStop()
    } catch (e) {
      setStatus(String(e), false)
    }
  }

  function onPresetChange(e: Event) {
    const name = (e.target as HTMLSelectElement).value
    activePreset.set($presets.find(p => p.name === name) ?? null)
  }

  init()
</script>

{#if showBatch}
  <BatchExportDialog
    on:close={() => showBatch = false}
    on:error={e => setStatus(e.detail, false)}
  />
{/if}

<main>
  <header>
    <span class="title">footage</span>
    <span class="subtitle">v0.1 — logging core</span>
    <div class="header-right">
      <span class="mpv-path" class:not-found={!mpvPath} title={mpvPath || 'mpv not found'}>
        {mpvPath ? 'mpv ✓' : 'mpv not found'}
      </span>
      {#if !mpvPath}
        <button on:click={browseMPV} class="small-btn">locate…</button>
      {/if}
    </div>
  </header>

  <div class="layout">
    <!-- left column: session + tag panel -->
    <div class="left-col">
      <section class="panel">
        <SessionPanel
          on:error={e => setStatus(e.detail, false)}
          on:opened={onSessionOpened}
        />
      </section>

      <section class="panel">
        <div class="preset-row">
          <label class="dim-label">preset</label>
          <select
            value={$activePreset?.name ?? ''}
            on:change={onPresetChange}
          >
            {#each $presets as p}
              <option value={p.name}>{p.name}</option>
            {/each}
          </select>
        </div>
        <TagPanel
          on:refresh={() => { refreshEntries(); refreshOpenTags() }}
          on:refreshTags={refreshOpenTags}
          on:error={e => setStatus(e.detail, false)}
        />
      </section>

      <section class="panel transport-panel">
        <Transport />
        <button
          on:click={stop}
          disabled={!$mpvRunning}
          class="stop-btn"
        >stop mpv</button>
      </section>
    </div>

    <!-- right column: region list + notes -->
    <div class="right-col">
      {#if $inProgressRegions.length > 0}
        <div class="recording-banner">
          {#each $inProgressRegions as ip}
            <span class="recording-tag" style="--tag-color: {ip.tag_color};">
              <span class="rec-dot">●</span>
              <span class="rec-key">{ip.tag_key}</span>
              {ip.tag_label}
              <span class="rec-time">{fmt($timePos - ip.start_sec)}s</span>
            </span>
          {/each}
        </div>
      {/if}
      <NotesField
        on:saved={refreshEntries}
        on:error={e => setStatus(e.detail, false)}
      />
      <section class="panel region-panel">
        <RegionList
          inProgress={$inProgressRegions}
          on:refresh={refreshEntries}
          on:status={e => setStatus(e.detail, true)}
          on:error={e => setStatus(e.detail, false)}
          on:openBatch={() => showBatch = true}
        />
      </section>
    </div>
  </div>

  {#if status}
    <div class="status" class:ok={statusOk} class:err={!statusOk}>
      {status}
    </div>
  {/if}

  <footer>
    <span class:running={$mpvRunning} class:stopped={!$mpvRunning}>
      mpv: {$mpvRunning ? 'running' : 'not running'}
    </span>
    {#if $currentSession}
      <span class="dim">  ·  {$currentSession.name}</span>
    {/if}
  </footer>
</main>

<style>
  main {
    display: flex;
    flex-direction: column;
    height: 100vh;
    padding: 12px;
    gap: 10px;
    overflow: hidden;
  }

  header {
    display: flex;
    align-items: baseline;
    gap: 10px;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--border);
    flex-shrink: 0;
  }

  .title {
    font-size: 16px;
    color: var(--accent);
    letter-spacing: 0.05em;
  }

  .subtitle {
    color: var(--text-dim);
    font-size: 11px;
  }

  .header-right {
    margin-left: auto;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .mpv-path {
    font-size: 11px;
    color: var(--green);
  }

  .mpv-path.not-found {
    color: var(--red);
  }

  .small-btn {
    font-size: 11px;
    padding: 2px 6px;
  }

  .layout {
    display: flex;
    gap: 10px;
    flex: 1;
    min-height: 0;
    overflow: hidden;
  }

  .left-col {
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: 280px;
    flex-shrink: 0;
    overflow-y: auto;
  }

  .right-col {
    display: flex;
    flex-direction: column;
    gap: 8px;
    flex: 1;
    min-width: 0;
    overflow: hidden;
  }

  .panel {
    background: var(--bg2);
    border: 1px solid var(--border);
    padding: 10px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .region-panel {
    flex: 1;
    min-height: 0;
    overflow: hidden;
  }

  .transport-panel {
    flex-direction: column;
    gap: 8px;
  }

  .preset-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .dim-label {
    color: var(--text-dim);
    font-size: 11px;
    white-space: nowrap;
  }

  select {
    flex: 1;
    background: var(--bg);
    color: var(--text);
    border: 1px solid var(--border);
    padding: 3px 6px;
    font-family: inherit;
    font-size: 12px;
  }

  .stop-btn {
    font-size: 11px;
    color: var(--red);
    border-color: var(--red);
    opacity: 0.7;
  }

  .stop-btn:hover:not(:disabled) {
    opacity: 1;
  }

  .recording-banner {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
    flex-shrink: 0;
  }

  .recording-tag {
    display: flex;
    align-items: center;
    gap: 5px;
    padding: 4px 10px;
    border: 1px solid var(--tag-color, var(--accent));
    background: color-mix(in srgb, var(--tag-color, var(--accent)) 12%, transparent);
    font-size: 12px;
  }

  .rec-dot {
    color: var(--tag-color, var(--accent));
    font-size: 9px;
    animation: pulse 1s ease-in-out infinite;
  }

  .rec-key {
    background: var(--bg);
    border: 1px solid var(--border);
    padding: 0 4px;
    font-size: 11px;
    color: var(--text-dim);
  }

  .rec-time {
    color: var(--tag-color, var(--accent));
    font-size: 11px;
    font-variant-numeric: tabular-nums;
    margin-left: 4px;
    opacity: 0.8;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.3; }
  }

  .status {
    padding: 5px 10px;
    border-left: 3px solid;
    font-size: 12px;
    flex-shrink: 0;
  }

  .status.ok {
    border-color: var(--green);
    color: var(--green);
    background: #8cc4a011;
  }

  .status.err {
    border-color: var(--red);
    color: var(--red);
    background: #e8919e11;
  }

  footer {
    padding-top: 6px;
    border-top: 1px solid var(--border);
    font-size: 11px;
    flex-shrink: 0;
  }

  .running {
    color: var(--green);
  }

  .stopped {
    color: var(--text-dim);
  }

  .dim {
    color: var(--text-dim);
  }
</style>
