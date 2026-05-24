<script lang="ts">
  import { UpdateNotes } from '../../wailsjs/go/main/App.js'
  import { pendingNotes } from '../stores/session'
  import { createEventDispatcher, tick, afterUpdate } from 'svelte'

  const dispatch = createEventDispatcher()

  let textarea: HTMLTextAreaElement
  let notes = ''
  let shouldFocus = false

  // When pendingNotes is set, schedule a focus after the DOM renders.
  $: if ($pendingNotes) {
    notes = ''
    shouldFocus = true
  }

  afterUpdate(() => {
    if (shouldFocus && textarea) {
      shouldFocus = false
      textarea.focus()
    }
  })

  async function save() {
    if (!$pendingNotes) return
    try {
      await UpdateNotes($pendingNotes, notes)
      dispatch('saved')
    } catch (e) {
      dispatch('error', String(e))
    }
    pendingNotes.set(null)
    notes = ''
  }

  function dismiss() {
    pendingNotes.set(null)
    notes = ''
  }

  function handleKey(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      e.preventDefault()
      dismiss()
    } else if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      save()
    }
  }
</script>

{#if $pendingNotes}
  <div class="notes-overlay">
    <span class="label">notes for last region</span>
    <textarea
      bind:this={textarea}
      bind:value={notes}
      on:keydown={handleKey}
      placeholder="enter notes… (enter to save, esc to skip)"
      rows="2"
    ></textarea>
    <div class="row">
      <button on:click={save}>save</button>
      <button on:click={dismiss} class="dim">skip</button>
    </div>
  </div>
{/if}

<style>
  .notes-overlay {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 8px 10px;
    background: var(--bg2);
    border: 1px solid var(--accent);
    border-left: 3px solid var(--accent);
  }

  .label {
    font-size: 11px;
    color: var(--text-dim);
    letter-spacing: 0.04em;
  }

  textarea {
    width: 100%;
    resize: vertical;
    font-family: inherit;
    font-size: 12px;
    background: var(--bg);
    color: var(--text);
    border: 1px solid var(--border);
    padding: 6px;
  }

  .row {
    display: flex;
    gap: 6px;
  }

  .dim {
    color: var(--text-dim);
  }
</style>
