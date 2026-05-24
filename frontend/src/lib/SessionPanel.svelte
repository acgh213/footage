<script lang="ts">
  import { BrowseForFile, AddFileToSession, RemoveFileFromSession, SetActiveFile } from '../../wailsjs/go/main/App.js'
  import { currentSession, activeFile } from '../stores/session'
  import { createEventDispatcher } from 'svelte'

  const dispatch = createEventDispatcher()

  async function browseAdd() {
    try {
      const path = await BrowseForFile()
      if (!path) return
      const s = await AddFileToSession(path)
      currentSession.set(s)
    } catch (e) {
      dispatch('error', String(e))
    }
  }

  async function remove(path: string) {
    try {
      const s = await RemoveFileFromSession(path)
      currentSession.set(s)
    } catch (e) {
      dispatch('error', String(e))
    }
  }

  async function activate(idx: number) {
    try {
      await SetActiveFile(idx)
      if ($currentSession) {
        currentSession.set({ ...$currentSession, active_idx: idx })
      }
      dispatch('opened')
    } catch (e) {
      dispatch('error', String(e))
    }
  }

  function basename(p: string): string {
    return p.replace(/.*[\\/]/, '')
  }
</script>

<div class="session-panel">
  <div class="panel-header">
    <span class="panel-title">session files</span>
    <button on:click={browseAdd} class="add-btn" title="add video to session">+</button>
  </div>

  {#if $currentSession && $currentSession.files && $currentSession.files.length > 0}
    <div class="file-list">
      {#each $currentSession.files as file, i}
        <div
          class="file-row"
          class:active={file === $activeFile}
          role="button"
          tabindex="0"
          on:click={() => activate(i)}
          on:keydown={e => e.key === 'Enter' && activate(i)}
        >
          <span class="file-name" title={file}>{basename(file)}</span>
          <button
            class="remove-btn"
            on:click|stopPropagation={() => remove(file)}
            title="remove from session"
          >×</button>
        </div>
      {/each}
    </div>
  {:else}
    <div class="empty">no files — click + to add</div>
  {/if}
</div>

<style>
  .session-panel {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .panel-title {
    color: var(--text-dim);
    font-size: 11px;
    letter-spacing: 0.05em;
  }

  .add-btn {
    padding: 0 6px;
    font-size: 16px;
    line-height: 1;
  }

  .file-list {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .file-row {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 4px 6px;
    cursor: pointer;
    border: 1px solid transparent;
  }

  .file-row:hover {
    background: var(--bg2);
    border-color: var(--border);
  }

  .file-row.active {
    border-color: var(--accent);
    background: var(--bg2);
  }

  .file-name {
    flex: 1;
    font-size: 12px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .file-row.active .file-name {
    color: var(--accent);
  }

  .remove-btn {
    padding: 0 4px;
    opacity: 0.4;
    font-size: 14px;
  }

  .remove-btn:hover {
    opacity: 1;
    color: var(--red);
  }

  .empty {
    color: var(--text-dim);
    font-size: 12px;
    padding: 4px 0;
  }
</style>
