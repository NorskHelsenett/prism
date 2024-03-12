<script>
import { goto } from '$app/navigation';
import { slide } from 'svelte/transition';
import { quintOut } from 'svelte/easing';
import Avatar from '$lib/components/Avatar.svelte';
import AvatarList from '$lib/components/calendar/Avatarlist.svelte';
import DeleteModal from '$lib/components/DeleteModal.svelte';
import Dropdown from '$lib/components/Dropdown.svelte';
import Icon from '$lib/components/Icon.svelte';
import Markdown from '$lib/components/Markdown.svelte';
import Criticality from '$lib/components/dashboard/Criticality.svelte';
import CountVulnerabilities from '$lib/components/dashboard/countVulnerabilities.svelte';
import { Fetch } from '$lib/fetchUtil';
import { accessLevels } from '$lib/userStore';
import { onMount } from 'svelte';
import ProjectList from '$lib/components/calendar/ProjectList.svelte';
import Vulnerability from '$lib/components/Lists/Vulnerability.svelte';
import OwaspTable from '$lib/components/dashboard/OwaspTable.svelte';
import CriticalityPie from '$lib/components/charts/CriticalityPie.svelte';
	import EditAssessment from '$lib/components/calendar/EditAssessment.svelte';

export let data;
let assessment = null
let showDropdown = false;
let showDeleteModal = false;
let editModus = false;
let vulnerabilities = []

onMount(async () => {
  assessment = await Fetch(`/api/planning/${data.id}`)
  await getUser(assessment.responsible_hacker)
  assessment.projects.forEach(async project => {
    const vulns = await Fetch(`/api/project/${project.id}/vulnerabilities?from=${assessment.dateFrom}&to=${assessment.dateTo}`)
    vulnerabilities = [...vulnerabilities, ...vulns]
  });
})

async function deleteProject(){
  await Fetch(`/api/planning/${data.id}`, {method: "DELETE"})
  goto("/planning")
}

async function getUser(email) {
  const user = await Fetch(`/api/profile/${email}`);
  responsibleHackerName = user.name
}

$: responsibleHackerName = ""

let showModalEdit = false

let modalEditProp = null
let modalEditValue = ""

function editProp(prop, value){
  if(!editModus) { return }
  showModalEdit = true
  modalEditProp = prop
  modalEditValue = value
}

async function saveChanges(prop, value) {
  // Clone the assessment object to avoid mutating the original state
  const updatedAssessment = { ...assessment };
  if(prop != null){
    assessment[prop] = value;
  }
  updatedAssessment.estimate = Number(updatedAssessment.estimate)
  updatedAssessment.hackers = []
  assessment.hackers.forEach((/** @type {{ Email: any; email: any; }} */ hacker) => {
    updatedAssessment.hackers.push({"email": hacker.email})
  });
  await Fetch(`/api/planning/${data.id}`, {method: "PUT", body: JSON.stringify(updatedAssessment)})
  showModalEdit = false
}

async function handleKeydown(event) {
  if (event.key === 'Enter') {
    // Enter key was pressed
    await saveChanges(modalEditProp, modalEditValue)
  }
}

  function handleUpdateHackers(event) {
    assessment.hackers = event.detail; // Assuming event.detail contains the updated hackers array
  }

function formatDate(dateStr){
  const options = { day: 'numeric', month: 'short' };
  const date = new Date(dateStr);
  return new Intl.DateTimeFormat('en-GB', options).format(date).replace(/\s/g, '. ');
}

let showEditModal = false
</script>

{#if showEditModal}
  <EditAssessment {assessment} bind:showModal={showEditModal} on:change="{e => assessment = e.detail.assessment}"/>
{/if}

{#if showModalEdit && editModus}
<div class="modal modal-blur fade show" id="modal-small" tabindex="-1" style="display: block;" aria-modal="true" role="dialog">
  <div class="modal-dialog modal-sm modal-dialog-centered" role="document">
    <div class="modal-content">
      <div class="modal-body">
        <!-- Bind the input to modalEditValue instead of modalEditProp -->
        <input autofocus bind:value={modalEditValue} class="form-control" on:keydown={handleKeydown}/>
      </div>
      <div class="modal-footer">
        <!-- Update showModalEdit and optionally save changes on Save button click -->
        <button type="button" class="btn btn-link link-secondary me-auto" on:click="{() => showModalEdit = false}">Cancel</button>
        <button type="button" class="btn btn-info" on:click="{() => { saveChanges(modalEditProp, modalEditValue); }}">Save</button>
      </div>
    </div>
  </div>
</div>
{/if}

<div class="row g-2 align-items-center mb-3">
  <div class="col">
    <div class="page-pretitle">Assessment</div>
    <h2 class="page-title" on:click={() => editProp('title', assessment.title)}>
      {assessment?.title}
    </h2>
  </div>
{#if $accessLevels["/planning"]?.write}
  <div class="col-auto ms-auto d-print-non">
    <div class="btn-list">
      <a
        href="#"
        class="d-none d-sm-inline-block"
        on:click={() => showDropdown = !showDropdown}
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-dots-vertical" width="36" height="36" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M12 12m-1 0a1 1 0 1 0 2 0a1 1 0 1 0 -2 0" /><path d="M12 19m-1 0a1 1 0 1 0 2 0a1 1 0 1 0 -2 0" /><path d="M12 5m-1 0a1 1 0 1 0 2 0a1 1 0 1 0 -2 0" /></svg>
          </a
          >
        </div>

        <Dropdown bind:show={showDropdown}>
          <a class="dropdown-item" href="#" on:click={()=> {editModus = !editModus; showDropdown = false; goto(`/planning/${data.id}/edit`)}}>Edit</a>
          <!-- <a class="dropdown-item" href="#" on:click={()=> {showEditModal = !showEditModal; showDropdown = false}}>Edit</a> -->
          <div class="dropdown-divider"></div>
          <a class="dropdown-item text-warning" href="#" on:click={() => showDeleteModal = true}>
            <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-trash" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M4 7l16 0" /><path d="M10 11l0 6" /><path d="M14 11l0 6" /><path d="M5 7l1 12a2 2 0 0 0 2 2h8a2 2 0 0 0 2 -2l1 -12" /><path d="M9 7v-3a1 1 0 0 1 1 -1h4a1 1 0 0 1 1 1v3" /></svg>
            Delete
          </a>
        </Dropdown>

  </div>
{/if}

</div>

{#if editModus}
  <div class="card alert alert-info alert-dismissible" role="alert" transition:slide={{ delay: 100, duration: 300, easing: quintOut, axis: 'y' }}>
  <div class="d-flex">
    <div>
      <!-- Download SVG icon from http://tabler-icons.io/i/info-circle -->
      <svg xmlns="http://www.w3.org/2000/svg" class="icon alert-icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M3 12a9 9 0 1 0 18 0a9 9 0 0 0 -18 0"></path><path d="M12 9h.01"></path><path d="M11 12h1v4h1"></path></svg>
    </div>
    <div>
      <h4 class="alert-title">Edit mode!</h4>
      <div class="text-secondary">You've found the edit mode. Now you can make changes to your assessment.</div>
    </div>
  </div>
  <a class="btn-close" data-bs-dismiss="alert" aria-label="close" on:click="{() => {saveChanges(null,null); editModus = false}}"></a>
</div>
{/if}

<div class="row-deck row-cards" >
	<Criticality {vulnerabilities}/>
</div>

{#if assessment}

<div class="row row-deck row-cards">
  <div class="col12">
    <div class="row row-cards">
      <div class="col-sm-3 col-lg-3">
        <div class="card card-sm">
          <CountVulnerabilities data={vulnerabilities.length} />
        </div>
      </div>

      <div class="col-sm-3 col-lg-3">
        <div class="card card-sm">
          <div class="card-body">
            <div class="row align-items-center">
              <div class="col-auto">
                <span class="bg-orange text-white avatar">
                  <Icon icon="calendar-week" />
                </span>
              </div>
              <div class="col">
                <div class="font-weight-medium">
                  {formatDate(assessment?.dateFrom)}
                </div>
                <div class="text-secondary">{formatDate(assessment?.dateTo)}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- <div class="col-sm-3 col-lg-3">
        <div class="card card-sm">
          <div class="card-body">
            <div class="row align-items-center">
              <div class="col-auto">
                <span class="bg-teal text-white avatar">
                  <Icon icon="calendar-check" />
                </span>
              </div>
              <div class="col">
                <div class="font-weight-medium">
                  {assessment?.dateTo}
                </div>
                <div class="text-secondary">Planned finished by</div>
              </div>
            </div>
          </div>
        </div>
      </div> -->

      <div class="col-sm-3 col-lg-3">
        <div class="card card-sm">
          <div class="card-body">
            <div class="row align-items-center">
              <div class="col-auto">
                <span class="bg-teal text-white avatar">
                  <Icon icon="clock-dollar" />
                </span>
              </div>
              <div class="col">
                <div class="font-weight-medium">
                  {assessment.estimate} h
                </div>
                <div class="text-secondary">Estimate</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="col-sm-3 col-lg-3">
        <div class="card card-sm">
          <div class="card-body">
            <div class="row">
              <div class="col-auto">
                <Avatar email="{assessment?.responsible_hacker}" option={{ showName: false, size: "md", emptyFields: false, circle: true}}/>
              </div>
              <div class="col">
                <div class="font-weight-medium">{responsibleHackerName}</div>
                <div class="text-secondary">Responsible</div>
              </div>
            </div>
          </div>
        </div>
      </div>

  <div class="col-8">
    <div class="card h-100">
      <div class="card-body" style="min-height: 10rem">
        <div class="datagrid">

          <div class="datagrid-item" on:click={() => editProp('workorder', assessment.workorder)}>
            <div class="datagrid-title">Work order</div>
            <div class="datagrid-content">
              {assessment?.workorder || ""}
            </div>
          </div>

        <div class="datagrid-item">
          <div class="datagrid-title">Projects</div>
          <div class="datagrid-content">
            {#if editModus}
              <ProjectList projects={assessment.projects} on:updateProjects="{e => assessment.projects = e.detail}"/>
            {:else}
              {#each assessment.projects as project}
                <div class="badge bg-cyan-lt mt-1 mr-1"><a href="/project/{project.id}/view">{project.name}</a></div>
              {/each}
            {/if}
          </div>
        </div>

        <div class="datagrid-item">
          <div class="datagrid-title">Assigned Hackers</div>
          <div class="datagrid-content">
            {#if editModus}
              <AvatarList hackers={assessment.hackers} on:updateHackers="{handleUpdateHackers}"/>
            {:else}
              {#each assessment.hackers as hacker }
                <Avatar email={hacker?.email || hacker?.email} option={{ showName: false, circle: true, size: "sm" }}/>
              {/each}
            {/if}
          </div>
        </div>

        <div class="datagrid-item">
          <div class="datagrid-title">Responsible</div>
          <div class="datagrid-content">
            {#if editModus}
              <AvatarList hackers={[{"email": assessment.responsible_hacker}]} on:updateHackers="{handleUpdateHackers}"/>
            {:else}
              <Avatar email={assessment.responsible_hacker} option={{ showName: false, circle: true, size: "sm" }}/>
            {/if}
          </div>
        </div>

        <div class="datagrid-item">
          <div class="datagrid-title">Requested by</div>
          <div class="datagrid-content">
            <Avatar email={assessment.requester} option={{ emptyFields: true, showName: false, circle: true, size: "sm" }}/>
          </div>
        </div>

        <div class="datagrid-item">
          <div class="datagrid-title">Status</div>
          <div class="datagrid-content">
            <span style="margin:0;">
            {#if assessment.status == "Finished"}
              <i class="ti ti-circle-check-filled text-azure" title="Finished"></i>
            {:else if assessment.status == "Approved"}
              <i class="ti ti-player-record-filled text-green" title="Approved"></i>
            {:else}
              <i class="ti ti-circle text-yellow"  title="Planning"></i>
            {/if}
          </span>
            {assessment.status || "Planning" }
          </div>
        </div>

        </div>
        <div class="row mt-3">
            <div class="datagrid-title">Note</div>
            {#if editModus}
              <textarea
                class="form-control"
                placeholder="Notes..."
                bind:value={assessment.note}
              />
            {:else}
              <Markdown markdown={assessment.note}/>
            {/if}
        </div>
      </div>
    </div>
  </div>

  <div class="col-4">
    <div class="card h-100">
        <div class="card-body " style="min-height: 10rem">
        <CriticalityPie {vulnerabilities}/>
      </div>
    </div>
  </div>

  <div class="col-5">
    <div class="card h-100" >
      <div class="card-body">
        <!-- <div class="grid-container">
          <i class="ti ti-player-record-filled text-green" title="Approved"></i>
          <div>Team size</div>
          <i class="ti ti-player-record-filled text-green" title="Approved"></i>
          <div>It is approved</div>
          <i class="ti ti-player-record-filled text-green" title="Approved"></i>
          <div>All team members are available</div>
        </div> -->
      </div>
    </div>
  </div>


  <div class="col-7" >
    <div class="card h-100" >
      <div class="card-body">
        <div class="table-responsive w-100">
          <OwaspTable {vulnerabilities}/>
        </div>
      </div>
    </div>
  </div>

  <div class="col-12">
    <div class="card">
      <div class="card-body">
        <Vulnerability {vulnerabilities} />
      </div>
    </div>
  </div>

    </div>
  </div>
</div>
{/if}

<DeleteModal bind:showDeleteModal onDelete={deleteProject} deleteButtonText="Yes, delete it!" text="Delete this assessment. This action is irreversible."/>

<style>
  .alert {
    background-color: var(--tblr-card-bg);
  }
  .mr-1{
    margin-right: 0.5em;
  }
  .grid-container {
    display: grid; /* Use CSS Grid to create the layout */
    grid-template-columns: 0.05fr 1fr; /* Create two columns of equal width */
    align-items: center; /* Vertically center align items */
  }
</style>