<script>
  import { run, self } from 'svelte/legacy';

    import { createEventDispatcher } from 'svelte';
    const dispatch = createEventDispatcher();
    let dialog = $state();
  /**
   * @typedef {Object} Props
   * @property {any} showModal
   * @property {boolean} [large]
   * @property {boolean} [modalBlur]
   * @property {boolean} [showHeader]
   * @property {import('svelte').Snippet} [title]
   * @property {import('svelte').Snippet} [children]
   * @property {import('svelte').Snippet} [footer]
   */

  /** @type {Props} */
  let {
    showModal,
    large = true,
    modalBlur = true,
    showHeader = true,
    title,
    children,
    footer
  } = $props();

    run(() => {
    if (dialog) {
          if (showModal) {
              dialog.showModal();
          } else {
              dialog.close();
          }
      }
  });

    function handleClose() {
        dispatch('close');
    }
</script>
<dialog bind:this={dialog} onclose={handleClose}>
  <div class:modal-blur="{modalBlur}" class="modal fade show" id="modal-report" tabindex="-1" role="dialog" aria-modal="true" style="display: block;">
      <div class="modal-dialog modal-dialog-centered" class:modal-lg={large} role="document">
        <div class="modal-content">
          {#if showHeader}
          <div class="modal-header">
              {#if title}{@render title()}{:else}Default Title{/if}
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close" onclick={self(() => dialog.close())}></button>
          </div>
          {/if}
            {@render children?.()}
            {@render footer?.()}
      </div>
    </div>
</dialog>