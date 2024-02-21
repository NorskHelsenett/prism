<script>
	import { Fetch } from "$lib/fetchUtil";
  import { onMount } from "svelte";
  import TomSelect from 'tom-select';
  import 'tom-select/dist/css/tom-select.bootstrap5.min.css';
  import { createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher();

  let selectElement;
  export let selectedValues;

  let users = [];
  let tomSelect

onMount(async () => {
  users = await Fetch(`/api/profile/all`);

  if (users) {
    tomSelect = new TomSelect(selectElement, {
      plugins: ['remove_button'],
      persist: false,
      createOnBlur: true,
      createFilter: /^[a-zA-Z0-9._-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,6}$/,
      create: function(input) {
        // Returnerer et nytt alternativobjekt for input som ikke eksisterer som alternativ
        return {
          value: input,
          text: input // Viser input som tekst for nye alternativer (f.eks. e-postadresser)
        };
      },
      onItemAdd: function() {
        this.setTextboxValue('');
      }
    });

    users.forEach(user => {
      tomSelect.addOption({ value: user.email, text: user.name });
    });

    // Splitt selectedValues og sjekk hver e-post
    let selectedEmails = selectedValues.split(",").filter(email => email.trim() !== ""); // Fjerner tomme strenger
    selectedEmails.forEach(email => {
      // Sjekk om e-posten allerede finnes som et alternativ, hvis ikke, legg til som nytt alternativ
      if (email && !tomSelect.options[email]) { // Sjekker at e-posten ikke er tom
        tomSelect.addOption({ value: email, text: email });
      }
    });

    // Sett de valgte verdiene, men ignorer tomme strenger
    if (selectedEmails.length > 0) {
      tomSelect.setValue(selectedEmails);
    }
    tomSelect.refreshOptions(false);
  }
});

$: if(selectedValues){
// Splitt selectedValues og sjekk hver e-post
    let selectedEmails = selectedValues.split(",").filter(email => email.trim() !== ""); // Fjerner tomme strenger
    selectedEmails.forEach(email => {
      // Sjekk om e-posten allerede finnes som et alternativ, hvis ikke, legg til som nytt alternativ
      if (email && !tomSelect.options[email]) { // Sjekker at e-posten ikke er tom
        tomSelect.addOption({ value: email, text: email });
      }
    });

    // Sett de valgte verdiene, men ignorer tomme strenger
    if (selectedEmails.length > 0) {
      tomSelect.setValue(selectedEmails);
    }
}

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
  :global(.item) {
    background: var(--tblr-modal-bg);
    color: var(--tblr-body-color);
  }

  :global(input) {
    background: var(--tblr-modal-bg);
    color: var(--tblr-body-color);
  }

  :global(.ts-dropdown-content) {
    background: var(--tblr-modal-bg);
    color: var(--tblr-body-color);
    background: var(--tblr-card-bg);
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