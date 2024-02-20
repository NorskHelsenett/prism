<script>
import { goto } from '$app/navigation';

import Avatar from '$lib/components/Avatar.svelte';
import DeleteModal from '$lib/components/DeleteModal.svelte';
import Dropdown from '$lib/components/Dropdown.svelte';
import Icon from '$lib/components/Icon.svelte';
	import Markdown from '$lib/components/Markdown.svelte';
import Criticality from '$lib/components/dashboard/Criticality.svelte';
import CountVulnerabilities from '$lib/components/dashboard/countVulnerabilities.svelte';
import { Fetch } from '$lib/fetchUtil';
import { accessLevels } from '$lib/userStore';
import { onMount } from 'svelte';

export let data;
let assessment = null
let showDropdown = false;
let showDeleteModal = false;
let showEditModal = false;
let severityData = { critical: 1, high: 1, medium: 2, low: 0, information: 0 }
let vulnerabilitiesTotal = 10

async function deleteProject(){
  await Fetch(`/api/planning/${data.id}`, {method: "DELETE"})
  goto("/planning")
}

onMount(async () => {
  assessment = await Fetch(`/api/planning/${data.id}`)
  await getUser(assessment.responsible_hacker)
})

  async function getUser(email) {
    const user = await Fetch(`/api/profile/${email}`);
    responsibleHackerName = user.Name
  }

$: responsibleHackerName = ""
</script>

<div class="row g-2 align-items-center">
  <div class="col">
    <div class="page-pretitle">Assessment</div>
    <h2 class="page-title">
      {assessment?.description}
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
          <a class="dropdown-item" href="#" on:click={()=> showEditModal = !showEditModal}>Edit</a>
          <div class="dropdown-divider"></div>
          <a class="dropdown-item text-warning" href="#" on:click={() => showDeleteModal = true}>
            <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-trash" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M4 7l16 0" /><path d="M10 11l0 6" /><path d="M14 11l0 6" /><path d="M5 7l1 12a2 2 0 0 0 2 2h8a2 2 0 0 0 2 -2l1 -12" /><path d="M9 7v-3a1 1 0 0 1 1 -1h4a1 1 0 0 1 1 1v3" /></svg>
            Delete
          </a>
        </Dropdown>

  </div>
{/if}
      <div class="page-body">
    <div class="container-xl">
    </div>
  </div>
</div>

<div class="row-deck row-cards" >
	<Criticality {severityData}/>
</div>

{#if assessment}

<div class="row row-deck row-cards">
  <div class="col12">
    <div class="row row-cards">
      <div class="col-sm-3 col-lg-3">
        <div class="card card-sm">
          <CountVulnerabilities data={vulnerabilitiesTotal} />
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
                  {assessment?.dateFrom}
                </div>
                <div class="text-secondary">Planned starting date</div>
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
                  450 h
                </div>
                <div class="text-secondary">Estimate</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="col-sm-3 col-lg-3">
        <div class="card card-sm">

          <div class="card-body placeholder-glow">
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

          <div class="datagrid-item">
            <div class="datagrid-title">AO</div>
            <div class="datagrid-content">
              {assessment?.ao || ""}
            </div>
          </div>

          <div class="datagrid-item">
            <div class="datagrid-title">Estimate</div>
            <div class="datagrid-content">
              {assessment?.estimate || ""}
            </div>
          </div>


        <div class="datagrid-item">
          <div class="datagrid-title">Projects</div>
          <div class="datagrid-content">
            {#each assessment.projects as project}
              <span>{project.name}</span>
            {/each}
          </div>
        </div>

        <div class="datagrid-item">
          <div class="datagrid-title">Assigned Hackers</div>
          <div class="datagrid-content">
            {#each assessment.hackers as hacker }
              <Avatar email={hacker.email.trim()} option={{ showName: false, circle: true, size: "sm" }}/>
            {/each}
          </div>
        </div>
        </div>
        <div class="row mt-3">
            <div class="datagrid-title">Note</div>
              <Markdown markdown={assessment.note}/>
        </div>
      </div>
    </div>
  </div>

  <div class="col-4 h-100">
    <div class="card">
<div class="card-body text-end placeholder-glow">
                        <div class="placeholder col-9 mb-3"></div>
                        <div class="placeholder placeholder-xs col-10"></div>
                        <div class="placeholder placeholder-xs col-12"></div>
                        <div class="placeholder placeholder-xs col-11"></div>
                        <div class="placeholder placeholder-xs col-8"></div>
                        <div class="placeholder placeholder-xs col-10"></div>
                      </div>

    </div>
  </div>

  <div class="col-5">
    <div class="card h-100" >
      <div class="card-body">
          <!-- <EndpointVulnerability dataset={vulnerabilities}/> -->
      </div>
    </div>
  </div>


  <div class="col-7" >
    <div class="card h-100" >
      <div class="card-body">
        <div class="table-responsive w-100">
          <!-- <OwaspTable owaspData={owaspData}/> -->
        </div>
      </div>
    </div>
  </div>

  <div class="col-12">
    <div class="card">
      <div class="card-body">
        <!-- <Vulnerability {vulnerabilities} /> -->
      </div>
    </div>
  </div>

    </div>
  </div>
</div>
{/if}

<DeleteModal bind:showDeleteModal onDelete={deleteProject} deleteButtonText="Yes, delete it!" text="Delete this assessment. This action is irreversible."/>