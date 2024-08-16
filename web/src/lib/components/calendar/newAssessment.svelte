<script>
	import { onDestroy, onMount } from 'svelte';
	import TomSelect from 'tom-select';
  import 'tom-select/dist/css/tom-select.bootstrap5.min.css';
	import { Fetch } from '$lib/fetchUtil';
  import AvatarList from '$lib/components/calendar/Avatarlist.svelte';
	import { toast } from 'svelte-sonner';
	import { goto } from '$app/navigation';

  let usersOriginal = []
  let users = []
  let projectSelectElement;
  export let showModal = false

  let assessment = resetData()

  function resetData() {
    return {
      title: "",
      projects: [],
      dateFrom: null,
      dateTo: null,
      note: "",
      hackers: []
    }
  }

  $: if(showModal) { assessment = resetData() }

  let error

  async function postassessment() {
    const result = await Fetch("/api/planning/new", {method: "POST", body: JSON.stringify(assessment)})
    if(result.error) {
      error = result.error
      toast.error('Unable to save the plan');
    } else {
      toast.success('Plan has been created');
      assessment = resetData()
      showModal = false
      goto(`/planning/${result.id}/edit`)
    }
  }

	onMount(async () => {
		const projects = await Fetch('/api/project/all');
		const profiles = await Fetch('/api/profile/all');
		usersOriginal = profiles.users
    users = usersOriginal

		let tomSelect = new TomSelect(projectSelectElement, {
      plugins: ['remove_button'],
			valueField: 'ID',
			labelField: 'ProjectName',
			searchField: 'ProjectName',
			placeholder: "Choose a project...",
			options: projects,
			create: false,
      onChange: (values) => {
        // 'values' should be an array of selected IDs
        assessment.projects = values.map(value => {
          let project = projects.find(project => project.ID == value);
          return project ? { id: project.ID, name: project.ProjectName } : null;
        }).filter(id => id !== null);
      }
		});

		// Add options to TomSelect
		projects.forEach(project => {
			tomSelect.addOption({ value: project.ID, text: project.ProjectName });
		});

		// Refresh the TomSelect instance to show the new options
		// tomSelect.setValue(model?.ProjectID);
		tomSelect.refreshOptions(false);

    window.addEventListener('click', handleClickOutside);
  });

  function handleClickOutside(event) {
    const cardElement = document.getElementById('hackersDropdownList');
    if (cardElement && !cardElement.contains(event.target)) {
      showHackersList = false;
    }
  }

  onDestroy(() => {
    window.removeEventListener('click', handleClickOutside);
  });

  let showRemoveHacker = []
  let showHackersList = false

  function addHacker(user){
  // Add the user to the hackers list if not already included
  if (!assessment.hackers.includes(user)) {
    assessment.hackers.push(user);
  }

  // Filter the users list to exclude users that are in the hackers list
  users = users.filter(u => !assessment.hackers.includes(u));
  assessment.hackers = assessment.hackers
  }

function removeHacker(user) {
  // Remove the user from the hackers list
  assessment.hackers = assessment.hackers.filter(h => h !== user);

  // Optionally, add the user back to the users list if not already present
  if (!users.includes(user)) {
    users.push(user);
    users = users
  }
}

let filterText = ""

function filterUsers(event){
  filterText = event.target.value.toLowerCase()
  if (filterText == "") {
    users = usersOriginal?.filter(u => !assessment.hackers.includes(u));
  } else {
    users = users.filter(user =>
      user.name.toLowerCase().includes(filterText) // Convert user.name to lowercase as well
    );
  }
}

$: if(showHackersList){
  document.getElementById("filterQuery")?.focus()
} else {
  filterText = ""
  users = usersOriginal?.filter(u => !assessment.hackers.includes(u));
}
</script>

  <div class="card">
    <div class="card-body">
      {#if error}
        <div class="alert alert-warning" role="alert">
          {error}
        </div>
      {/if}
    <div class="mb-3">
			<!-- svelte-ignore a11y-autofocus -->
			<input
				type="text"
				class="form-control"
				name="example-text-input"
				placeholder="Title"
				autofocus
				bind:value={assessment.title}
			/>
		</div>

    <div class="row">
      <div class="">
        <div class="input-icon mb-3">
          <select multiple class="form-control" bind:this={projectSelectElement}></select>
          <span class="input-icon-addon">
            <!-- Download SVG icon from http://tabler-icons.io/i/search -->
            <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round" style="--darkreader-inline-stroke: currentColor;" data-darkreader-inline-stroke=""><path stroke="none" d="M0 0h24v24H0z" fill="none" style="--darkreader-inline-stroke: none;" data-darkreader-inline-stroke=""></path><path d="M10 10m-7 0a7 7 0 1 0 14 0a7 7 0 1 0 -14 0"></path><path d="M21 21l-6 -6"></path></svg>
          </span>
        </div>
      </div>
    </div>

		<div class="row">
			<div class="col-sm-5">
				<div class="mb-3">
					<input type="date" class="form-control" bind:value={assessment.dateFrom} />
          <input type="date" class="form-control" bind:value={assessment.dateTo} min={assessment.dateFrom}/>
				</div>
			</div>
		</div>

    <AvatarList hackers={assessment.hackers} on:updateHackers="{e => assessment.hackers = e.detail}"/>

<div class="row">
      <div class="mt-3">
			<!-- svelte-ignore a11y-autofocus -->
			<textarea
				class="form-control"
				placeholder="Notes..."
				bind:value={assessment.note}
			/>
		</div>
</div>

    </div>
    <div class="card-footer text-end">
      <a href="#" class="btn btn-primary" on:click="{postassessment}">Save</a>
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