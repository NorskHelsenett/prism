<script>
	import { Fetch } from "$lib/fetchUtil";
	import { onMount } from "svelte";

  let users = []
  let roles = []

  onMount(async () => {
    users = await Fetch("/api/settings/users/all")
    roles = await Fetch("/api/settings/roles-list")

    users.sort((a, b) => a.Role.localeCompare(b.Role));
  });

  function isRole(userRole, role){
    return userRole == role
  }
  function userIsAdmin(userRole) {
    return userRole == "admin"
  }

  function updateUser(user){
    return async function(event) {
      const newRole = event.target.value;
      user.Role = newRole

      const response = await Fetch('/api/settings/profile', {
        method: 'PUT',
        body: JSON.stringify(user)
      });
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
</script>

<div class="card me-3 mb-3 mt-3">
  <div class="table-responsive">
    <table class="table table-vcenter table-mobile-md card-table">
      <thead>
        <tr>
          <th>Name</th>
          <th>Created</th>
          <th>Last seen</th>
          <th>Role</th>
          <th class="w-1"></th>
        </tr>
      </thead>
      <tbody>
      {#each users as user}
        <tr>
          <td data-label="name">
            <div class="d-flex py-1 align-items-center">
              <span class="avatar me-2" style="background-image: url({user.Picture})"></span>
              <div class="flex-fill">
                <div class="font-weight-medium">{user.Name}</div>
                <div class="text-secondary"><a href="#" class="text-reset">{user.Email}</a></div>
              </div>
            </div>
          </td>
          <td class="text-secondary" title={formatDateText(user.CreatedAt)}>{formatDate(user.CreatedAt)}</td>
          <td class="text-secondary" title={formatDateText(user.UpdatedAt)}>{formatDate(user.UpdatedAt)}</td>
          <td data-label="Role">
            <select class="form-select" disabled={userIsAdmin(user.Role)} on:change={updateUser(user)}>
              {#each roles as role}
                <option value="{role}" class="text-capitalize" selected={isRole(user.Role, role)}>{role}</option>
              {/each}
            </select>
          </td>
        </tr>
      {/each}
      </tbody>
    </table>
  </div>
</div>