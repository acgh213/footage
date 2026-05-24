<script lang="ts">
  import { BrowseForFile, OpenFile, GetTimePos, StopPlayer, GetPresets, GetMPVPath, BrowseForMPV } from '../wailsjs/go/main/App.js'
  import type { preset } from '../wailsjs/go/models'
  type Preset = preset.Preset

  let filePath = ''
  let timePos: number | null = null
  let status = ''
  let statusOk = true
  let running = false
  let presets: Preset[] = []
  let mpvPath = ''

  async function refreshMPVPath() {
    try { mpvPath = await GetMPVPath() } catch (_) {}
  }

  async function browse() {
    try {
      const p = await BrowseForFile()
      if (p) filePath = p
    } catch (e) {
      setStatus(String(e), false)
    }
  }

  async function openFile() {
    if (!filePath) return
    status = 'opening mpv…'
    statusOk = true
    try {
      await OpenFile(filePath)
      running = true
      setStatus('mpv running', true)
      await loadPresets()
    } catch (e) {
      running = false
      setStatus(String(e), false)
    }
  }

  async function getTime() {
    try {
      timePos = await GetTimePos()
      setStatus('', true)
    } catch (e) {
      setStatus(String(e), false)
    }
  }

  async function stop() {
    try {
      await StopPlayer()
      running = false
      timePos = null
      setStatus('mpv stopped', true)
    } catch (e) {
      setStatus(String(e), false)
    }
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

  async function loadPresets() {
    try {
      presets = await GetPresets()
    } catch (_) {}
  }

  function setStatus(msg: string, ok: boolean) {
    status = msg
    statusOk = ok
  }

  function formatTime(s: number): string {
    const h = Math.floor(s / 3600)
    const m = Math.floor((s % 3600) / 60)
    const sec = (s % 60).toFixed(3)
    return `${String(h).padStart(2,'0')}:${String(m).padStart(2,'0')}:${sec.padStart(6,'0')}`
  }

  // Load presets and mpv path on mount
  loadPresets()
  refreshMPVPath()
</script>

<main>
  <header>
    <span class="title">footage</span>
    <span class="subtitle">phase 0 — smoke test</span>
  </header>

  <section class="panel">
    <div class="row">
      <label>mpv</label>
      <span class="path-display" class:not-found={!mpvPath}>
        {mpvPath || 'not found'}
      </span>
      <button on:click={browseMPV} title="manually locate mpv.exe">locate…</button>
    </div>

    <div class="row">
      <label>video file</label>
      <input type="text" bind:value={filePath} placeholder="path to video…" class="path-input" readonly />
      <button on:click={browse}>browse</button>
    </div>

    <div class="row">
      <label>&nbsp;</label>
      <button on:click={openFile} disabled={!filePath || running || !mpvPath}>open in mpv</button>
      <button on:click={stop} disabled={!running}>stop mpv</button>
    </div>

    <div class="row">
      <label>playback</label>
      <button on:click={getTime} disabled={!running}>get time</button>
      {#if timePos !== null}
        <span class="time-display">{formatTime(timePos)}</span>
      {/if}
    </div>
  </section>

  {#if status}
    <div class="status" class:ok={statusOk} class:err={!statusOk}>
      {status}
    </div>
  {/if}

  {#if presets.length > 0}
    <section class="panel presets">
      <div class="panel-title">presets ({presets.length})</div>
      {#each presets as preset}
        <div class="preset-row">
          <span class="preset-name">{preset.name}</span>
          <span class="preset-tags">
            {#each preset.tags as tag}
              <span class="tag" style="background: {tag.color}22; border: 1px solid {tag.color}44; color: {tag.color};">
                {tag.key}  {tag.label}
              </span>
            {/each}
          </span>
        </div>
      {/each}
    </section>
  {/if}

  <footer>
    <span class:running={running} class:stopped={!running}>
      mpv: {running ? 'running' : 'not running'}
    </span>
  </footer>
</main>

<style>
  main {
    display: flex;
    flex-direction: column;
    height: 100vh;
    padding: 16px;
    gap: 12px;
  }

  header {
    display: flex;
    align-items: baseline;
    gap: 12px;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--border);
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

  .panel {
    background: var(--bg2);
    border: 1px solid var(--border);
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .panel-title {
    color: var(--text-dim);
    font-size: 11px;
    margin-bottom: 4px;
    letter-spacing: 0.05em;
  }

  .row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  label {
    width: 80px;
    color: var(--text-dim);
    flex-shrink: 0;
  }

  .path-input {
    flex: 1;
  }

  .path-display {
    flex: 1;
    color: var(--green);
    font-size: 11px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .path-display.not-found {
    color: var(--red);
  }

  .time-display {
    color: var(--accent);
    font-size: 15px;
    letter-spacing: 0.05em;
    padding: 0 8px;
  }

  .status {
    padding: 6px 10px;
    border-left: 3px solid;
    font-size: 12px;
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

  .presets {
    flex: 1;
    overflow-y: auto;
  }

  .preset-row {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 4px 0;
    border-bottom: 1px solid var(--border);
  }

  .preset-row:last-child {
    border-bottom: none;
  }

  .preset-name {
    width: 120px;
    flex-shrink: 0;
    color: var(--accent);
  }

  .preset-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }

  footer {
    padding-top: 8px;
    border-top: 1px solid var(--border);
    font-size: 11px;
  }

  .running {
    color: var(--green);
  }

  .stopped {
    color: var(--text-dim);
  }
</style>
