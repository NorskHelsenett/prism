<script>
	import { Fetch } from "$lib/fetchUtil";
	import DebouncedInput from "../DebouncedInput.svelte";
	import { toast } from "svelte-sonner";

  let persisting = false;
  export let settings = {
    ID: 0,
    slack: {
      channelID: "",
      enabled: false,
      workspace: ""
    },
    auditlog: {
      enabled: false
    }
  }

  function debounce(func, wait) {
      let timeout;
      return function(...args) {
          const later = () => {
              clearTimeout(timeout);
              func(...args);
          };
          clearTimeout(timeout);
          timeout = setTimeout(later, wait);
      };
  }

  async function persistChannelID() {
      try {
          persisting = true;
          // Replace with actual API call
          await updateChannelIDAPI();
      } catch (error) {
          console.error("Failed to persist channel ID:", error);
      } finally {
          persisting = false;
      }
  }

  // Debounce the persistChannelID function
  const debouncePersist = debounce(persistChannelID, 500);

  function handleBlur() {
      debouncePersist();
  }

  // Mock API function with a timeout to simulate delay
  async function updateChannelIDAPI() {
    const response = await Fetch("/api/settings", { method: "POST", body: JSON.stringify(settings) });
      if(!response.error) {
        toast.success('Settings updated successfully');
			} else {
        toast.error('Failed to update settings');
      }
  }

  async function changeAuditlog() {
    settings.auditlog.enabled = !settings.auditlog.enabled;

    const response = await Fetch("/api/settings", { method: "POST", body: JSON.stringify(settings) });
        if(!response.error) {
          toast.success('Settings updated successfully');
    } else {
      toast.error('Failure to update settings');
    }
  }

  async function persistSlackStatus() {
    settings.slack.enabled = !settings.slack.enabled; // Toggle the value (you can modify this logic as needed)

    const response = await Fetch("/api/settings", { method: "POST", body: JSON.stringify(settings) });
    if(!response.error) {
      toast.success('Settings updated successfully');
    } else {
      toast.error('Failure to update settings');
    }
  }

  function handleChannelIDChange(newVal) {
    settings.slack.channelID = newVal.detail;
    debouncePersist();
  }

  function handleWorkspaceChange(newVal) {
    settings.slack.workspace = newVal.detail;
    debouncePersist();
  }
</script>
<div class="card-body">
  <h3 class="card-title mt-4">Slack
    {#if settings.slack.enabled}
      <span class="badge bg-green-lt">On</span>
    {/if}
  </h3>
  <p class="card-subtitle"> With Slack notification activated, you'll receive instant Slack notifications for each new vulnerability detected. This integration ensures you stay informed in real-time, enabling quicker responses and seamless collaboration within your team.
  </p>
  <div>
    <label class="form-check form-switch form-switch-lg">
      <input class="form-check-input" type="checkbox" on:change={persistSlackStatus} bind:checked={settings.slack.enabled}>
      <span class="form-check-label form-check-label-on">Slack integration is now turned on</span>
      <span class="form-check-label form-check-label-off">Slack integration disabled</span>
    </label>
  </div>

  <div class="row">
    <div class="col-6">
      <label class="form-label text-secondary">Default slack channelID if not set in project.</label>

      <div class="form-floating mb-3">
        <DebouncedInput
            id="channel-id"
            placeholder="Channel ID"
            bind:value={settings.slack.channelID}
            on:change={handleChannelIDChange}
            persisting={persisting} />
          </div>
    </div>
    <div class="col-6">

      <label class="form-label text-secondary">The workspace where your bot is installed.</label>
      <div class="form-floating mb-3">

    <DebouncedInput
        id="workspace"
        placeholder="Workspace"
        bind:value={settings.slack.workspace}
        on:change={handleWorkspaceChange}
        persisting={persisting} />
      </div>
    </div>
  </div>
</div>

<div class="card-body">
  <h3 class="card-title mt-4">Audit logging
    {#if settings.auditlog.enabled}
      <span class="badge bg-green-lt">On</span>
    {/if}
  </h3>
  <p class="card-subtitle">
      Enabling audit logging captures detailed records of all user requests, enhancing security and aiding in compliance. This feature is invaluable for security monitoring, regulatory adherence, and troubleshooting. However, it's important to be aware of the potential impacts on system performance and data management. Audit logs can increase system load and generate significant data volumes, necessitating efficient storage and management solutions. Additionally, careful consideration must be given to privacy and data protection laws, as these logs often contain sensitive user information.
  </p>

  <div>
    <label class="form-check form-switch form-switch-lg">
      <input class="form-check-input" type="checkbox" on:change={changeAuditlog} bind:checked={settings.auditlog.enabled}>
      <span class="form-check-label form-check-label-on">Audit log is now turned on</span>
      <span class="form-check-label form-check-label-off">Audit log is disabled</span>
    </label>
  </div>
</div>
