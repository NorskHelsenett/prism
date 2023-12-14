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

  let assessments = 0;
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
        <div class="col-2">
          <div class="datagrid-title">Description</div>
        </div>
        <div class="col-10">
          <div class="datagrid-content">{project.Description || 'N/A'}</div>
        </div>
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
{:else}
  <p>Loading project details...</p>
{/if}