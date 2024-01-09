<script>
	import { Fetch } from "$lib/fetchUtil";
  import { onMount } from "svelte";
  import TomSelect from 'tom-select';
  import 'tom-select/dist/css/tom-select.bootstrap5.min.css';
  import { createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher();

  let selectElement;

  let users = [];

  onMount(async () => {
    users = await Fetch(`/api/user/all`);

    // We must check if users are not undefined or null
    if (users) {
      // Initialize TomSelect after users have been fetched
      let tomSelect = new TomSelect(selectElement, {
        plugins: ['remove_button'],
        persist: false,
        create: true, // Change to 'false' if you do not want users to add new entries
        onItemAdd: function() {
          // update your value binding here if necessary
          this.setTextboxValue('');
        },
        onItemRemove: function() {
          // update your value binding here if necessary
        }
      });

      // Add options to TomSelect
      users.forEach(user => {
        console.log(user)
        tomSelect.addOption({ value: user.Email, text: user.Name });
      });

      // Refresh the TomSelect instance to show the new options
      tomSelect.refreshOptions(false);
    }
  });

  function handleSelectChange(event) {
    const selectedValues = Array.from(event.target.selectedOptions).map(o => o.value);
    dispatch('selection', { selectedEmails: selectedValues });
  }
</script>

<select
  bind:this={selectElement}
  class="form-select"
  multiple
  id="select-states"
  on:change={handleSelectChange}>
</select>

<style>
  :global(.ts-dropdown-content) {
    background: var(--tblr-modal-bg);
    color: var(--tblr-body-color);
  }

  :global(.ts-wrapper.multi .ts-control > div) {
    background: var(--tblr-bg-surface-secondary);
    border: 1px solid var(--tblr-border-color);
    color: var(--tblr-body-color);
  }

  :global(.ts-wrapper.plugin-remove_button:not(.rtl) .item .remove) {
    border-left: 1px solid var(--tblr-border-color);
  }

</style>