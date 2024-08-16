<script>
	import { Fetch } from '$lib/fetchUtil';
  import { onMount, createEventDispatcher } from 'svelte';
  import TomSelect from 'tom-select';
  import 'tom-select/dist/css/tom-select.bootstrap5.min.css';

  export let projects = [];
  const dispatch = createEventDispatcher();
  let projectSelectElement;
  let tomSelect;

onMount(async () => {
    const allProjects = await Fetch('/api/project/all');

    tomSelect = new TomSelect(projectSelectElement, {
        plugins: ['remove_button'],
        valueField: 'ID',
        labelField: 'ProjectName',
        searchField: 'ProjectName',
        placeholder: "Choose a project...",
        options: allProjects,
        create: false,
        onItemAdd: (value) => {
            updateProjects(value, allProjects);
        },
        onItemRemove: (value) => {
            updateProjects(value, allProjects, false);
        }
    });

    // Add options to TomSelect
    allProjects.forEach(project => {
        tomSelect.addOption({ value: project.ID, text: project.ProjectName });
    });

    // Set initially selected values
    const initialSelectedValues = projects.map(p => p.id);
    tomSelect.setValue(initialSelectedValues);

    tomSelect.refreshOptions(false);
});

$: if (tomSelect && projects) {
      const selectedValues = projects.map(p => p.id);
      // Get current values from TomSelect to compare
      const currentValues = tomSelect.getValue();

      // Convert both arrays to strings for easy comparison
      const selectedValuesStr = selectedValues.sort().join(",");
      const currentValuesStr = currentValues.sort().join(",");

      // Update TomSelect values only if there are changes
      if (selectedValuesStr !== currentValuesStr) {
          tomSelect.setValue(selectedValues);
      }
  }

function updateProjects(value, allProjects, add = true) {
    if (add) {
        // Check if the project is already in the list to avoid duplication
        if (!projects.some(p => p.id == value)) {
            let project = allProjects.find(project => project.ID == value);
            if (project) {
                // Add the new project if it's not already in the list
                projects = [...projects, { id: project.ID, name: project.ProjectName }];
            }
        }
    } else {
        // Remove the project from the list
        projects = projects.filter(p => p.id != value);
    }

    dispatch('updateProjects', projects);
}

</script>

<div class="row">
  <div class="">
    <div class="input-icon mb-3">
      <select multiple class="form-control" bind:this={projectSelectElement}></select>
      <span class="input-icon-addon">
        <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round" style="--darkreader-inline-stroke: currentColor;" data-darkreader-inline-stroke=""><path stroke="none" d="M0 0h24v24H0z" fill="none" style="--darkreader-inline-stroke: none;" data-darkreader-inline-stroke=""></path><path d="M10 10m-7 0a7 7 0 1 0 14 0a7 7 0 1 0 -14 0"></path><path d="M21 21l-6 -6"></path></svg>
      </span>
    </div>
  </div>
</div>

<style>
  .avatar-container {
    cursor: pointer;
    position: relative;
    display: inline-block; /* Or 'block' depending on your layout */
  }

  .overlay {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5); /* 50% black overlay */
    display: flex;
    justify-content: center; /* Center horizontally */
    align-items: center; /* Center vertically */
    color: white; /* Text color */
    font-size: 16px; /* Adjust as needed */
  }

  ul {
    list-style-type: none;
    padding: 0;
    margin: 0;
  }
  :global(.item) {
    background: var(--tblr-modal-bg);
    color: var(--tblr-body-color);
  }

  :global(input) {
    background: var(--tblr-modal-bg);
    color: var(--tblr-body-color);
  }

	:global(.ts-control) {
		height: 2.7em;
	}

  li.option:hover{
    cursor: pointer;
    background-color: rgba(var(--tblr-secondary-rgb),.08);
    color: inherit;
  }

  :global(.ts-wrapper.multi .ts-control > div) {
    background: var(--tblr-bg-surface-secondary);
    border: 1px solid var(--tblr-border-color);
    color: var(--tblr-body-color);
  }

	:global(.ts-control input) {
		color: var(--tblr-body-color);
	}

  :global(.ts-wrapper.plugin-remove_button:not(.rtl) .item .remove) {
    border-left: 1px solid var(--tblr-border-color);
  }
</style>