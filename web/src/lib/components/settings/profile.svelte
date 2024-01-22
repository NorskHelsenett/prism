<script>
  import { onMount } from 'svelte';
  import { Fetch } from '$lib/fetchUtil.js';
	import DebouncedInput from '../DebouncedInput.svelte';
	import { goto } from '$app/navigation';

  let user;
  let persisting = false
  let sessions = []

  onMount(async () => {
    user = await Fetch(`/api/user`);
    sessions = await Fetch("/api/user/session/all")
  });

  async function handleTitleChange(newVal) {
    persisting = true
    user.title = newVal.detail
    const response = await Fetch("/api/user", {method: "PUT", body: JSON.stringify(user)})
    persisting = false
  }

  async function resetMFA() {
    await Fetch("/api/session/otp/reset")
    goto("/")
  }

  async function endSession(id) {
    const response = await Fetch(`/api/user/session/${id}`, {method: "DELETE"})
    if (!response?.error){
      sessions = await Fetch("/api/user/session/all")
    }
  }

  function formatDate(dateString) {
    const options = { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hour12: false };
    return new Date(dateString).toLocaleDateString('en-US', options).replace(/\//g, '.').replace(',', '');
  }
function formatDateText(expiresAt) {
    const now = new Date();
    const expiry = new Date(expiresAt);
    let difference = expiry - now;

    const isInPast = difference < 0;
    difference = Math.abs(difference);

    const days = Math.floor(difference / (1000 * 60 * 60 * 24));
    difference -= days * (1000 * 60 * 60 * 24);

    const hours = Math.floor(difference / (1000 * 60 * 60));
    difference -= hours * (1000 * 60 * 60);

    const minutes = Math.floor(difference / (1000 * 60));

    let formattedDate = "";
    if (days > 0) formattedDate += `${days} day${days > 1 ? 's' : ''} `;
    if (hours > 0) formattedDate += `${hours} hour${hours > 1 ? 's' : ''} `;
    if (minutes > 0) formattedDate += `${minutes} minute${minutes > 1 ? 's' : ''}`;

    // Add a prefix based on whether the date is in the past or future
    if (isInPast) {
        formattedDate = formattedDate ? `${formattedDate} ago` : "Already expired";
    } else {
        formattedDate = formattedDate ? `In ${formattedDate}` : "Now";
    }

    return formattedDate.trim();
}

</script>

{#if user}
<div class="card-body">
  <h2 class="mb-4">My Account</h2>
  <h3 class="card-title">Profile Details</h3>
  <div class="row align-items-center">
    <div class="col-auto">
      <span class="avatar avatar-xl" style="background-image: url({user.Picture})"></span>
    </div>
    <div class="col m-2">
      <strong class="row">Name</strong>
      <div class="row text-secondary">{user.Name}</div>
      <strong class="row">Email</strong>
      <a href="mailto:{user.Email}" class="row">{user.Email}</a>
    </div>
  </div>
  <h3 class="card-title mt-4">Profile</h3>
  <div class="row g-3">
    <div class="col-md">
      <DebouncedInput
          id="userTitle"
          placeholder="Title"
          bind:value={user.Title}
          on:change={handleTitleChange}
          persisting={persisting} />
    </div>
  </div>

    <h3 class="card-title mt-4">Active sessions</h3>
  <p class="card-subtitle">Below is a comprehensive list of all your active sessions. You have the option to end any of these sessions at your discretion. Please be aware that if you choose to terminate the session you are currently using, you will be immediately logged out of your account. Exercise caution when selecting the session to end, especially if it is your current one.</p>

  <div class="table-responsive">
    <table class="table table-vcenter card-table">
      <thead>
        <tr>
          <th>Created At</th>
          <th>Expiring At</th>
          <th class="w-1"></th>
        </tr>
      </thead>
      <tbody>
      {#each sessions as session}
        <tr>
          <td class="text-secondary" title="{formatDateText(session.CreatedAt)}">{formatDate(session.CreatedAt)}</td>
          <td class="text-secondary" title="{formatDateText(session.ExpiresAt)}">
            {formatDate(session.ExpiresAt)}
          </td>
          <td>
            <a href="#" class="" on:click={endSession(session.SessionID)}>Disconnect</a>
          </td>
        </tr>
      {/each}
      </tbody>
    </table>
  </div>

  <h3 class="card-title mt-4">Multi factor authentication <span class="badge bg-green-lt">On</span></h3>
  <p class="card-subtitle">Your account is currently safeguarded by a One-Time Password (OTP) mechanism, an essential part of our multi-factor authentication process. This additional layer of security ensures that only you have access to your account, even if someone knows your password.</p>
  <p class="text-secondary">Remember, this will log you out, and you will be required to go through the OTP generation flow before being able to log in again.</p>
  <div class="btn-list justify-content-start">
    <button class="btn btn-ghost-warning d-none d-sm-inline-block" on:click={resetMFA}>Reset 2MFA</button>
  </div>

</div>
{/if}