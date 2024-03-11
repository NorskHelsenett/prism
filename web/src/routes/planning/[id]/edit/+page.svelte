<script>
	import { goto } from '$app/navigation';
	import { Fetch } from '$lib/fetchUtil.js';
  import AvatarList from '$lib/components/calendar/Avatarlist.svelte';
	import { onMount } from 'svelte';
	import ProjectList from '$lib/components/calendar/ProjectList.svelte';
	import UserSearch from '$lib/components/UserSearch.svelte';
  	import { slide } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import Avatar from '$lib/components/Avatar.svelte';
	import { notification } from '$lib/stores/notificationStore.js';

  export let data;
  export let assessment

  onMount(async () => {
    assessment = await Fetch(`/api/planning/${data.id}`)
    if (!assessment.status) {
      assessment.status = 'Planning';
    }
    responsibleHackers = [{email: assessment.responsible_hacker}]
  })

  async function updateAssessment() {
    const result = await Fetch(`/api/planning/${data.id}`, {method:"PUT", body: JSON.stringify(assessment)})
    if (result.error) {
      notification.addAlert({type: "alert", title: "Error", message: result.error})
    } else {
      notification.addAlert("Saved")
      goto(`/planning/${assessment.id}/view`)
    }
  }

  function updateResponsibleHacker(event) {
    const hackers = event.detail
    const lastOne = hackers[hackers.length-1]
    assessment.responsible_hacker = lastOne.email
    responsibleHackers = [lastOne]
  }

let responsibleHackers = []
let busyHackers = []
$: teamSize = calculateTeamSize();
$: assessment && findAvailableUsers();
$: if(assessment){
  teamSize = calculateTeamSize();
}

async function findAvailableUsers() {
  busyHackers = await Fetch(`/api/planning/${assessment.id}/assignedHackers?from=${assessment.dateFrom}&to=${assessment.dateTo}`)
}

function calculateTeamSize() {
  if(!assessment ) {return 0}
  const start = new Date(assessment.dateFrom); // Convert string to Date
  const end = new Date(assessment.dateTo); // Convert string to Date
  let numberOfDays = 0;

  for (let d = new Date(start); d <= end; d.setDate(d.getDate() + 1)) {
    // Get the day of the week: 0 for Sunday, 6 for Saturday
    const dayOfWeek = d.getDay();
    // Increment numberOfDays if it's not Saturday (6) or Sunday (0)
    if (dayOfWeek !== 0 && dayOfWeek !== 6) {
      numberOfDays++;
    }
  }

  const workingHours = 7;

  // Ensure not to divide by zero
  if (numberOfDays === 0) {
    console.error("No working days in the range. Please check the dates.");
    return 0;
  }

  return Math.floor(assessment.estimate / (numberOfDays * workingHours)) || 1;
}

function isAvailable(user) {
  console.log(user)
  return false
}
</script>

<div class="row align-items-center mb-3">
  <div class="col-auto">
    <button href="#" class="btn btn-dark w-100 link" on:click="{() => goto(`/planning/${data.id}/view`)}">Back</button>
  </div>
  <div class="col-auto">
    <a href="#" class="btn btn-primary w-100" on:click="{() => updateAssessment()}">Save</a>
  </div>
</div>

{#if assessment}

<div class="row">
<div class="card">
  <!-- Photo -->
  <div class="ribbon bg-red">EDIT</div>
                  <!-- <div class="img-responsive img-responsive-21x9 card-img-top" style="background-image: url(/edit-banner.webp)"></div> -->
    <div class="card-body">
      <h3 class="card-title">Edit assessment</h3>
      <div class="card-body">

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Title</label>
          <div class="col">
            <input type="text" class="form-control" aria-describedby="emailHelp" placeholder="Enter title" bind:value={assessment.title}>
          </div>
        </div>

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Work order</label>
          <div class="col">
            <input type="text" class="form-control" aria-describedby="emailHelp" placeholder="Enter work order" bind:value={assessment.workorder}>
          </div>
        </div>

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Estimate</label>
          <div class="col">
            <input type="number" class="form-control" aria-describedby="emailHelp" placeholder="Enter estimation" bind:value={assessment.estimate}>
          </div>
        </div>

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Date</label>
          <div class="col">
            <input type="date" class="form-control" aria-describedby="emailHelp" placeholder="Enter start date" bind:value={assessment.dateFrom}>
          </div>
          <div class="col">
            <input type="date" class="form-control" aria-describedby="emailHelp" placeholder="Enter end date" bind:value={assessment.dateTo}>
          </div>
        </div>

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Requested by</label>
          <div class="col">
            <UserSearch on:selection={e => assessment.requester = e.detail.selectedEmails[0]} bind:selectedValues={assessment.requester}/>
          </div>
        </div>

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Status</label>
          <div class="col">
            <label class="form-check">
              <input name="radios-inline" class="form-check-input" type="radio" bind:group={assessment.status} value="Planning">
              <span class="form-check-label">
                Planning
              </span>
              <span class="form-check-description">
                This assessment is currently in the planning phase. Details such as dates, participants, and objectives are under review and subject to change. Please check for updates regularly.
              </span>
            </label>
            <label class="form-check">
              <input name="radios-inline" class="form-check-input" type="radio" bind:group={assessment.status} value="Approved">
              <span class="form-check-label">
                Approved
              </span>
              <span class="form-check-description">
                This assessment has been finalized and approved. All key details including dates, participants, and objectives have been established and agreed upon. It is now in the implementation phase.
              </span>
            </label>
            <label class="form-check">
              <input name="radios-inline" class="form-check-input" type="radio" bind:group={assessment.status} value="Finished">
              <span class="form-check-label">
                Finished
              </span>
              <span class="form-check-description">
                This assessment has been successfully completed. All objectives and tasks have been addressed, and the final outcomes are now available for review. Please refer to the provided documentation for detailed results and findings.
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
              <ProjectList projects={assessment?.projects} on:updateProjects="{e => assessment.projects = e.detail}"/>
          </div>
        </div>

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Hackers</label>
          <div class="col">
            <AvatarList hackers={assessment.hackers} on:updateHackers="{e => assessment.hackers = e.detail}"/>
            {#if assessment.hackers.length != teamSize}
            <div class="alert alert-info mt-3" role="alert" transition:slide={{ delay: 10, duration: 300, easing: quintOut, axis: 'y' }}>
                    <div class="d-flex">
                      <div>
                        <!-- Download SVG icon from http://tabler-icons.io/i/info-circle -->
                        <svg xmlns="http://www.w3.org/2000/svg" class="icon alert-icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M3 12a9 9 0 1 0 18 0a9 9 0 0 0 -18 0"></path><path d="M12 9h.01"></path><path d="M11 12h1v4h1"></path></svg>
                      </div>
                      <div>
                        <h4 class="alert-title">Did you know?</h4>
                        <div class="text-secondary">Based on the estimate and timeframe, the recommended team size is {teamSize}</div>
                      </div>
                    </div>
                  </div>
            {/if}
            {#each assessment.hackers as hacker}
            {#if busyHackers?.includes(hacker.email)}

                <div class="alert alert-warning" role="alert">
                      <div class="d-flex">
                        <div>
                          <!-- Download SVG icon from http://tabler-icons.io/i/alert-triangle -->
                          <svg xmlns="http://www.w3.org/2000/svg" class="icon alert-icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M12 9v4"></path><path d="M10.363 3.591l-8.106 13.534a1.914 1.914 0 0 0 1.636 2.871h16.214a1.914 1.914 0 0 0 1.636 -2.87l-8.106 -13.536a1.914 1.914 0 0 0 -3.274 0z"></path><path d="M12 16h.01"></path></svg>
                        </div>
                        <div>
                          <Avatar email={hacker.email} option={{ emptyFields: true, showName: true, circle: true, size: "xs" }}/>
                          <span class="align-middle"> is unavaible</span>
                        </div>
                      </div>
                    </div>
                    {/if}
            {/each}
          </div>

        </div>

<hr />
                <div class="mb-3 row">
          <div class="col">
            			<textarea
                    class="form-control"
                    rows=10
                    placeholder="Notes..."
                    bind:value={assessment.note} />
          </div>
        </div>

      </div>
    </div>
  </div>
</div>

{/if}