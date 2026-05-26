<script>
  import { onMount } from 'svelte';
  import { pageMeta } from '$lib/stores/pageMeta';
  import { Fetch } from '$lib/fetchUtil';

  let reports = $state([]);
  let loading = $state(true);

  onMount(async () => {
    pageMeta.set({ pretitle: 'Reports', title: 'Vulnerability reports' });
    const result = await Fetch('/api/report');
    reports = Array.isArray(result) ? result : [];
    loading = false;
  });
</script>

<div class="d-print-none">
  <div class="row g-2 align-items-center">
    <div class="col">
      <div class="page-pretitle">Overview</div>
      <h2 class="page-title">Reports</h2>
    </div>
    <div class="col-auto ms-auto d-print-none">
      <a href="/report/new" class="btn btn-primary">
        <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M12 5l0 14"></path><path d="M5 12l14 0"></path></svg>
        New report
      </a>
    </div>
  </div>
</div>

<div class="row row-cards mt-3">
  <div class="col-12">
    <div class="card">
      <div class="table-responsive">
        <table class="table card-table table-vcenter">
          <thead>
            <tr>
              <th>Title</th>
              <th>Owner</th>
              <th>Projects</th>
              <th>Latest version</th>
              <th>Updated</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {#if loading}
              <tr><td colspan="6" class="text-muted">Loading…</td></tr>
            {:else if reports.length === 0}
              <tr><td colspan="6" class="text-muted">No reports yet.</td></tr>
            {:else}
              {#each reports as r}
                <tr>
                  <td><a href={`/report/${r.ID}`}>{r.Title}</a></td>
                  <td>{r.OwnerEmail}</td>
                  <td>{(r.projectIds || []).length}</td>
                  <td>
                    {#if r.latestPublishedVersionId}
                      <span class="badge bg-green-lt">published</span>
                    {:else}
                      <span class="badge bg-yellow-lt">draft</span>
                    {/if}
                  </td>
                  <td>{new Date(r.UpdatedAt).toLocaleString()}</td>
                  <td class="text-end">
                    <a class="btn btn-sm" href={`/report/${r.ID}`}>Open</a>
                  </td>
                </tr>
              {/each}
            {/if}
          </tbody>
        </table>
      </div>
    </div>
  </div>
</div>
