<script lang="ts">
  import { createEventDispatcher } from 'svelte'

  export let message = 'delete this entry?'

  const dispatch = createEventDispatcher()
</script>

<div class="confirm-overlay" role="dialog" aria-modal="true">
  <div class="confirm-box">
    <span class="msg">{message}</span>
    <div class="btns">
      <button class="del-btn" on:click={() => dispatch('confirm')}>delete</button>
      <button on:click={() => dispatch('cancel')}>cancel</button>
    </div>
  </div>
</div>

<style>
  .confirm-overlay {
    position: fixed;
    inset: 0;
    background: #00000066;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }

  .confirm-box {
    background: var(--bg2);
    border: 1px solid var(--border);
    padding: 16px 20px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-width: 220px;
  }

  .msg {
    font-size: 13px;
    color: var(--text);
  }

  .btns {
    display: flex;
    gap: 8px;
    justify-content: flex-end;
  }

  .del-btn {
    color: var(--red);
    border-color: var(--red);
  }

  .del-btn:hover {
    background: #e8919e22;
  }
</style>
