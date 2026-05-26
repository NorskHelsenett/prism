<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { pageMeta } from '$lib/stores/pageMeta';
  import { Fetch } from '$lib/fetchUtil';
  import RichTextEditor from '$lib/components/editor/RichTextEditor.svelte';
  import DeleteModal from '$lib/components/DeleteModal.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import Markdown from '$lib/components/Markdown.svelte';
  import SeverityBucket from '$lib/components/severityBucket.svelte';
  import { apiEndpoint } from '$lib/stores/configStore';
  import { get } from 'svelte/store';

  let showDeleteModal = $state(false);
  let showPreview = $state(false);
  /** @type {{ raw: any, data: Record<string, any> } | null} */
  let previewVuln = $state(null);

  function vulnData(v) {
    if (!v) return {};
    if (typeof v.Vulnerability === 'string') {
      try { return JSON.parse(v.Vulnerability); } catch (_) { return {}; }
    }
    return v.Vulnerability || {};
  }

  function statusBadgeClass(status) {
    switch (status) {
      case 'Reported': return 'bg-azure-lt';
      case 'Validated': return 'bg-pink-lt';
      case 'In Progress': return 'bg-orange-lt';
      case 'Rejected':
      case 'Resolved': return 'bg-green-lt';
      default: return 'bg-secondary-lt';
    }
  }

  function openPreview(v) {
    previewVuln = { raw: v, data: vulnData(v) };
    showPreview = true;
  }

  function closePreview() {
    showPreview = false;
    previewVuln = null;
  }

  let report = $state(null);
  let projects = $state([]);
  let vulnerabilities = $state([]);
  let versions = $state([]);
  let saving = $state(false);
  let publishing = $state(false);

  let executiveSummary = $state('');
  let title = $state('');
  let invitedEmailsInput = $state('');

  let reportId = $derived($page.params.id);

  let selectedVulnIds = $derived(report?.vulnerabilityIds || []);
  let selectedProjectIds = $derived(report?.projectIds || []);

  onMount(async () => {
    pageMeta.set({ pretitle: 'Report', title: 'Edit report' });
    await load();
  });

  async function load() {
    const r = await Fetch(`/api/report/${reportId}`);
    if (!r || r.error) return;
    report = r;
    title = r.Title || '';
    executiveSummary = r.ExecutiveSummary || '';
    invitedEmailsInput = (r.invitedEmails || []).join(', ');
    const [allProjects, vulns, vs] = await Promise.all([
      Fetch('/api/project/all'),
      fetchVulnsForProjects(r.projectIds || []),
      Fetch(`/api/report/${reportId}/versions`)
    ]);
    projects = Array.isArray(allProjects) ? allProjects : [];
    vulnerabilities = vulns;
    versions = Array.isArray(vs) ? vs : [];
  }

  async function fetchVulnsForProjects(projectIds) {
    const combined = [];
    for (const pid of projectIds) {
      const result = await Fetch(`/api/project/${pid}/vulnerabilities`);
      if (Array.isArray(result)) {
        for (const v of result) combined.push(v);
      }
    }
    return combined;
  }

  async function refreshVulnPool() {
    vulnerabilities = await fetchVulnsForProjects(report?.projectIds || []);
  }

  function toggleProject(id) {
    const list = new Set(report.projectIds || []);
    if (list.has(id)) list.delete(id); else list.add(id);
    report = { ...report, projectIds: [...list] };
  }

  function toggleVuln(id) {
    const list = new Set(report.vulnerabilityIds || []);
    if (list.has(id)) list.delete(id); else list.add(id);
    report = { ...report, vulnerabilityIds: [...list] };
  }

  async function save() {
    saving = true;
    try {
      const updated = await Fetch(`/api/report/${reportId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          title,
          executiveSummary,
          projectIds: report.projectIds,
          vulnerabilityIds: report.vulnerabilityIds,
          findingOverrides: report.findingOverrides || {}
        })
      });
      if (updated && !updated.error) {
        report = updated;
        await refreshVulnPool();
      }
    } finally {
      saving = false;
    }
  }

  async function saveShare() {
    const emails = invitedEmailsInput
      .split(/[,\n]/)
      .map((e) => e.trim())
      .filter(Boolean);
    const result = await Fetch(`/api/report/${reportId}/share`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ invitedEmails: emails })
    });
    if (result && !result.error) {
      report = { ...report, invitedEmails: result.invitedEmails };
    }
  }

  async function publish() {
    publishing = true;
    try {
      await save();
      const v = await Fetch(`/api/report/${reportId}/publish`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: '{}'
      });
      if (v && !v.error) {
        const vs = await Fetch(`/api/report/${reportId}/versions`);
        versions = Array.isArray(vs) ? vs : [];
        await load();
      }
    } finally {
      publishing = false;
    }
  }

  async function remove() {
    const result = await Fetch(`/api/report/${reportId}`, { method: 'DELETE' });
    if (result && !result.error) {
      await goto('/report');
    }
  }

  function pdfUrl(version) {
    return `${get(apiEndpoint)}/api/report/${reportId}/versions/${version}/pdf`;
  }

  function shareUrl() {
    if (!report?.ShareToken) return '';
    return `${window.location.origin}/s/r/${report.ShareToken}`;
  }
</script>

<DeleteModal bind:showDeleteModal onDelete={remove} deleteButtonText="Yes, delete it" text="Delete this report and all its published versions. This cannot be undone." />

<Modal bind:showModal={showPreview} fullscreen={true} on:close={closePreview}>
  {#snippet title()}
    <div class="d-flex align-items-center gap-2 flex-wrap">
      <h3 class="modal-title mb-0">{previewVuln?.data.title || 'Untitled finding'}</h3>
      {#if previewVuln?.data.criticality || previewVuln?.data.severity}
        <SeverityBucket severity={previewVuln?.data.criticality || previewVuln?.data.severity || ''} />
      {/if}
      {#if previewVuln?.raw.Status}
        <span class={'badge ' + statusBadgeClass(previewVuln.raw.Status)}>{previewVuln.raw.Status}</span>
      {/if}
    </div>
  {/snippet}

  {#if previewVuln}
    {@const d = previewVuln.data}
    <div class="modal-body">
      <div class="row g-3 mb-3">
        <div class="col-md-3">
          <div class="text-muted small">Category</div>
          <div>{d.category || '—'}</div>
        </div>
        <div class="col-md-3">
          <div class="text-muted small">Endpoint</div>
          <div class="text-break">{d.endpoint || '—'}</div>
        </div>
        <div class="col-md-2">
          <div class="text-muted small">Ease of exploitation</div>
          <div>{d.easeOfExploitation || '—'}</div>
        </div>
        <div class="col-md-2">
          <div class="text-muted small">Impact</div>
          <div>{d.impact || '—'}</div>
        </div>
        <div class="col-md-2">
          <div class="text-muted small">Date</div>
          <div>{d.date || '—'}</div>
        </div>
      </div>

      <div class="row g-3 mb-3">
        <div class="col-md-4">
          <div class="text-muted small">Assigned to</div>
          <div>{d.assignedTo || '—'}</div>
        </div>
        <div class="col-md-4">
          <div class="text-muted small">Public-facing</div>
          <div>{d.isPublicFacing ? 'Yes' : 'No'}</div>
        </div>
        <div class="col-md-4">
          <div class="text-muted small">Visibility</div>
          <div>{d.visibility || '—'}</div>
        </div>
      </div>

      <div class="mb-4">
        <h4>Evidence</h4>
        {#if d.evidence}
          <Markdown markdown={d.evidence} />
        {:else}
          <div class="text-muted">No evidence captured.</div>
        {/if}
      </div>

      <div>
        <h4>Remediation</h4>
        {#if d.remediation}
          <Markdown markdown={d.remediation} />
        {:else}
          <div class="text-muted">No remediation captured.</div>
        {/if}
      </div>
    </div>
    <div class="modal-footer">
      <a class="btn btn-link link-secondary me-auto" href={`/vulnerability/${previewVuln.raw.ID}/view`} target="_blank" rel="noopener">Open full view</a>
      <button type="button" class="btn" onclick={closePreview}>Close</button>
    </div>
  {/if}
</Modal>

{#if !report}
  <div class="text-muted">Loading…</div>
{:else}
  <div class="row g-2 align-items-center">
    <div class="col">
      <div class="page-pretitle">Report</div>
      <h2 class="page-title">{title || 'Untitled report'}</h2>
    </div>
    <div class="col-auto ms-auto">
      <button class="btn btn-outline-danger me-2" onclick={() => (showDeleteModal = true)}>Delete</button>
      <button class="btn btn-outline-primary me-2" disabled={saving} onclick={save}>
        {saving ? 'Saving…' : 'Save draft'}
      </button>
      <button class="btn btn-primary" disabled={publishing} onclick={publish}>
        {publishing ? 'Publishing…' : 'Publish new version'}
      </button>
    </div>
  </div>

  <div class="row row-cards mt-3">
    <div class="col-lg-8">
      <div class="card mb-3">
        <div class="card-body">
          <div class="mb-3">
            <label class="form-label" for="r-title">Title</label>
            <input id="r-title" class="form-control" bind:value={title} />
          </div>
          <div class="mb-3">
            <label class="form-label">Executive summary</label>
            <RichTextEditor bind:value={executiveSummary} placeholder="What's the headline for non-technical readers?" minHeight="240px" />
          </div>
        </div>
      </div>

      <div class="card mb-3">
        <div class="card-header"><h3 class="card-title">Findings</h3></div>
        {#if vulnerabilities.length === 0}
          <div class="card-body text-muted">No vulnerabilities available for the selected projects.</div>
        {:else}
          <div class="table-responsive">
            <table class="table card-table table-vcenter">
              <thead>
                <tr>
                  <th style="width: 2.5rem"></th>
                  <th>Title</th>
                  <th style="width: 11rem">Criticality</th>
                  <th>Category</th>
                  <th style="width: 8rem">Status</th>
                  <th>Endpoint</th>
                  <th class="text-end" style="width: 7rem"></th>
                </tr>
              </thead>
              <tbody>
                {#each vulnerabilities as v}
                  {@const data = vulnData(v)}
                  <tr>
                    <td>
                      <input type="checkbox" class="form-check-input"
                        checked={selectedVulnIds.includes(v.ID)}
                        onchange={() => toggleVuln(v.ID)}
                        aria-label="Include in report" />
                    </td>
                    <td>
                      <button type="button" class="btn btn-link p-0 text-start" onclick={() => openPreview(v)}>
                        {data.title || 'Untitled finding'}
                      </button>
                    </td>
                    <td><SeverityBucket severity={data.criticality || data.severity || ''} /></td>
                    <td class="text-muted">{data.category || '—'}</td>
                    <td><span class={'badge ' + statusBadgeClass(v.Status)}>{v.Status}</span></td>
                    <td class="text-muted text-truncate" style="max-width: 18rem">{data.endpoint || '—'}</td>
                    <td class="text-end">
                      <button type="button" class="btn btn-sm" onclick={() => openPreview(v)}>Preview</button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </div>

      <div class="card">
        <div class="card-header"><h3 class="card-title">Published versions</h3></div>
        <div class="table-responsive">
          <table class="table card-table">
            <thead><tr><th>Version</th><th>Published</th><th>By</th><th></th></tr></thead>
            <tbody>
              {#each versions as v}
                <tr>
                  <td>v{v.version}</td>
                  <td>{new Date(v.publishedAt).toLocaleString()}</td>
                  <td>{v.publishedBy}</td>
                  <td class="text-end"><a class="btn btn-sm" href={pdfUrl(v.version)} target="_blank" rel="noopener">Download PDF</a></td>
                </tr>
              {/each}
              {#if versions.length === 0}
                <tr><td colspan="4" class="text-muted">Not published yet.</td></tr>
              {/if}
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="col-lg-4">
      <div class="card mb-3">
        <div class="card-header"><h3 class="card-title">Projects in scope</h3></div>
        <div class="card-body">
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
          <small class="text-muted d-block mt-2">Save the draft to refresh the finding list for these projects.</small>
        </div>
      </div>

      <div class="card mb-3">
        <div class="card-header"><h3 class="card-title">Share</h3></div>
        <div class="card-body">
          <div class="mb-3">
            <label class="form-label">Share URL</label>
            <input class="form-control" readonly value={shareUrl()} />
            <small class="text-muted">Send to invited readers; their email must match.</small>
          </div>
          <div class="mb-3">
            <label class="form-label" for="invited">Invited emails</label>
            <textarea id="invited" class="form-control" rows="4" bind:value={invitedEmailsInput} placeholder="alice@acme.com, bob@acme.com"></textarea>
          </div>
          <button class="btn btn-primary w-100" onclick={saveShare}>Update invites</button>
        </div>
      </div>
    </div>
  </div>
{/if}
