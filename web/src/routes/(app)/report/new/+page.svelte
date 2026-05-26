<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { pageMeta } from '$lib/stores/pageMeta';
  import { Fetch } from '$lib/fetchUtil';

  let projects = $state([]);
  let loadingProjects = $state(true);

  let title = $state('');
  let selectedProjectIds = $state([]);
  let creating = $state(false);
  let error = $state('');

  onMount(async () => {
    pageMeta.set({ pretitle: 'Reports', title: 'New report' });
    const result = await Fetch('/api/project/all');
    projects = Array.isArray(result) ? result : [];
    loadingProjects = false;
  });

  function toggleProject(id) {
    if (selectedProjectIds.includes(id)) {
      selectedProjectIds = selectedProjectIds.filter((p) => p !== id);
    } else {
      selectedProjectIds = [...selectedProjectIds, id];
    }
  }

  async function create() {
    error = '';
    if (!title.trim()) {
      error = 'Give the report a title.';
      return;
    }
    if (selectedProjectIds.length === 0) {
      error = 'Pick at least one project.';
      return;
    }
    creating = true;
    try {
      const created = await Fetch('/api/report', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          title: title.trim(),
          projectIds: selectedProjectIds,
          vulnerabilityIds: [],
          findingOverrides: {}
        })
      });
      if (created && created.ID) {
        await goto(`/report/${created.ID}`);
        return;
      }
      error = "Couldn't create the report — check you have write access to every selected project.";
    } finally {
      creating = false;
    }
  }
</script>

<div class="row g-2 align-items-center">
  <div class="col">
    <div class="page-pretitle">Reports</div>
    <h2 class="page-title">New report</h2>
  </div>
  <div class="col-auto ms-auto">
    <a class="btn btn-link link-secondary" href="/report">Cancel</a>
  </div>
</div>

<div class="row row-cards mt-3">
  <div class="col-lg-8 mx-auto">
    <div class="card">
      <div class="card-body">
        {#if error}
          <div class="alert alert-danger">{error}</div>
        {/if}

        <div class="mb-3">
          <label class="form-label" for="report-title">Title</label>
          <input id="report-title" type="text" class="form-control" bind:value={title} placeholder="Q2 disclosure for Acme" />
        </div>

        <div class="mb-3">
          <label class="form-label">Projects in scope</label>
          {#if loadingProjects}
            <div class="text-muted">Loading…</div>
          {:else if projects.length === 0}
            <div class="text-muted">No projects available.</div>
          {:else}
            <div class="list-group">
              {#each projects as project}
                <label class="list-group-item">
                  <input type="checkbox" class="form-check-input me-2"
                    checked={selectedProjectIds.includes(project.ID)}
                    onchange={() => toggleProject(project.ID)} />
                  {project.ProjectName}
                </label>
              {/each}
            </div>
          {/if}
        </div>
      </div>
      <div class="card-footer">
        <button type="button" class="btn btn-primary ms-auto"
          disabled={creating || !title.trim() || selectedProjectIds.length === 0}
          onclick={create}>
          {creating ? 'Creating…' : 'Create draft'}
        </button>
      </div>
    </div>
  </div>
</div>
