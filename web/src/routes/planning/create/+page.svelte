<script>
	import { onMount } from 'svelte';
	import TomSelect from 'tom-select';
  import 'tom-select/dist/css/tom-select.bootstrap5.min.css';
	import { Fetch } from '$lib/fetchUtil';
	import Avatar from '$lib/components/Avatar.svelte';

  let users = []
  let projectSelectElement;
  let formData = {
    title: "",
    projects: "",
    dateFrom: null,
    dateTo: null
  }

	onMount(async () => {
		const projects = await Fetch('/api/project/all');
		users = await Fetch('/api/profile/all');

		let tomSelect = new TomSelect(projectSelectElement, {
      plugins: ['remove_button'],
			valueField: 'ID',
			labelField: 'ProjectName',
			searchField: 'ProjectName',
			placeholder: "Choose a project...",
			options: projects,
			create: false,
			onChange: (value) => {
				selectedProject = projects.find(project => project.ID == value);
			}
		});

		// Add options to TomSelect
		projects.forEach(project => {
			tomSelect.addOption({ value: project.ID, text: project.ProjectName });
		});

		// Refresh the TomSelect instance to show the new options
		tomSelect.setValue(model?.ProjectID);
		tomSelect.refreshOptions(false);
  });

  let hackers = []
  let showHackersList = true

</script>
<div class="col-4">
  <div class="card">
    <div class="card-header">
      <div class="card-title">New Assassment
      </div>
    </div>
    <div class="card-body">
    <div class="mb-3">
			<!-- svelte-ignore a11y-autofocus -->
			<input
				type="text"
				class="form-control"
				name="example-text-input"
				placeholder="Description"
				autofocus
				bind:value={formData.title}
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
					<input type="date" class="form-control" bind:value={formData.dateFrom} />
          <input type="date" class="form-control" bind:value={formData.dateTo} />
				</div>
			</div>
		</div>


      <div class="avatar-list" style="position:relative">
        {#each hackers as hacker}
          <span class="avatar rounded-circle" style="background-image: url({hacker.Picture})"></span>
        {/each}
          <!-- svelte-ignore a11y-click-events-have-key-events -->
          <!-- svelte-ignore a11y-no-static-element-interactions -->
          <span class="avatar rounded-circle cursor-pointer" on:click="{() => showHackersList = !showHackersList}"><i class="ti ti-plus"></i></span>
        <div class="card" style="position:absolute;margin-top: 42px;" hidden={showHackersList}>
          <div class="card-body">
            {#each users as user}
              <Avatar email={user.Email}/>
            {/each}
          </div>
        </div>
      </div>


<!--
    							<div class="col-xl-12">
								<div class="mb-3">
									<label for="clientEmail" class="form-label required">Whos in charge</label>
									<input
										type="email"
										class="form-control"
										name="clientEmail"
										bind:value={clientEmail}
										hidden
									/>
									<UserSearch on:selection={handleClientSearchChange} bind:selectedValues={clientEmail}/>
									<small class="form-hint">
										This refers to the individual or team accountable for overseeing this project.
										They will serve as the primary point of contact for all project-related matters.
									</small>
								</div>
								<div class="mb-3">
									<label for="hackername" class="form-label required">Executoners</label>
									<input
										type="email"
										class="form-control"
										name="hackername"
										bind:value={hackerName}
										hidden
									/>
									<UserSearch on:selection={handleUserSearchChange} bind:selectedValues={hackerName}/>
									<small class="form-hint">
										This denotes the individual(s) tasked with carrying out and following up on the
										testing phase of this project. Please provide a list of their email addresses,
										separated by commas. Note that while the system allows for searching existing
										users, it does not verify the validity of the provided email addresses.
									</small>
								</div>
							</div> -->

<div class="row">
      <div class="mt-3">
			<!-- svelte-ignore a11y-autofocus -->
			<textarea
				class="form-control"
				placeholder="Note"
				bind:value={formData.title}
			/>
		</div>
</div>

    </div>
    <div class="card-footer text-end">
      <a href="#" class="btn btn-primary">Save</a>
    </div>
  </div>
</div>

<style>
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

	:global(.ts-dropdown-content) {
    background: var(--tblr-card-bg);
    color: var(--tblr-body-color);
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