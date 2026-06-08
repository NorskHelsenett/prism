<script>
  import { preventDefault, stopPropagation } from 'svelte/legacy';

	import { Fetch } from "$lib/fetchUtil";
	import { onMount } from "svelte";
	import Dropdown from "../Dropdown.svelte";
  import DeleteModal from '$lib/components/DeleteModal.svelte';
  import InfoModal from '$lib/components/modals/InfoModal.svelte';
	import { toast } from "svelte-sonner";
	import Icon from "../Icon.svelte";
	import Avatar from "../Avatar.svelte";
	import Avatarlist from "../calendar/Avatarlist.svelte";
	import Team from "./teams/Team.svelte";
	import UserStats from "./UserStats.svelte";

  let teamComponent = $state();
  let users = $state([])
  let roles = $state([])
  let showDeleteModal = $state(false);
  let showInfoModal = $state(false);
  let showToggleModal = $state(false);
  let userFilter = $state("active");

  onMount(async () => {
    const fetchedUsers = await Fetch("/api/settings/users/all")
    showDropdown = fetchedUsers.map(() => false)
    users = fetchedUsers
    roles = await Fetch("/api/settings/roles-list")
  });

  let filteredUsers = $derived(users.filter(user => {
    const isActive = user.active === undefined || user.active === null || user.active;
    if (userFilter === "active") return isActive;
    if (userFilter === "disabled") return !isActive;
    return true;
  }));

  let desc = false

  function orderBy(n) {
    return (event) => {
    switch (n) {
      case "name":
        users = users.sort((a, b) => a.name.localeCompare(b.name));
        break;
      case "role":
        users = users.sort((a, b) => a.role.localeCompare(b.role));
        break;
      default:
        users = users.sort((a, b) => a.name.localeCompare(b.name));
        break;
      }
      desc = !desc
      if (desc) users = users.reverse();
  }}

  function isRole(userRole, role){
    return userRole == role
  }
  function userIsAdmin(userRole) {
    return userRole == "admin"
  }

  function updateUser(user){
    return async function(event) {
      const newRole = event.target.value;
      user.role = newRole

      const response = await Fetch('/api/settings/profile', {
        method: 'PUT',
        body: JSON.stringify(user)
      });

      if(!response.error) {
        toast.success('User updated');
      } else {
        toast.error('Unable to perform task');
      }
    };
  }

function formatDate(dateString) {
  const options = { day: '2-digit', month: 'short', year: 'numeric' };
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', options);
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

  let showDropdown = $state([false])
  let userToResetMFA = $state(null)
  let mfaStatus = $state(null)
  let loadingMfaStatus = $state(false)

  async function resetMFAok() {
    const response = await Fetch(`/api/settings/session/otp/reset/${userToResetMFA.email}`, {method: "DELETE"})
    userToResetMFA = null
    mfaStatus = null
    showInfoModal = false

    if(!response.error) {
      toast.success('Successfully reset MFA');
    } else {
      toast.error('Unable to reset MFA');
    }
  }

  async function resetMFA(user){
    userToResetMFA = user
    mfaStatus = null
    loadingMfaStatus = true
    showInfoModal = true
    showDropdown = users.map(() => false)

    const status = await Fetch(`/api/settings/session/mfa-status/${user.email}`);
    if (status && !status.error) {
      mfaStatus = status;
    }
    loadingMfaStatus = false
  }


  let userMarkedForToggle = $state(null);
  let toggleDialogText = "";
  let toggleDialogButton = $state("");

  function toggleUserActive(user) {
    userMarkedForToggle = user;
    const isActive = user.active !== false;
    toggleDialogText = isActive
      ? `Deactivate ${user.name}? The user will no longer be able to log in.`
      : `Activate ${user.name}? The user will be able to log in again.`;
    toggleDialogButton = isActive ? "Deactivate" : "Activate";
    showToggleModal = true;
    showDropdown = users.map(() => false);
  }

  async function toggleUserActivePrompted() {
    const response = await Fetch(`/api/settings/user/${userMarkedForToggle.ID}/active`, { method: "PATCH" });
    if (!response.error) {
      const idx = users.findIndex(u => u.ID === userMarkedForToggle.ID);
      if (idx !== -1) {
        users[idx].active = response.active;
        users = users;
      }
      toast.success(response.active ? 'User activated' : 'User deactivated');
      teamComponent?.fetchTeams();
    } else {
      toast.error('Unable to toggle user status');
    }
    userMarkedForToggle = null;
    showToggleModal = false;
  }

  let userMarkedForDeletion = $state(null)
  let deleteDialogText = ""
  let deleteDialogButton= $state("")

  function deleteUser(user){
    userMarkedForDeletion = user

    deleteDialogText= `Delete ${userMarkedForDeletion?.name}?. This action is irreversible.`
    deleteDialogButton= "Delete the user!"

    showDeleteModal = true
  }

  async function deleteUserPrompted(){
    const response = await Fetch(`/api/settings/user/${userMarkedForDeletion.ID}`, {method: "DELETE"})
    users = users.filter(user => user.ID !== userMarkedForDeletion.ID);
    userMarkedForDeletion = null
    showDeleteModal = false

    if(!response.error) {
      toast.success('Successfully deleted user');
      teamComponent?.fetchTeams();
    } else {
      toast.error('Unable to delete user');
    }
  }
</script>
<!-- <div class="row">
  <h3 class="card-title mt-4">Team</h3>
  <p class="card-subtitle"> With Slack notification activated, you'll receive instant Slack notifications for each new vulnerability detected. This integration ensures you stay informed in real-time, enabling quicker responses and seamless collaboration within your team.
  </p>
</div> -->


<Team bind:this={teamComponent} />
<!-- <UserStats /> -->

<div class="ms-3 me-3 mt-3 d-flex justify-content-center">
  <div class="btn-group" role="group">
    <input type="radio" class="btn-check" name="userFilter" id="radio-filter-active" value="active" bind:group={userFilter}>
    <label for="radio-filter-active" class="btn">Active</label>
    <input type="radio" class="btn-check" name="userFilter" id="radio-filter-disabled" value="disabled" bind:group={userFilter}>
    <label for="radio-filter-disabled" class="btn">Disabled</label>
    <input type="radio" class="btn-check" name="userFilter" id="radio-filter-all" value="all" bind:group={userFilter}>
    <label for="radio-filter-all" class="btn">All</label>
  </div>
</div>

<div class="card me-3 mb-3 mt-3 ml-1">
  <div class="table-responsive">
    <table class="table table-vcenter table-mobile-md card-table">
      <thead>
        <tr>
          <th onclick={orderBy("name", true)}><button class="table-sort" data-sort="sort-name">name</button></th>
          <th>Created</th>
          <th>Last seen</th>
          <th><button onclick={orderBy("role", true)} class="table-sort" data-sort="sort-name">Role</button></th>
          <th class="w-1"></th>
        </tr>
      </thead>
      <tbody>
      {#each filteredUsers as user, index}
        <tr>
          <td data-label="name">
            <div class="d-flex py-1 align-items-center">
              <span class="avatar me-2" style="background-image: url({user.picture})"></span>
              <div class="flex-fill">
                <div class="font-weight-medium">{user.name} {#if user.active === false}<span class="badge bg-orange-lt">Disabled</span>{/if}</div>
                <!-- <div class="text-secondary"><a href="#" class="text-reset">{user.Email}</a></div> -->
                <div class="mt-2 list-inline list-inline-dots mb-0 text-secondary d-sm-block d-none">
                                  <div class="list-inline-item"><!-- Download SVG icon from http://tabler-icons.io/i/building-community -->
                                    <i class="ti ti-mail"></i>
                                    {user.email}</div>
                                </div>
              </div>
            </div>
          </td>
          <td class="text-secondary" title={formatDateText(user.CreatedAt)}><i class="ti ti-license"></i> {formatDate(user.CreatedAt)}</td>
          <td class="text-secondary" title={user.lastSeen ? formatDateText(user.lastSeen) : ''}>{user.lastSeen ? formatDate(user.lastSeen) : 'Never'}</td>
          <td data-label="Role">
            <select class="form-select" disabled={userIsAdmin(user.role)} onchange={updateUser(user)}>
              {#each roles as role}
                <option value="{role}" class="text-capitalize" selected={isRole(user.role, role)}>{role}</option>
              {/each}
            </select>
          </td>
          <td>
            <i class="ti ti-dots cursor-pointer" onclick={stopPropagation(preventDefault(() => showDropdown[index] = !showDropdown[index]))}></i>
            <Dropdown bind:show={showDropdown[index]}>
              <a class="dropdown-item" href="#" onclick={()=> resetMFA(user)}>Reset MFA</a>
              {#if user.active === false}
                <a class="dropdown-item text-green" href="#" onclick={()=> toggleUserActive(user)}>Activate User</a>
              {:else}
                <a class="dropdown-item text-orange" href="#" onclick={()=> toggleUserActive(user)}>Deactivate User</a>
              {/if}
              <div class="dropdown-divider"></div>
              <a class="dropdown-item text-red" href="#" onclick={()=> deleteUser(user)}>Delete User</a>
            </Dropdown>
          </td>
        </tr>
      {/each}
      </tbody>
    </table>
  </div>
</div>

<DeleteModal bind:showDeleteModal onDelete={deleteUserPrompted} deleteButtonText={deleteDialogButton}>
  <div class="mt-3">
    {#if userMarkedForDeletion}
      <div class="d-flex align-items-center justify-content-center mb-3">
        <span class="avatar avatar-md me-3" style="background-image: url({userMarkedForDeletion.picture})"></span>
        <div class="text-start">
          <div class="fw-bold">{userMarkedForDeletion.name}</div>
          <div class="text-secondary">{userMarkedForDeletion.email}</div>
        </div>
      </div>
      <div class="text-secondary">This action is irreversible.</div>
    {/if}
  </div>
</DeleteModal>
<InfoModal bind:showInfoModal onOK={resetMFAok} buttonText="Reset MFA">
  <div class="mt-3">
    {#if userToResetMFA}
      <div class="d-flex align-items-center justify-content-center mb-3">
        <span class="avatar avatar-md me-3" style="background-image: url({userToResetMFA.picture})"></span>
        <div class="text-start">
          <div class="fw-bold">{userToResetMFA.name}</div>
          <div class="text-secondary">{userToResetMFA.email}</div>
        </div>
      </div>
      <div class="text-secondary mb-2">
        This will reset all MFA methods for this user. The next time they log in, they will have to set up MFA again.
      </div>
      {#if loadingMfaStatus}
        <div class="text-secondary">Loading MFA details...</div>
      {:else if mfaStatus}
        <div class="list-group list-group-flush mt-2">
          <div class="list-group-item d-flex justify-content-between align-items-center px-0">
            <span>TOTP authenticator</span>
            {#if mfaStatus.hasOTP}
              <span class="badge bg-green-lt">1 configured</span>
            {:else}
              <span class="badge bg-secondary-lt">Not configured</span>
            {/if}
          </div>
          <div class="list-group-item d-flex justify-content-between align-items-center px-0">
            <span>Passkeys</span>
            {#if mfaStatus.passkeyCount > 0}
              <span class="badge bg-green-lt">{mfaStatus.passkeyCount} registered</span>
            {:else}
              <span class="badge bg-secondary-lt">None</span>
            {/if}
          </div>
        </div>
        {#if !mfaStatus.hasOTP && mfaStatus.passkeyCount === 0}
          <div class="alert alert-warning mt-2 mb-0 py-2">
            This user has no MFA methods configured.
          </div>
        {/if}
      {/if}
    {/if}
  </div>
</InfoModal>
<InfoModal bind:showInfoModal={showToggleModal} onOK={toggleUserActivePrompted} buttonText={toggleDialogButton}>
  <div class="mt-3">
    {#if userMarkedForToggle}
      <div class="d-flex align-items-center justify-content-center mb-3">
        <span class="avatar avatar-md me-3" style="background-image: url({userMarkedForToggle.picture})"></span>
        <div class="text-start">
          <div class="fw-bold">{userMarkedForToggle.name}</div>
          <div class="text-secondary">{userMarkedForToggle.email}</div>
        </div>
      </div>
      <div class="text-secondary">
        {#if userMarkedForToggle.active !== false}
          The user will no longer be able to log in.
        {:else}
          The user will be able to log in again.
        {/if}
      </div>
    {/if}
  </div>
</InfoModal>

<style>
  :global(td .dropdown){
    position: absolute !important;
  }

  .ml-1{
    margin-left: 1em !important;
  }
</style>