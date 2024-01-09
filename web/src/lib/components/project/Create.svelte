<script>
	import { notification } from '$lib/stores/notificationStore';
	import { Fetch } from '$lib/fetchUtil.js';
	import Modal from '../Modal.svelte';
	import UserSearch from '../UserSearch.svelte';
	export let showModal = false;

	// Declare local state for form inputs
	let projectName = '';
	let slackChannel = '';
	let description = '';
	let clientEmail = '';
	let hackerName = '';
	let file; // For file input
	let isBugBounty = false

    function closeModal() {
        showModal = false;
    }

    function openFileUploadDialog() {
        document.getElementById("fileInput").click();
    }

	function handleUserSearchChange(event) {
    hackerName = event.detail.selectedEmails.join(',');
  }

	function handleClientSearchChange(event) {
    clientEmail = event.detail.selectedEmails.join(',');
  }

let errorMessage = '';

    async function validateImage(file) {
        return new Promise((resolve, reject) => {
            const img = new Image();
            img.onload = () => {
                const { width, height } = img;
                if (width !== height) {
                    reject("Image must be square.");
                } else if (width < 512 || height < 512 || width > 1024 || height > 1024) {
                    reject("Image dimensions must be between 512x512 and 1024x1024 pixels.");
                } else {
                    resolve();
                }
            };
            img.onerror = () => {
                reject("Invalid image file.");
            };
            img.src = URL.createObjectURL(file);
        });
    }

    async function handleSubmit() {
        // Reset error message
        errorMessage = '';

        // Validate file
        if (file && file.length > 0) {
            const selectedFile = file[0];
            if (selectedFile.type !== 'image/jpeg' && selectedFile.type !== 'image/png') {
                errorMessage = 'File must be a JPEG or PNG image.';
                return;
            }

            try {
                await validateImage(selectedFile);
            } catch (error) {
                errorMessage = error;
								console.log(errorMessage)
                return;
            }
        }

			let projectData = {
        projectName: projectName,
        slackChannel: slackChannel,
        description: description,
        clientEmail: clientEmail,
        hackerName: hackerName,
				isBugBounty: isBugBounty,
        fileData: '' // For file data in Base64 (optional)
    };
			const formData = new FormData();
			formData.append('projectName', projectName);
			formData.append('slackChannel', slackChannel);
			formData.append('description', description);
			formData.append('clientEmail', clientEmail);
			formData.append('hackerName', hackerName);

			// Add file to FormData if a file is selected
			if (file) {
					formData.append('file', file[0]);
			}
try {
			const response = await Fetch(`/api/project`, {
				method: 'POST',
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
		}
    }
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
									<div class="col-md-8 col-xl-8">
										<div class="mb-3">
											<label for="projectname" class="form-label required">Project Name</label>
											<input type="text" class="form-control" name="projectname" bind:value={projectName}/>
										</div>

										<div class="mb-3">
											<div class="form-label">Bug Bounty</div>
												<label class="form-check form-switch cursor-pointer">
														<input class="form-check-input" type="checkbox" bind:checked={isBugBounty}>
														<span class="form-check-label">This project is part of the bug bounty program</span>
													</label>
											</div>

									<div class="mb-3">
									<label class="form-label">Slack Channel</label>
									<div class="input-group mb-3">
										<span class="input-group-text"> # </span>
										<input type="text" class="form-control" placeholder="Search…" bind:value={slackChannel}/>
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
												></path><path d="M10 10m-7 0a7 7 0 1 0 14 0a7 7 0 1 0 -14 0"></path><path
													d="M21 21l-6 -6"
												></path></svg
											>
										</span>
									</div>
									</div>
									</div>

									<div class="col-md-4 col-xl-4">
										<div class="mb-3 h-100">
											<div
												class="dropzone cursor-pointer"
												on:click={openFileUploadDialog}
												tabindex="0">

												<!-- Use a label element as the clickable button -->
												<label for="fileInput" class="file-button cursor-pointer text-secondary">
													Add icon
												</label>

												<!-- Input type file (hidden) for file selection -->
												<input type="file" id="fileInput" accept="image/*" bind:files={file}/>
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
										></textarea>
								</div>
							</div>
						</div>
						<div class="row g-3">
							<div class="col-xl-12">
									<div class="mb-3">
										<label for="clientEmail" class="form-label required">Client emails</label>
										<input type="email" class="form-control" name="clientEmail" bind:value={clientEmail} hidden/>
										<UserSearch on:selection={handleClientSearchChange}/>
									</div>
									<div class="mb-3">
										<label for="hackername" class="form-label required">Responsible hackers</label>
										<input type="email" class="form-control" name="hackername" bind:value={hackerName} hidden/>
										<UserSearch on:selection={handleUserSearchChange}/>
										<small class="form-hint">
											Comma separated list of emails. There will be no verification that an email is valid, but you can search for an existing user.
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
		<a href="#" class="btn btn-link link-secondary" on:click="{closeModal}">
			Cancel
		</a>
		<a href="#" class="btn btn-primary ms-auto" on:click|preventDefault={handleSubmit}>
			<!-- Download SVG icon from http://tabler-icons.io/i/plus -->
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
				><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M12 5l0 14"></path><path
					d="M5 12l14 0"
				></path></svg
			>
			Create Project
		</a>
	</div>
</Modal>

<style>
	  .dropzone {
    border: 2px dashed #1B58A0;
    border-width: 1px;
    border-style: dashed;
    padding: 20px;
    display: flex; /* Use Flexbox to center the label */
    flex-direction: column;
    align-items: center; /* Center horizontally */
    justify-content: center; /* Center vertically */
    cursor: pointer;
    transition: box-shadow 0.3s ease;
		height: 100%;
  }

    .dropzone.glow {
        box-shadow: 0 0 15px 5px rgba(0, 0, 255, 0.5);
    }

    .dropzone.glow:hover {
        box-shadow: 0 0 25px 10px rgba(0, 0, 255, 0.7);
    }

    .dropzone input[type="file"] {
        display: none;
    }
</style>