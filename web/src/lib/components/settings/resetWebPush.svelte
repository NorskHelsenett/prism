<script>
  import { Fetch } from '$lib/fetchUtil.js';
  import { toast } from 'svelte-sonner';

  let warningAccepted = $state(false)

  async function resetVAAPI() {
    if (!warningAccepted){
      return
    }

    const response = await Fetch("/api/settings/notification", {method: "DELETE"})

    if (!response.error){
      toast.success("Reset web push settings")
    } else{
      toast.error("Unable to reset web push")
    }
    warningAccepted = false
  }
</script>
<div class="card-body">
  <h2 class="card-title">Reset Web Push</h2>
  <div class="text-secondary">
    <p>Deleting the VAPID keys is a drastic measure that is only recommended as a last resort when the web push notification system is not functioning correctly. This action invalidates the existing keys and requires generating a new set of keys, which must then be implemented into the service worker configuration. It's important to understand that this process will disrupt the service for all current subscribers, as the established trust between the server and the client's browser is broken, necessitating a complete reconfiguration.</p>
    <p>Alongside the reset of VAPID keys, removing all subscribers from the server is another extreme step taken only when necessary. This action erases all stored subscription information, effectively cutting off communication with users who had previously opted-in to receive notifications. As a consequence, users will no longer receive notifications until they re-subscribe to the service. It is critical to communicate this requirement to the user base, guiding them through the re-subscription process to re-establish the delivery of web push notifications.</p>  </div>

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
    <button disabled={!warningAccepted} class="btn btn-warning btn-ghost-warning d-none d-sm-inline-block" onclick={resetVAAPI}>Reset Web Push</button>
  </div>
</div>
