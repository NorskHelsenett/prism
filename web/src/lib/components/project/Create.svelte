<script>
  import { onMount } from "svelte";
	import { notification } from '$lib/stores/notificationStore';
	import { Fetch } from '$lib/fetchUtil.js';
	import Modal from '../Modal.svelte';
	import UserSearch from '../UserSearch.svelte';
	import { textAreaListHelper } from "$lib/textAreaListHelper";
	export let showModal = false;
	export let model;

	// Declare local state for form inputs
	let projectName = '';
	let slackChannel = '';
	let description = '';
	let clientEmail = '';
	let hackerName = '';
	let isBugBounty = false;
	let isPersisting = false

	function closeModal() {
		showModal = false;
	}

	function handleUserSearchChange(event) {
		hackerName = event.detail.selectedEmails.join(',');
	}

	function handleClientSearchChange(event) {
		clientEmail = event.detail.selectedEmails.join(',');
	}

	let errorMessage = '';

	async function handleSubmit() {
		// Reset error message
		errorMessage = '';
		isPersisting = true

		let projectData = {
			projectName: projectName,
			slackChannel: slackChannel,
			description: description,
			clientEmail: clientEmail,
			hackerName: hackerName,
			isBugBounty: isBugBounty,
			ID: model?.ID
		};
		const formData = new FormData();
		formData.append('projectName', projectName);
		formData.append('slackChannel', slackChannel);
		formData.append('description', description);
		formData.append('clientEmail', clientEmail);
		formData.append('hackerName', hackerName);
		try {
			const method = (model == null) ? "POST" : "PUT";
			const url = (model == null) ? "/api/project" : `/api/project/${model.ID}`
			const response = await Fetch(url, {
				method: method,
				headers: {
					'Content-Type': 'application/json'
				},
				credentials: 'include',
				body: `${JSON.stringify(projectData)}`
			});

			if (!response) {
				throw new Error('Network response was not ok');
			}

			notification.addAlert({
				type: 'success',
				title: 'Project added',
				message: 'Project succesfully stored'
			});

			closeModal();

			// Handle response here
		} catch (error) {
			notification.addAlert({
				type: 'error',
				title: 'Error',
				message: 'Unable to store new vulnerability!'
			});
			console.error('There was a problem storing the new project:', error);
			isPersisting = false
		}
	}
  onMount(async () => {
		if (model != null) {
			projectName = model.ProjectName;
			slackChannel = model.SlackChannel;
			description = model.Description;
			clientEmail = model.ClientEmail;
			hackerName = model.HackerName;
			isBugBounty = model.IsBugBounty;
		}
	});
</script>

<Modal bind:showModal on:close={closeModal}>
	<h5 class="modal-title" slot="title">Create new project</h5>
	<div class="modal-body">
		<div class="row row-cards">
			<div class="col-12">
				<form action="" method="POST" class="card">
					<div class="card-body">
						<div class="row g-3">
							<div class="col-xl-12">
								<div class="row">
									<div class="col-md-12 col-xl-12">
										<div class="mb-3">
											<label for="projectname" class="form-label required">Project Name</label>
											<input
												type="text"
												class="form-control"
												name="projectname"
												bind:value={projectName}
											/>
										</div>
										<div class="mb-3">
											<div class="form-label">Bug Bounty</div>
											<label class="form-check form-switch cursor-pointer">
												<input
													class="form-check-input"
													type="checkbox"
													bind:checked={isBugBounty}
												/>
												<span class="form-check-label"
													>This project is part of the bug bounty program</span
												>
											</label>
										</div>
										<div class="mb-3">
											<label class="form-label">Slack Channel</label>
											<div class="input-group mb-3">
												<span class="input-group-text"> # </span>
												<input
													type="text"
													class="form-control"
													placeholder="Search…"
													bind:value={slackChannel}
												/>
												<span class="input-icon-addon">
													<!-- Download SVG icon from http://tabler-icons.io/i/search -->
													<svg
														xmlns="http://www.w3.org/2000/svg"
														class="icon"
														width="24"
														height="24"
														viewBox="0 0 24 24"
														stroke-width="2"
														stroke="currentColor"
														fill="none"
														stroke-linecap="round"
														stroke-linejoin="round"
														style="--darkreader-inline-stroke: currentColor;"
														data-darkreader-inline-stroke=""
														><path
															stroke="none"
															d="M0 0h24v24H0z"
															fill="none"
															style="--darkreader-inline-stroke: none;"
															data-darkreader-inline-stroke=""
														></path><path d="M10 10m-7 0a7 7 0 1 0 14 0a7 7 0 1 0 -14 0"
														></path><path d="M21 21l-6 -6"></path></svg
													>
												</span>
											</div>
										</div>
									</div>
								</div>
							</div>
						</div>
						<div class="row g-3">
							<div class="col-xl-12">
								<div class="mb-3">
									<label class="form-label"
										>Description <span class="form-label-description">56/100</span></label
									>
									<textarea
										class="form-control"
										name="example-textarea-input"
										rows="6"
										bind:value={description}
										placeholder="Content.."
                    use:textAreaListHelper
									></textarea>
								</div>
							</div>
						</div>
						<div class="row g-3">
							<div class="col-xl-12">
								<div class="mb-3">
									<label for="clientEmail" class="form-label required">Client emails</label>
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
									<label for="hackername" class="form-label required">Responsible hackers</label>
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
							</div>
						</div>
					</div>
				</form>
			</div>
		</div>
	</div>

	<div class="modal-footer" slot="footer">
		<a href="#" class="btn btn-link link-secondary" on:click={closeModal}> Cancel </a>
		<button disabled={isPersisting} href="#" class="btn btn-primary ms-auto" on:click|preventDefault={handleSubmit}>
			<!-- Download SVG icon from http://tabler-icons.io/i/plus -->
			{#if model == null}
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="icon"
					width="24"
					height="24"
					viewBox="0 0 24 24"
					stroke-width="2"
					stroke="currentColor"
					fill="none"
					stroke-linecap="round"
					stroke-linejoin="round"
					><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M12 5l0 14"
					></path><path d="M5 12l14 0"></path></svg
				>
				Create Project
			{:else}
				Update Project
			{/if}
		</button>
	</div>
</Modal>
