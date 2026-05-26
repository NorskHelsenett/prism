<script>
  import { onMount } from 'svelte';
  import { Fetch } from '$lib/fetchUtil';
  import { apiEndpoint } from '$lib/stores/configStore';
  import { get } from 'svelte/store';

  /** @type {{ data: { token: string } }} */
  let { data } = $props();
  let token = data.token;

  let payload = $state(null);
  let unauthorized = $state(false);
  let notFound = $state(false);

  onMount(async () => {
    const res = await fetch(`${get(apiEndpoint)}/api/share/report/${token}`, {
      credentials: 'include'
    });
    if (res.status === 401) {
      unauthorized = true;
      return;
    }
    if (!res.ok) {
      notFound = true;
      return;
    }
    const body = await res.json();
    // body.data is a JSON-encoded string (datatypes.JSON marshalled as RawMessage).
    let parsed = body.data;
    if (typeof parsed === 'string') {
      try { parsed = JSON.parse(parsed); } catch (_) { parsed = null; }
    }
    payload = { ...body, data: parsed };
  });

  function pdfUrl() {
    return `${get(apiEndpoint)}/api/share/report/${token}/pdf`;
  }
</script>

<div class="container my-4">
  {#if unauthorized}
    <div class="empty">
      <p class="empty-title">Not authorized</p>
      <p class="empty-subtitle text-muted">You need to log in with an invited address to view this report.</p>
      <div class="empty-action">
        <a href="/login" class="btn btn-primary">Sign in</a>
      </div>
    </div>
  {:else if notFound}
    <div class="empty">
      <p class="empty-title">Report not found</p>
      <p class="empty-subtitle text-muted">The link may have expired, or no version has been published yet.</p>
    </div>
  {:else if !payload}
    <div class="text-muted">Loading…</div>
  {:else}
    <div class="d-flex align-items-center mb-4">
      <div>
        <div class="page-pretitle">Disclosed report · v{payload.version}</div>
        <h1 class="page-title">{payload.title}</h1>
        <div class="text-muted">Published {new Date(payload.publishedAt).toLocaleString()} by {payload.publishedBy}</div>
      </div>
      <div class="ms-auto">
        <a class="btn btn-primary" href={pdfUrl()} target="_blank" rel="noopener">Download PDF</a>
      </div>
    </div>

    {#if payload.data}
      <div class="card mb-3">
        <div class="card-body">
          <h3>Executive summary</h3>
          <p style="white-space: pre-wrap">{payload.data.executiveSummary || '—'}</p>
        </div>
      </div>

      <div class="card mb-3">
        <div class="card-header"><h3 class="card-title">Projects in scope</h3></div>
        <ul class="list-group list-group-flush">
          {#each (payload.data.projects || []) as p}
            <li class="list-group-item">{p.name}</li>
          {/each}
        </ul>
      </div>

      <div class="card mb-3">
        <div class="card-header"><h3 class="card-title">Findings ({(payload.data.findings || []).length})</h3></div>
        <div class="card-body">
          {#each (payload.data.findings || []) as f, idx}
            <div class="mb-4">
              <h4>{idx + 1}. {f.title}</h4>
              <div class="text-muted small mb-2">
                Severity: <strong>{f.severity || '—'}</strong> ·
                Status: {f.status || 'Reported'} ·
                Project: {f.projectName || '—'}
              </div>
              <p style="white-space: pre-wrap">{f.summary || ''}</p>
            </div>
          {/each}
        </div>
      </div>
    {/if}
  {/if}
</div>
