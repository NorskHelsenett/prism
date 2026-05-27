<script>
  import { preventDefault } from 'svelte/legacy';

  import { onMount } from "svelte";
  import { toast } from 'svelte-sonner';
  import { Fetch } from '$lib/fetchUtil.js';
  import Modal from '../Modal.svelte';
  import UserSearch from '../UserSearch.svelte';
  import RichTextEditor from '../editor/RichTextEditor.svelte';
  import Avatarlist from "../calendar/Avatarlist.svelte";

  /**
   * @typedef {Object} Props
   * @property {boolean} [showModal]
   * @property {any} model
   */

  /** @type {Props} */
  let { showModal = $bindable(false), model } = $props();

  // Declare local state for form inputs
  let projectName = $state('');
  let slackChannel = $state('');
  let description = $state('');
  let clientEmailInitilized = $state([]);
  let clientEmail = [];
  let hackerNameInitilized = $state([]);
  let hackerName = [];
  let isBugBounty = $state(false);
  let isPersisting = $state(false)
  let hackers = $state([]);
  let startDate = $state('');
  let endDate = $state('');
  let color = $state('#206bc4');

  const projectColors = [
    '#206bc4', '#4299e1', '#2fb344', '#f76707', '#d63939',
    '#ae3ec9', '#fbbf24', '#64748b', '#0ca678', '#6366f1'
  ];

  function closeModal() {
    showModal = false;
  }

  function handleUserSearchChange(event) {
    hackerName = event.detail;
    hackers = hackerName.map(item => ({ email: item.email }));
  }

  function handleClientSearchChange(event) {
    clientEmail = event.detail.selectedEmails
  }

  let errorMessage = '';

  async function handleSubmit() {
    // Reset error message
    errorMessage = '';
    isPersisting = true
    const clientEmailToPost = clientEmail.join(',')

    let projectData = {
      projectName: projectName,
      slackChannel: slackChannel,
      description: description,
      clientEmail: clientEmailToPost,
      hackerName: hackerName.map(item => item.email).join(','),
      isBugBounty: isBugBounty,
      startDate: startDate,
      endDate: endDate,
      color: color,
      ID: model?.ID
    };
    const formData = new FormData();
    formData.append('projectName', projectName);
    formData.append('slackChannel', slackChannel);
    formData.append('description', description);
    formData.append('clientEmail', clientEmailToPost);
    formData.append('hackerName', hackerName.map(item => item.email).join(','));
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
      toast.success('Project successfully stored');

      closeModal();

      // Handle response here
    } catch (error) {
      toast.error('Unable to store new vulnerability');
      console.error('There was a problem storing the new project:', error);
      isPersisting = false
    }
  }
  function toDateOnly(value) {
    if (!value || typeof value !== 'string') return '';
    const prefix = value.slice(0, 10);
    if (prefix === '0001-01-01') return '';
    return prefix;
  }

  onMount(async () => {
    if (model != null) {
      projectName = model.ProjectName;
      slackChannel = model.SlackChannel;
      description = model.Description;
      clientEmailInitilized = splitEmails(model.ClientEmail);
      hackerNameInitilized = splitEmails(model.HackerName);
      isBugBounty = model.IsBugBounty;
      hackers = hackerNameInitilized.map((email, index) => ({ id: index, email }));
      startDate = toDateOnly(model.StartDate);
      endDate = toDateOnly(model.EndDate);
      color = model.Color || '#206bc4';
    }
  });

  function splitEmails(value){
    let emails = value.split(",");
    if(emails.length == 1){
      return [value]
    }
    return emails;
  }
</script>

<Modal bind:showModal on:close={closeModal}>
  {#snippet title()}
    <h5 class="modal-title" >Create new project</h5>
  {/snippet}
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
                        autofocus
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
                    <div class="row mb-3 g-2">
                      <div class="col-md-6">
                        <label for="startDate" class="form-label">Start Date</label>
                        <input
                          id="startDate"
                          type="date"
                          class="form-control"
                          bind:value={startDate}
                        />
                      </div>
                      <div class="col-md-6">
                        <label for="endDate" class="form-label">End Date</label>
                        <input
                          id="endDate"
                          type="date"
                          class="form-control"
                          bind:value={endDate}
                        />
                      </div>
                      <small class="form-hint mt-1">
                        Project lifespan, shown on the planning calendar and list.
                      </small>
                    </div>
                    <div class="mb-3">
                      <label class="form-label">Color</label>
                      <div class="color-swatches">
                        {#each projectColors as swatch}
                          <button
                            type="button"
                            class="color-swatch"
                            class:selected={color === swatch}
                            style="background-color: {swatch}"
                            aria-label="Pick color {swatch}"
                            onclick={() => (color = swatch)}
                          ></button>
                        {/each}
                        <input
                          type="color"
                          class="color-swatch color-custom"
                          bind:value={color}
                          aria-label="Custom color"
                        />
                      </div>
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
                    >Description <span class="form-label-description">{description.length}</span></label
                  >
                  <RichTextEditor
                    bind:value={description}
                    placeholder="Content..."
                    minHeight="160px"
                  />
                </div>
              </div>
            </div>
            <div class="row g-3">
              <div class="col-xl-12">
                <div class="mb-3">
                  <label for="clientEmailInitilized" class="form-label required">Client emails</label>
                  <input
                    type="email"
                    class="form-control"
                    name="clientEmailInitilized"
                    bind:value={clientEmailInitilized}
                    hidden
                  />
                  <UserSearch on:selection={handleClientSearchChange} bind:selectedValues={clientEmailInitilized}/>
                  <small class="form-hint">
                    This refers to the individual or team accountable for overseeing this project.
                    They will serve as the primary point of contact for all project-related matters.
                  </small>
                </div>
                <div class="mb-3">
                  <label for="hackername" class="form-label required">Responsible hackers</label>
                  <Avatarlist hackers="{hackers}" on:updateHackers={handleUserSearchChange}/>
                  <input
                    type="email"
                    class="form-control"
                    name="hackername"
                    bind:value={hackerNameInitilized}
                    hidden
                  />
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

  {#snippet footer()}
    <div class="modal-footer" >
      <a href="#" class="btn btn-link link-secondary" onclick={closeModal}> Cancel </a>
      <button disabled={isPersisting} href="#" class="btn btn-primary ms-auto" onclick={preventDefault(handleSubmit)}>
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
  {/snippet}
</Modal>

<style>
  .color-swatches {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
  }
  .color-swatch {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    border: 2px solid white;
    box-shadow: 0 0 0 1px var(--tblr-border-color);
    cursor: pointer;
    padding: 0;
    transition: transform 0.1s ease;
  }
  .color-swatch:hover {
    transform: scale(1.1);
  }
  .color-swatch.selected {
    box-shadow: 0 0 0 2px var(--tblr-primary);
    transform: scale(1.1);
  }
  .color-custom {
    appearance: none;
    -webkit-appearance: none;
  }
  .color-custom::-webkit-color-swatch-wrapper { padding: 0; }
  .color-custom::-webkit-color-swatch { border: none; border-radius: 50%; }
</style>
