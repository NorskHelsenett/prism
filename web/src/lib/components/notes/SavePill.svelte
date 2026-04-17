<script>
  import { saveState } from '$lib/stores/notesStore';

  $: state = $saveState;
</script>

{#if state !== 'idle'}
  <div class="save-pill" class:error={state === 'error'} title={state}>
    {#if state === 'saving'}
      <svg class="spinner" viewBox="0 0 24 24" width="14" height="14" aria-hidden="true">
        <circle cx="12" cy="12" r="9" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-dasharray="14 36" />
      </svg>
    {:else if state === 'saved'}
      <svg viewBox="0 0 24 24" width="14" height="14" aria-hidden="true">
        <path d="M5 12l5 5L20 7" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
    {:else}
      <svg viewBox="0 0 24 24" width="14" height="14" aria-hidden="true">
        <circle cx="12" cy="12" r="9" fill="none" stroke="currentColor" stroke-width="2" />
        <path d="M12 7v6" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
        <circle cx="12" cy="16" r="1" fill="currentColor" />
      </svg>
    {/if}
  </div>
{/if}

<style>
  .save-pill {
    position: fixed;
    bottom: 16px;
    right: 16px;
    z-index: 500;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 26px;
    height: 26px;
    border-radius: 999px;
    background: var(--tblr-bg-surface, #fff);
    color: var(--tblr-success, #2fb344);
    border: 1px solid var(--tblr-border-color, #e6e7e9);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
    animation: pill-in 160ms ease-out;
    pointer-events: none;
  }
  .save-pill.error {
    color: var(--tblr-danger, #d63939);
  }
  .spinner {
    animation: spin 900ms linear infinite;
    color: var(--tblr-secondary, #6c757d);
  }
  @keyframes spin {
    to { transform: rotate(360deg); }
  }
  @keyframes pill-in {
    from { opacity: 0; transform: translateY(4px); }
    to { opacity: 1; transform: translateY(0); }
  }
</style>
