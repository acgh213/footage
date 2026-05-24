<script lang="ts">
  import { Seek, SetSpeed, Pause, Play } from '../../wailsjs/go/main/App.js'
  import { timePos, paused, speed, mpvRunning } from '../stores/playback'

  function fmt(s: number): string {
    const h = Math.floor(s / 3600)
    const m = Math.floor((s % 3600) / 60)
    const sec = (s % 60).toFixed(3)
    return `${String(h).padStart(2,'0')}:${String(m).padStart(2,'0')}:${sec.padStart(6,'0')}`
  }

  async function seek(delta: number) {
    try { await Seek(delta, true) } catch {}
  }

  async function togglePause() {
    try {
      if ($paused) {
        await Play()
        paused.set(false)
      } else {
        await Pause()
        paused.set(true)
      }
    } catch {}
  }

  async function setSpeed(s: number) {
    try {
      await SetSpeed(s)
      speed.set(s)
    } catch {}
  }
</script>

<div class="transport" class:disabled={!$mpvRunning}>
  <span class="time">{fmt($timePos)}</span>

  <div class="controls">
    <button on:click={() => seek(-30)} disabled={!$mpvRunning} title="-30s">«</button>
    <button on:click={() => seek(-5)} disabled={!$mpvRunning} title="-5s">‹</button>
    <button on:click={togglePause} disabled={!$mpvRunning} class="play-btn">
      {$paused ? '▶' : '⏸'}
    </button>
    <button on:click={() => seek(5)} disabled={!$mpvRunning} title="+5s">›</button>
    <button on:click={() => seek(30)} disabled={!$mpvRunning} title="+30s">»</button>
  </div>

  <div class="speeds">
    {#each [1, 1.5, 2, 4] as s}
      <button
        on:click={() => setSpeed(s)}
        disabled={!$mpvRunning}
        class:active={$speed === s}
      >{s}×</button>
    {/each}
  </div>
</div>

<style>
  .transport {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .time {
    font-size: 14px;
    color: var(--accent);
    letter-spacing: 0.04em;
    min-width: 100px;
  }

  .controls {
    display: flex;
    gap: 4px;
  }

  .play-btn {
    min-width: 32px;
  }

  .speeds {
    display: flex;
    gap: 4px;
  }

  .speeds button.active {
    background: var(--accent);
    color: var(--bg);
  }

  .disabled .time {
    color: var(--text-dim);
  }
</style>
