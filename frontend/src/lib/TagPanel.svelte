<script lang="ts">
  import { PressTag, AddBookmark } from '../../wailsjs/go/main/App.js'
  import { activePreset, openTags, pendingNotes } from '../stores/session'
  import { mpvRunning } from '../stores/playback'
  import { createEventDispatcher } from 'svelte'
  import type { region as regionNS } from '../../wailsjs/go/models'

  type Region = regionNS.Region

  const dispatch = createEventDispatcher()

  async function pressTag(key: string, label: string, color: string) {
    if (!$mpvRunning) return
    try {
      const r: Region | null = await PressTag(key, label, color)
      if (r) {
        dispatch('refresh')
        pendingNotes.set(r.id)
      } else {
        dispatch('refreshTags')
      }
    } catch (e) {
      dispatch('error', String(e))
    }
  }

  async function bookmark() {
    if (!$mpvRunning) return
    try {
      await AddBookmark()
      dispatch('refresh')
    } catch (e) {
      dispatch('error', String(e))
    }
  }

  function isOpen(key: string): boolean {
    return $openTags.includes(key)
  }

  function handleKey(e: KeyboardEvent) {
    if (!$mpvRunning || !$activePreset) return
    const target = e.target as HTMLElement
    if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA') return
    for (const tag of $activePreset.tags) {
      if (e.key === tag.key) {
        e.preventDefault()
        pressTag(tag.key, tag.label, tag.color)
        return
      }
    }
    if (e.key === '0') {
      e.preventDefault()
      bookmark()
    }
  }
</script>

<svelte:window on:keydown={handleKey} />

<div class="tag-panel">
  <div class="panel-header">
    <span class="panel-title">
      {$activePreset ? $activePreset.name : 'no preset'}
    </span>
  </div>

  {#if $activePreset}
    <div class="tag-grid">
      {#each $activePreset.tags as tag}
        <button
          class="tag-btn"
          class:open={isOpen(tag.key)}
          style="--tag-color: {tag.color};"
          on:click={() => pressTag(tag.key, tag.label, tag.color)}
          disabled={!$mpvRunning}
          title="{tag.key}: {tag.label}"
        >
          <span class="key">{tag.key}</span>
          <span class="label">{tag.label}</span>
          {#if isOpen(tag.key)}
            <span class="dot">●</span>
          {/if}
        </button>
      {/each}
    </div>
  {:else}
    <div class="empty">select a preset to start tagging</div>
  {/if}

  <div class="bookmark-row">
    <button on:click={bookmark} disabled={!$mpvRunning} title="add bookmark at current time (key: 0)">
      ◆ bookmark
    </button>
  </div>
</div>

<style>
  .tag-panel {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .panel-header {
    display: flex;
    align-items: center;
  }

  .panel-title {
    color: var(--text-dim);
    font-size: 11px;
    letter-spacing: 0.05em;
  }

  .tag-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .tag-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 10px;
    border: 1px solid var(--tag-color, var(--border));
    background: transparent;
    color: var(--text);
    cursor: pointer;
    transition: background 0.1s;
    min-width: 120px;
  }

  .tag-btn:hover:not(:disabled) {
    background: color-mix(in srgb, var(--tag-color, var(--accent)) 15%, transparent);
  }

  .tag-btn.open {
    background: color-mix(in srgb, var(--tag-color, var(--accent)) 20%, transparent);
    border-color: var(--tag-color, var(--accent));
  }

  .key {
    background: var(--bg);
    border: 1px solid var(--border);
    padding: 0 4px;
    font-size: 11px;
    color: var(--text-dim);
    border-radius: 2px;
    min-width: 16px;
    text-align: center;
  }

  .label {
    font-size: 12px;
    flex: 1;
  }

  .dot {
    color: var(--tag-color, var(--accent));
    font-size: 10px;
  }

  .bookmark-row {
    margin-top: 4px;
  }

  .empty {
    color: var(--text-dim);
    font-size: 12px;
  }
</style>
