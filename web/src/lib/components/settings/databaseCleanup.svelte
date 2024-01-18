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
</script>

<div class="card-body">
  <h2 class="card-title">Database Health</h2>
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
    <button disabled={!warningAccepted} class="btn btn-warning d-none d-sm-inline-block" on:click={cleanup}>Perform Database Cleanup</button>
  </div>
</div>

<style>
  .btn:disabled {
    border-color: rgba(0,0,0,0);
  }
</style>