<script>
    import { createEventDispatcher } from 'svelte';
    export let showModal;
    const dispatch = createEventDispatcher();
    let dialog;
    export let large = true;
    export let modalBlur = true;
    export let showHeader = true;

    $: if (dialog) {
        if (showModal) {
            dialog.showModal();
        } else {
            dialog.close();
        }
    }

    function handleClose() {
        dispatch('close');
    }
</script>
<dialog bind:this={dialog} on:close={handleClose}>
  <div class:modal-blur="{modalBlur}" class="modal fade show" id="modal-report" tabindex="-1" role="dialog" aria-modal="true" style="display: block;">
      <div class="modal-dialog modal-dialog-centered" class:modal-lg={large} role="document">
        <div class="modal-content">
          {#if showHeader}
          <div class="modal-header">
              <slot name="title">Default Title</slot>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close" on:click|self={() => dialog.close()}></button>
          </div>
          {/if}
            <slot></slot>
            <slot name="footer"></slot>
      </div>
    </div>
</dialog>