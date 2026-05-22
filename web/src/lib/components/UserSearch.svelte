<script>
  import { run } from 'svelte/legacy';

  import { Fetch } from "$lib/fetchUtil";
  import { onMount } from "svelte";
  import TomSelect from 'tom-select';
  import 'tom-select/dist/css/tom-select.bootstrap5.min.css';
  import { createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher();

  let selectElement = $state();
  let { selectedValues = [] } = $props();

  let users = {
    teams: [],
    users: []
  };

  let tomSelect = $state();

  onMount(async () => {
    users = await Fetch(`/api/profile/all`);

    if (users.teams.length > 0) {
      tomSelect = new TomSelect(selectElement, {
        plugins: ['remove_button'],
        persist: false,
        createOnBlur: true,
        createFilter: /^[a-zA-Z0-9._-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,6}$/,
        create: function(input) {
          return {
            value: input,
            text: input
          };
        },
        onItemAdd: function() {
          this.setTextboxValue('');
        }
      });

      users.users.forEach(user => {
      tomSelect.addOption({ value: user.email, text: user.name });
    });

      tomSelect.refreshOptions(false);
    }
  });

  run(() => {
    if(tomSelect && Array.isArray(selectedValues)){
      let selectedEmails = selectedValues.filter(email => email.trim() !== "");
      selectedEmails.forEach(email => {
        if (email && !tomSelect.options[email]) {
          tomSelect.addOption({ value: email, text: email });
        }
      });

      if (selectedEmails.length > 0) {
        tomSelect.setValue(selectedEmails);
      }
    }
  });

  function handleSelectChange(event) {
    const selectedValues = tomSelect.getValue();
    dispatch('selection', { selectedEmails: selectedValues });
  }
</script>

<select
  bind:this={selectElement}
  class="form-select"
  multiple
  id="select-states"
  onchange={handleSelectChange}>
</select>

<style>
  :global(.item) {
    background: var(--tblr-modal-bg);
    color: var(--tblr-body-color);
  }

  :global(input) {
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

  :global(.ts-wrapper .ts-control) {
    flex-wrap: wrap;
    min-height: auto;
    height: auto;
  }

  :global(.ts-control){
    padding-top: 0px !important;
    padding-bottom: 0px !important;
    height: 2.7em;
    color: var(--tblr-body-color);
  }

  :global(.ts-wrapper){
    height: 3em;
    margin-bottom: 0px !important;
  }

  :global(.ts-dropdown) {
    background: var(--tblr-card-bg, var(--tblr-modal-bg, default-bg-color));
    color: var(--tblr-body-color, default-body-color);
  }
</style>