<script>
  export let showDeleteModal = false;
  export let onDelete;
  export let text = "If you proceed, vulnerability and attachments will be deleted. What you've done cannot be undone."
  export let deleteButtonText = "Yes, delete it";

  function handleDelete() {
    if (typeof onDelete === 'function') {
      onDelete();
      showDeleteModal = false;
    }
  }
</script>

{#if showDeleteModal}
<div class="modal modal-blur fade show" id="modal-small" tabindex="-1" role="dialog" style="display: block;" aria-modal="true">
  <div class="modal-dialog modal-sm modal-dialog-centered" role="document">
    <div class="modal-content">
      <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close" on:click={() => showDeleteModal = false}></button>
      <div class="modal-status bg-danger"></div>
      <div class="modal-body text-center py-4">
        <svg xmlns="http://www.w3.org/2000/svg" class="icon mb-2 text-danger icon-lg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M12 9v4"></path><path d="M10.363 3.591l-8.106 13.534a1.914 1.914 0 0 0 1.636 2.871h16.214a1.914 1.914 0 0 0 1.636 -2.87l-8.106 -13.536a1.914 1.914 0 0 0 -3.274 0z"></path><path d="M12 16h.01"></path></svg>
        <div class="modal-title">Are you sure?</div>
        <div class="text-secondary">{text}</div>
      </div>
          <div class="modal-footer">
            <div class="w-100">
              <div class="row">
                <div class="col"><a href="#" class="btn w-100" data-bs-dismiss="modal" on:click={() => showDeleteModal = false}>
                    Cancel
                  </a></div>
                <div class="col"><a href="#" class="btn btn-danger w-100" data-bs-dismiss="modal" on:click={handleDelete}>
                    {deleteButtonText}
                  </a></div>
              </div>
            </div>
          </div>
    </div>
  </div>
</div>
{/if}