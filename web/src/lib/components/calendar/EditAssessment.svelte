<script>
  import { createEventDispatcher, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { Fetch } from '$lib/fetchUtil.js';
  import AvatarList from '$lib/components/calendar/Avatarlist.svelte';
	import ProjectList from '$lib/components/calendar/ProjectList.svelte';
	import UserSearch from '$lib/components/UserSearch.svelte';
	import Modal from '../Modal.svelte';

  export let assessment
  let localAssessment = assessment;
  const dispatch = createEventDispatcher();

  function updateResponsibleHacker(event) {
    const hackers = event.detail
    const lastOne = hackers[hackers.length-1]
    localAssessment.responsible_hacker = lastOne.email
    responsibleHackers = [lastOne]
  }

  let responsibleHackers = []
  export let showModal = false

  function closeModal() {
		showModal = false;
	}

  async function handleSubmit() {
    await Fetch(`/api/planning/${localAssessment.id}`, {method:"PUT", body: JSON.stringify(localAssessment)})
    dispatch('change', { assessment: localAssessment });
    closeModal()
  }

</script>

<div class="row align-items-center mb-3">
  <div class="col-auto">
    <a href="#" class="btn btn-dark w-100" on:click="{() => goto(`/planning/${data.id}/view`)}">Back</a>
  </div>
  <div class="col-auto">
    <a href="#" class="btn btn-primary w-100" on:click="{() => updateAssessment()}">Save</a>
  </div>
</div>

{#if localAssessment}

<Modal bind:showModal on:close={closeModal}>
	<h5 class="modal-title" slot="title">Edit Assessment</h5>

    <!-- Photo -->
    <!-- <div class="img-responsive img-responsive-21x9 card-img-top" style="background-image: url(/edit-banner.webp)"></div> -->
    <div class="modal-body">
    <div class="ribbon bg-red">EDIT</div>
      <!-- <h3 class="card-title">Edit localAssessment</h3> -->
      <div class="">

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Title</label>
          <div class="col">
            <input autofocus type="text" class="form-control" aria-describedby="emailHelp" placeholder="Enter title" bind:value={localAssessment.title}>
          </div>
        </div>

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Work order</label>
          <div class="col">
            <input type="text" class="form-control" aria-describedby="emailHelp" placeholder="Enter work order" bind:value={localAssessment.workorder}>
          </div>
        </div>

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Estimate</label>
          <div class="col">
            <input type="number" class="form-control" aria-describedby="emailHelp" placeholder="Enter estimation" bind:value={localAssessment.estimate}>
          </div>
        </div>

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Date</label>
          <div class="col">
            <input type="date" class="form-control" aria-describedby="emailHelp" placeholder="Enter start date" bind:value={localAssessment.dateFrom}>
          </div>
          <div class="col">
            <input type="date" class="form-control" aria-describedby="emailHelp" placeholder="Enter end date" bind:value={localAssessment.dateTo}>
          </div>
        </div>

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Requested by</label>
          <div class="col">
            <UserSearch on:selection={e => localAssessment.requester = e.detail.selectedEmails[0]} bind:selectedValues={localAssessment.requester}/>
          </div>
        </div>

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Status</label>
          <div class="col">
            <label class="form-check">
              <input name="radios-inline" class="form-check-input" type="radio" bind:group={localAssessment.status} value="Planning">
              <span class="form-check-label">
                Planning
              </span>
              <span class="form-check-description">
                This localAssessment is currently in the planning phase. Details such as dates, participants, and objectives are under review and subject to change. Please check for updates regularly.
              </span>
            </label>
            <label class="form-check">
              <input name="radios-inline" class="form-check-input" type="radio" bind:group={localAssessment.status} value="Approved">
              <span class="form-check-label">
                Approved
              </span>
              <span class="form-check-description">
                This localAssessment has been finalized and approved. All key details including dates, participants, and objectives have been established and agreed upon. It is now in the implementation phase.
              </span>
            </label>
            <label class="form-check">
              <input name="radios-inline" class="form-check-input" type="radio" bind:group={localAssessment.status} value="Finished">
              <span class="form-check-label">
                Finished
              </span>
              <span class="form-check-description">
                This localAssessment has been successfully completed. All objectives and tasks have been addressed, and the final outcomes are now available for review. Please refer to the provided documentation for detailed results and findings.
              </span>
            </label>
          </div>
        </div>

<hr />

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Responsible</label>
          <div class="col">
            <AvatarList hackers={responsibleHackers} on:updateHackers="{event => updateResponsibleHacker(event)}"/>
          </div>
        </div>

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Projects</label>
          <div class="col">
              <ProjectList projects={localAssessment?.projects} on:updateProjects="{e => localAssessment.projects = e.detail}"/>
          </div>
        </div>

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Hackers</label>
          <div class="col">
            <AvatarList hackers={localAssessment.hackers} on:updateHackers="{e => localAssessment.hackers = e.detail}"/>
          </div>
        </div>

<hr />
                <div class="mb-3 row">
          <div class="col">
            			<textarea
                    class="form-control"
                    rows=10
                    placeholder="Notes..."
                    bind:value={localAssessment.note} />
          </div>
        </div>

      </div>
    </div>
	<div class="modal-footer" slot="footer">
		<a href="#" class="btn btn-link link-secondary" on:click={closeModal}> Cancel </a>
		<button href="#" class="btn btn-primary ms-auto" on:click|preventDefault={handleSubmit}>
			<!-- Download SVG icon from http://tabler-icons.io/i/plus -->
				Save
		</button>
	</div>

</Modal>
{/if}