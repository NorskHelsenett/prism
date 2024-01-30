<script>
  import { onMount } from 'svelte';
  import { marked } from 'marked';
  import DOMPurify from 'dompurify'
	import formatNumber from '$lib/formatNumber';

  let buildInfo = {
    "user": "",
    "branch": "",
    "date": "",
    "digest": "",
    "totalLines": 0
  }
  let changelog = '';

  onMount(async () => {
    const response = await fetch(`/.well-known/buildinfo.json`)

    if (response.ok){
      buildInfo = await response.json()
    }

    const mdResponse = await fetch('/.well-known/CHANGELOG.md');
    const md = await mdResponse.text();
    changelog = DOMPurify.sanitize(marked.parse(md));
  });
</script>
<div class="card-body">
  <h2 class="mb-4">About</h2>

  <div class="datagrid">
    <div class="datagrid-item">
      <div class="datagrid-title">Build date</div>
      <div class="datagrid-content">
        <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-calendar" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M4 7a2 2 0 0 1 2 -2h12a2 2 0 0 1 2 2v12a2 2 0 0 1 -2 2h-12a2 2 0 0 1 -2 -2v-12z" /><path d="M16 3v4" /><path d="M8 3v4" /><path d="M4 11h16" /><path d="M11 15h1" /><path d="M12 15v3" /></svg>
        {buildInfo.date}</div>
    </div>
    <div class="datagrid-item">
      <div class="datagrid-title">Git author</div>
      <div class="datagrid-content">
      <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-user" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M8 7a4 4 0 1 0 8 0a4 4 0 0 0 -8 0" /><path d="M6 21v-2a4 4 0 0 1 4 -4h4a4 4 0 0 1 4 4v2" /></svg>
        {buildInfo.user}
      </div>
    </div>
    <div class="datagrid-item">
      <div class="datagrid-title">Git commit digest</div>
      <div class="datagrid-content">
        <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-hash" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M5 9l14 0" /><path d="M5 15l14 0" /><path d="M11 4l-4 16" /><path d="M17 4l-4 16" /></svg>
        {buildInfo.digest.slice(0,8)}</div>
    </div>
    <div class="datagrid-item">
      <div class="datagrid-title">Git branch</div>
      <div class="datagrid-content">
      <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-git-branch" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M7 18m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0" /><path d="M7 6m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0" /><path d="M17 6m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0" /><path d="M7 8l0 8" /><path d="M9 18h6a2 2 0 0 0 2 -2v-5" /><path d="M14 14l3 -3l3 3" /></svg>
        {buildInfo.branch}
      </div>
    </div>
    <div class="datagrid-item">
      <div class="datagrid-title">Code Lines</div>
      <div class="datagrid-content">
        <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-file-description" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M14 3v4a1 1 0 0 0 1 1h4" /><path d="M17 21h-10a2 2 0 0 1 -2 -2v-14a2 2 0 0 1 2 -2h7l5 5v11a2 2 0 0 1 -2 2z" /><path d="M9 17h6" /><path d="M9 13h6" /></svg>
        {formatNumber(buildInfo.totalLines)}
      </div>
    </div>
  </div>
  <hr />
  <div>
    {@html changelog}
  </div>
</div>