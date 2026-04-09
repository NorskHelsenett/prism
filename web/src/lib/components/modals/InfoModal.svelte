<script>
  export let showInfoModal = false;
  export let onOK;
  export let text = "This will alter something that cannot be reversed, but will not alter the app in a faulty state."
  export let buttonText = "Do it";

  function handleOK() {
    if (typeof onOK === 'function') {
      onOK();
      showInfoModal = false;
    }
  }
</script>

{#if showInfoModal}
<div class="modal modal-blur fade show" id="modal-small" tabindex="-1" role="dialog" style="display: block;" aria-modal="true">
  <div class="modal-dialog modal-sm modal-dialog-centered" role="document">
    <div class="modal-content">
      <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close" on:click={() => showInfoModal = false}></button>
      <div class="modal-status bg-info"></div>
      <div class="modal-body text-center py-4">
        <svg xmlns="http://www.w3.org/2000/svg" class="icon mb-2 text-info icon-lg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M12 9v4"></path><path d="M10.363 3.591l-8.106 13.534a1.914 1.914 0 0 0 1.636 2.871h16.214a1.914 1.914 0 0 0 1.636 -2.87l-8.106 -13.536a1.914 1.914 0 0 0 -3.274 0z"></path><path d="M12 16h.01"></path></svg>
        <div class="modal-title">Are you sure?</div>
        <slot>
          <div class="text-secondary">{text}</div>
        </slot>
      </div>
          <div class="modal-footer">
            <div class="w-100">
              <div class="row">
                <div class="col"><a href="#" class="btn w-100" data-bs-dismiss="modal" on:click={() => showInfoModal = false}>
                    Cancel
                  </a></div>
                <div class="col"><a href="#" class="btn btn-info w-100" data-bs-dismiss="modal" on:click={handleOK}>
                    {buttonText}
                  </a></div>
              </div>
            </div>
          </div>
    </div>
  </div>
</div>
{/if}