<script>
	import { Fetch } from "$lib/fetchUtil";
	import { notification } from "$lib/stores/notificationStore";

  export let settings = {
    ID: 0,
    MFAEnabled: false,
    slack: {
      channelID: "",
      enabled: false,
      workspace: ""
    },
    auditlog: {
      enabled: false
    }
  }

  async function changeMultiFactor() {
    settings.MFAEnabled = !settings.MFAEnabled;

    const response = await Fetch("/api/settings", { method: "POST", body: JSON.stringify(settings) });
    if(!response.error) {
      notification.addAlert({
        type: 'success',
        title: 'Settings',
        message: 'Settings updated successfully'
      });
    } else {
      notification.addAlert({
        type: 'warning',
        title: 'Failure',
        message: 'Failure to update settings'
      });
    }
  }
</script>
<div class="card-body">
  <h3 class="card-title mt-4">One Time Password
    {#if settings.MFAEnabled}
      <span class="badge bg-green-lt">On</span>
    {/if}
  </h3>
<p class="card-subtitle">
  A One Time Password (OTP) is a unique code that is valid for only one login session or transaction and is designed to provide an additional layer of security. When enabled, it supplements traditional authentication methods, such as a password, with a time-sensitive passcode that is sent to a user's trusted device. Since the OTP changes with each session and expires after a short period, this process significantly reduces the risk of unauthorized access arising from compromised user credentials. As a crucial component of multi-factor authentication (MFA) systems, OTP is widely used in verifying user identities for secure online transactions and account logins.
</p>

  <div>
    <label class="form-check form-switch form-switch-lg">
      <input class="form-check-input" type="checkbox" on:change={changeMultiFactor} bind:checked={settings.MFAEnabled}>
      <span class="form-check-label form-check-label-on">OTP is now turned on</span>
      <span class="form-check-label form-check-label-off">OTP is disabled</span>
    </label>
  </div>
</div>