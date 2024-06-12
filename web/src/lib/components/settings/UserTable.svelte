<script>
	import { Fetch } from "$lib/fetchUtil";
	import { onMount } from "svelte";
	import Dropdown from "../Dropdown.svelte";
  import DeleteModal from '$lib/components/DeleteModal.svelte';
  import InfoModal from '$lib/components/modals/InfoModal.svelte';
	import { toast } from "svelte-sonner";

  let users = []
  let roles = []
  let showDeleteModal = false;
  let showInfoModal = false;


  onMount(async () => {
    users = await Fetch("/api/settings/users/all")
    roles = await Fetch("/api/settings/roles-list")
  });

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

  let showDropdown = [false]
  let userToResetMFA = null

  async function resetMFAok() {
    const response = await Fetch(`/api/settings/session/otp/reset/${userToResetMFA.email}`, {method: "DELETE"})
    userToResetMFA = null
    showInfoModal = false

    if(!response.error) {
      toast.success('Successfully reset MFA');
    } else {
      toast.error('Unable to reset MFA');
    }
  }

  function resetMFA(user){
    userToResetMFA = user
    showInfoModal = true
    showDropdown = [false]
  }


  let userMarkedForDeletion = null
  let deleteDialogText = ""
  let deleteDialogButton= ""

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
    } else {
      toast.error('Unable to delete user');
    }
  }
</script>

<div class="card me-3 mb-3 mt-3">
  <div class="table-responsive">
    <table class="table table-vcenter table-mobile-md card-table">
      <thead>
        <tr>
          <th on:click={orderBy("name", true)}><button class="table-sort" data-sort="sort-name">name</button></th>
          <th>Created</th>
          <th>Last seen</th>
          <th><button on:click={orderBy("role", true)} class="table-sort" data-sort="sort-name">Role</button></th>
          <th class="w-1"></th>
        </tr>
      </thead>
      <tbody>
      {#each users as user, index}
        <tr>
          <td data-label="name">
            <div class="d-flex py-1 align-items-center">
              <span class="avatar me-2" style="background-image: url({user.picture})"></span>
              <div class="flex-fill">
                <div class="font-weight-medium">{user.name}</div>
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
          <td class="text-secondary" title={formatDateText(user.UpdatedAt)}>{formatDate(user.UpdatedAt)}</td>
          <td data-label="Role">
            <select class="form-select" disabled={userIsAdmin(user.role)} on:change={updateUser(user)}>
              {#each roles as role}
                <option value="{role}" class="text-capitalize" selected={isRole(user.role, role)}>{role}</option>
              {/each}
            </select>
          </td>
          <td>
            <i class="ti ti-dots cursor-pointer" on:click={() => showDropdown[index] = !showDropdown[index]}></i>
            <Dropdown bind:show={showDropdown[index]}>
              <a class="dropdown-item" href="#" on:click={()=> resetMFA(user)}>Reset MFA</a>
              <div class="dropdown-divider"></div>
              <a class="dropdown-item text-red" href="#" on:click={()=> deleteUser(user)}>Delete User</a>
            </Dropdown>
          </td>
        </tr>
      {/each}
      </tbody>
    </table>
  </div>
</div>

<DeleteModal bind:showDeleteModal onDelete={deleteUserPrompted} deleteButtonText={deleteDialogButton} text={deleteDialogText}/>
<InfoModal bind:showInfoModal onOK={resetMFAok} buttonText="Reset MFA" text="This will reset MFA. The next time the user logs in, {userToResetMFA?.name} has to register for a new MFA flow."/>

<style>
  :global(td .dropdown){
    position: absolute !important;
  }
</style>