<script>
  import { onMount } from 'svelte';
  import { Fetch } from '$lib/fetchUtil'
	import { formatDateToYYYYMMDD } from '$lib/utils';
	import CriticalityPie from '$lib/components/charts/CriticalityPie.svelte';
	import Criticality from '$lib/components/dashboard/Criticality.svelte';
	import CountVulnerabilities from '$lib/components/dashboard/countVulnerabilities.svelte';
	import Tasks from '$lib/components/dashboard/Tasks.svelte';
	import Avatar from '$lib/components/Avatar.svelte';
	import Vulnerability from '$lib/components/Lists/Vulnerability.svelte';
	import Assessments from '$lib/components/dashboard/Assessments.svelte';
	import Dropdown from '$lib/components/Dropdown.svelte';

  export let data;

  let project
  let vulnerabilitiesTotal
  let vulnerabilities = []

  onMount(async () => {
    const response = await Fetch(`/api/project/${data.id}`);
    project = response
    const total = await Fetch(`/api/project/${project.ID}/vulnerabilities/total`)
    vulnerabilitiesTotal = total.total_vulnerabilities

    vulnerabilities = await Fetch(`/api/project/${project.ID}/vulnerabilities`)
  });

  async function fetchUserData(email) {
    return await Fetch(`/api/userinfo/${email}`);
  }

let severityData = {
    critical: 0,
    high: 0,
    medium: 0,
    low: 0,
    information: 0
  };

  $: {
    // Reset counts
    severityData = { critical: 0, high: 0, medium: 0, low: 0, information: 0 };

    vulnerabilities.forEach(vulnerability => {
      const criticality = vulnerability.Vulnerability.criticality.toLowerCase();

      if (criticality === 'critical') {
        severityData.critical += 1;
      } else if (criticality === 'high') {
        severityData.high += 1;
      } else if (criticality === 'medium') {
        severityData.medium += 1;
      } else if (criticality === 'low') {
        severityData.low += 1;
      } else if (criticality === 'information') {
        severityData.information += 1;
      }
    });
  }

  async function deleteProject(id){
    showDeleteModal = true
    await Fetch(`/api/project/${id}`, {method: "DELETE"})
    window.location.href ="/project"
  }

  let assessments = 0;
  let showDropdown = false
  let showDeleteModal = false

</script>

<style>
  .fs-xx-small {
    font-size: xx-small;
    vertical-align: super;
  }
  .ml-10px {
    margin-left: 10px;
  }
</style>

{#if project}

				<div class="row g-2 align-items-center">
					<div class="col">
						<div class="page-pretitle">Project</div>
						<h2 class="page-title">
              {project.ProjectName}
              {#if project.IsBugBounty}
              <span class="badge badge-outline text-azure fs-xx-small ml-10px">Bug Bounty</span>
              {/if}
            </h2>
					</div>

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
                  <a class="dropdown-item" href="#">Edit</a>
                  <div class="dropdown-divider"></div>
                  <a class="dropdown-item text-warning" href="#" on:click={() => showDeleteModal = true}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-trash" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M4 7l16 0" /><path d="M10 11l0 6" /><path d="M14 11l0 6" /><path d="M5 7l1 12a2 2 0 0 0 2 2h8a2 2 0 0 0 2 -2l1 -12" /><path d="M9 7v-3a1 1 0 0 1 1 -1h4a1 1 0 0 1 1 1v3" /></svg>
                    Delete
                  </a>
                </Dropdown>

					</div>

              <div class="page-body">
						<div class="container-xl">
						</div>
					</div>
				</div>

<div class="row-deck row-cards" >
	<Criticality {severityData}/>
</div>

		<div class="row row-deck row-cards">
			<div class="col12">
				<div class="row row-cards">
					<div class="col-sm-4 col-lg-4">
						<div class="card card-sm">
							<CountVulnerabilities data={vulnerabilitiesTotal} />
						</div>
					</div>

          <div class="col-sm-4 col-lg-4">
						<div class="card card-sm">
							<Tasks />
						</div>
					</div>

          <div class="col-sm-4 col-lg-4">
						<div class="card card-sm">
							<Assessments data={assessments} />
						</div>
					</div>

			<div class="col-8">
				<div class="card">
					<div class="card-body" style="height: 10rem">
            <div class="datagrid">

            <div class="datagrid-item">
              <div class="datagrid-title">Slack Channel</div>
              <div class="datagrid-content">{project.SlackChannel || 'N/A'}</div>
            </div>

            <div class="datagrid-item">
              <div class="datagrid-title">Client Email</div>
              <div class="datagrid-content">
                <Avatar email={project.ClientEmail} option={{ showName: false }}/>
              </div>
            </div>

            <div class="datagrid-item">
              <div class="datagrid-title">Assigned Hackers</div>
              <div class="datagrid-content">
                <Avatar email={project.HackerName} option={{ showName: false }}/>
              </div>
            </div>
            </div>
            <div class="row mt-3">
                <div class="datagrid-title">Description</div>
                <div class="datagrid-content">{project.Description || 'N/A'}</div>
            </div>
					</div>
				</div>
			</div>

			<div class="col-4">
				<div class="card">
					<div class="card-body" style="height: 10rem">
            <CriticalityPie {severityData} />
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

{#if showDeleteModal}
<div class="modal modal-blur fade show" id="modal-small" tabindex="-1" role="dialog" style="display: block;" aria-modal="true">
  <div class="modal-dialog modal-sm modal-dialog-centered" role="document">
    <div class="modal-content">
      <div class="modal-body">
        <div class="modal-title">Are you sure?</div>
        <div>If you proceed, all data related to this project will be permanently lost. This action is irreversible.</div>
      </div>
      <div class="modal-footer">
        <button type="button" class="btn btn-link link-secondary me-auto" data-bs-dismiss="modal" on:click={showDeleteModal = false}>Cancel</button>
        <button type="button" class="btn btn-danger" data-bs-dismiss="modal" on:click={deleteProject(project.ID)}>Yes, delete the project</button>
      </div>
    </div>
  </div>
</div>
{/if}

{:else}
  <p>Loading project details...</p>
{/if}


