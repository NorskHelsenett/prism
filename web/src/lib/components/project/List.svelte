<script>
	import { goto } from '$app/navigation';
  import { Fetch } from '$lib/fetchUtil.js';
	import Avatar from '../Avatar.svelte';

  let projects = $state([]); // Reactive variable to store projects

  function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-CA', { // 'en-CA' format is 'YYYY-MM-DD'
        year: 'numeric',
        month: '2-digit',
        day: '2-digit'
    }).replace(/-/g, '.'); // Replace '-' with '.'
  }

    async function getTotalVulnerabilites(projectID) {
      const total = await Fetch(`/api/project/${projectID}/vulnerabilities/total`);
      if (total) {
        return total.total_vulnerabilities;
      }
      return 0;
    }

export function refreshList() {
    // The logic previously in onMount
    async function fetchData() {
      projects = await Fetch("/api/project/all");
    }

    fetchData();
  }

  let selectedRow = $state(-1)
</script>

<div class="card mt-4">
	<div class="card-header">
    <div class="table-responsive col-12">
      <table class="table table-vcenter card-table">
        <tbody>
      {#each projects as project, index}
          <tr ondblclick={() => goto(`/project/${project.ID}/view`)} onclick={() => selectedRow === index ? selectedRow = -1 : selectedRow = index} class:selected="{selectedRow === index}">
              <td>
                <div class="flex-fill">
                  <div class="font-weight-medium">
                    <button class="link" onclick={() => goto(`/project/${project?.ID}/view`)}>
                      {project.ProjectName}
                    </button>
                  </div>
                </div>
              </td>
              <td>
                  <div class="flex-fill">
                    <div class="text-secondary">
                      <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-calendar" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M4 7a2 2 0 0 1 2 -2h12a2 2 0 0 1 2 2v12a2 2 0 0 1 -2 2h-12a2 2 0 0 1 -2 -2v-12z" /><path d="M16 3v4" /><path d="M8 3v4" /><path d="M4 11h16" /><path d="M11 15h1" /><path d="M12 15v3" /></svg>
                      {formatDate(project.CreatedAt)}</div>
                  </div>
              </td>
              <td>
                  <div class="flex-fill">
                    {#if project.IsBugBounty}
                    <span class="badge badge-outline text-azure">
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-bug" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M9 9v-1a3 3 0 0 1 6 0v1" /><path d="M8 9h8a6 6 0 0 1 1 3v3a5 5 0 0 1 -10 0v-3a6 6 0 0 1 1 -3" /><path d="M3 13l4 0" /><path d="M17 13l4 0" /><path d="M12 20l0 -6" /><path d="M4 19l3.35 -2" /><path d="M20 19l-3.35 -2" /><path d="M4 7l3.75 2.4" /><path d="M20 7l-3.75 2.4" /></svg>
                    Bug Bounty
</span>
                    {/if}
                  </div>
              </td>
              <td>
                  <div class="flex-fill">
                    <div class="text-secondary">
                      <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-briefcase" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M3 7m0 2a2 2 0 0 1 2 -2h14a2 2 0 0 1 2 2v9a2 2 0 0 1 -2 2h-14a2 2 0 0 1 -2 -2z" /><path d="M8 7v-2a2 2 0 0 1 2 -2h4a2 2 0 0 1 2 2v2" /><path d="M12 12l0 .01" /><path d="M3 13a20 20 0 0 0 18 0" /></svg>
                      <div class="avatar-list avatar-list-stacked">
                        {#each project.ClientEmail.split(',').map(e => e.trim()).filter(Boolean).slice(0, 5) as email (email)}
                          <Avatar email={email} option={{ showName: false, emptyFields: true, size: 'xs', circle: true }}/>
                        {/each}
                        {#if project.ClientEmail.split(',').filter(Boolean).length > 5}
                          <span class="avatar avatar-xs rounded-circle">+{project.ClientEmail.split(',').filter(Boolean).length - 5}</span>
                        {/if}
                      </div>
                    </div>
                  </div>
              </td>
              <td>
                  <div class="flex-fill">
                    <div class="text-secondary fw-bold text-end">
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-shield-half-filled" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M12 3a12 12 0 0 0 8.5 3a12 12 0 0 1 -8.5 15a12 12 0 0 1 -8.5 -15a12 12 0 0 0 8.5 -3" /><path d="M12 3v18" /><path d="M12 11h8.9" /><path d="M12 8h8.9" /><path d="M12 5h3.1" /><path d="M12 17h6.2" /><path d="M12 14h8" /></svg>
                      {#await getTotalVulnerabilites(project.ID)}
                      {:then totalVulnerabilities}
                        {totalVulnerabilities}
                      {/await}
                    </div>
                  </div>
              </td>
              <td>
                <a href={`/project/${project.ID}/view`}>View</a>
              </td>
          </tr>
      {/each}
        </tbody>
      </table>
    </div>
	</div>
</div>


<div hidden>

                    <div class="text-secondary">
                      <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-note" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M13 20l7 -7" /><path d="M13 20v-6a1 1 0 0 1 1 -1h6v-7a2 2 0 0 0 -2 -2h-12a2 2 0 0 0 -2 2v12a2 2 0 0 0 2 2h7" /></svg>
                      Description</div>
</div>

<style>
.selected {
  background-color: rgba(184, 196, 228, 0.05);
  cursor: pointer;
}
</style>