<script>
	import { Fetch } from "$lib/fetchUtil";
	import { notification } from "$lib/stores/notificationStore";

  async function cleanup(){
    const response = await Fetch("/api/settings/cleanup")
    if(!response.error) {
      notification.addAlert({
        type: 'success',
        title: 'Settings',
        message: 'Database optimized successfully'
      });
    } else {
      notification.addAlert({
        type: 'warning',
        title: 'Failure',
        message: 'Unable to perform task'
      });
    }
    warningAccepted = false
  }

  let warningAccepted = false
  export let settings = {}

  export function formatSize(size) {
    const sizeInMB = size / 1024 / 1024;
    if (sizeInMB < 1024) {
        return `${sizeInMB.toFixed(2)} MB`;
    } else {
        const sizeInGB = sizeInMB / 1024;
        return `${sizeInGB.toFixed(2)} GB`;
    }
}
</script>

<div class="card-body">
  <h2 class="card-title">Database Health</h2>

  <div class="mb-3 list-inline list-inline-dots mb-0 text-secondary d-sm-block d-none">
    <div class="list-inline-item" title="Database file size">
      <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-database" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M12 6m-8 0a8 3 0 1 0 16 0a8 3 0 1 0 -16 0" /><path d="M4 6v6a8 3 0 0 0 16 0v-6" /><path d="M4 12v6a8 3 0 0 0 16 0v-6" /></svg>
      {formatSize(settings?.metrics?.DatabaseSize)}
    </div>
    <div class="list-inline-item" title="Memory usage">
      <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-artboard" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M8 8m0 1a1 1 0 0 1 1 -1h6a1 1 0 0 1 1 1v6a1 1 0 0 1 -1 1h-6a1 1 0 0 1 -1 -1z" /><path d="M3 8l1 0" /><path d="M3 16l1 0" /><path d="M8 3l0 1" /><path d="M16 3l0 1" /><path d="M20 8l1 0" /><path d="M20 16l1 0" /><path d="M8 20l0 1" /><path d="M16 20l0 1" /></svg>
      {formatSize(settings?.metrics?.Memory)}
    </div>
  </div>

<p class="text-secondary">
    Initiate a <strong>Database Cleanup</strong> process. This will <strong>permanently delete</strong> all soft-deleted records and perform optimization tasks on the database. <strong>It is irreversible!</strong>
  </p>
  <p class="text-secondary">
    <strong class="badge bg-orange-lt">Warning:</strong> This action may lead to <strong>irreversible data loss</strong> for any soft-deleted records. It is highly recommended to <strong>export and backup your database</strong> before proceeding.
    During the cleanup process, the database may become temporarily unavailable, and this might affect active sessions or operations.
  </p>
  <span class="badge bg-orange-lt mb-3">
    Ensure you have a recent backup. The cleanup operation cannot be undone.
  </span>
  <div class="text-secondary">This task will execute the following</div>
  <ul class="text-secondary">
    <li>Permanently delete all soft-deleted <strong>vulnerabilities</strong></li>
    <li>Permanently delete all soft-deleted <strong>projects</strong></li>
    <li>Permanently delete all soft-deleted <strong>user data</strong></li>
    <li><strong>VACUUM</strong> the database</li>
    <li><strong>REINDEX</strong> the database</li>
    <li><strong>ANALYZE</strong> the database</li>
  </ul>

  <div class="btn-list justify-content-end">
    <label class="form-check me-3">
      <input class="form-check-input" type="checkbox" bind:checked={warningAccepted}>
      <span class="form-check-label">
        Confirm
      </span>
      <span class="form-check-description">
        I accept the risk of my action.
      </span>
    </label>
    <button disabled={!warningAccepted} class="btn btn-warning btn-ghost-warning d-none d-sm-inline-block" on:click={cleanup}>Perform Database Cleanup</button>
  </div>
</div>

<style>
  .btn:disabled {
    border-color: rgba(0,0,0,0);
  }
</style>