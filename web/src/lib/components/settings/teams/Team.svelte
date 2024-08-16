<script>
  import Avatarlist from "$lib/components/calendar/Avatarlist.svelte";
  import Dropdown from "$lib/components/Dropdown.svelte";
  import Icon from "$lib/components/Icon.svelte";
  import Modal from '$lib/components/Modal.svelte';
  import { Fetch } from "$lib/fetchUtil";
  import { onMount } from "svelte";

  let teams = [];
  let roles = [];
  let showModal = false;
  let selectedTeam = {
    id: null,
    name: "",
    role: "visitor",
    archived: false,
    members: []
  };
  let newMemberEmail = "";

  onMount(async () => {
    roles = await Fetch("/api/settings/roles-list");
    await fetchTeams();
  });

  async function fetchTeams() {
    const response = await Fetch("/api/settings/teams");
    teams = response;
  }

  function openNewTeamModal() {
    selectedTeam = { id: null, name: "", role: "visitor", archived: false, members: [] };
    showModal = true;
  }

  function editTeam(team) {
    selectedTeam = { ...team, members: Array.isArray(team.members) ? team.members : JSON.parse(team.members || '[]') };
    showModal = true;
  }

  async function deleteTeam(teamToDelete) {
    await Fetch(`/api/settings/teams/${teamToDelete.ID}`, { method: 'DELETE' });
    await fetchTeams();
  }

  function toggleDropdown(event, index) {
    event.preventDefault();
    event.stopPropagation();
    teams = teams.map((team, i) => ({
      ...team,
      showDropdown: i === index ? !team.showDropdown : false
    }));
  }

  async function storeTeam() {
    const teamToStore = { ...selectedTeam, members: selectedTeam.members };
    if (selectedTeam.ID) {
      await Fetch(`/api/settings/teams/${selectedTeam.ID}`, {
        method: 'PUT',
        body: JSON.stringify(teamToStore)
      });
    } else {
      await Fetch("/api/settings/teams", {
        method: 'POST',
        body: JSON.stringify(teamToStore)
      });
    }
    await fetchTeams();
    showModal = false;
  }

  async function archiveTeam(teamToArchive) {
    await Fetch(`/api/settings/teams/${teamToArchive.ID}/archive`, { method: 'POST' });
    await fetchTeams();
  }

  async function addMember() {
    await Fetch(`/api/settings/teams/${selectedTeam.id}/members`, {
      method: 'POST',
      body: JSON.stringify({ email: newMemberEmail })
    });
    selectedTeam.members = [...selectedTeam.members, newMemberEmail];
    newMemberEmail = "";
    await updateTeamMembers(selectedTeam);
  }

  async function removeMember(email) {
    await Fetch(`/api/settings/teams/${selectedTeam.ID}/members`, {
      method: 'DELETE',
      body: JSON.stringify({ email })
    });
    selectedTeam.members = selectedTeam.members.filter(member => member !== email);
    await updateTeamMembers(selectedTeam);
  }

  async function updateTeamMembers(team) {
    const updatedTeam = await Fetch(`/api/settings/teams/${team.ID}`, {
      method: 'PUT',
      body: JSON.stringify(team)
    });
    teams = teams.map(t => t.ID === updatedTeam.ID ? { ...updatedTeam, members: updatedTeam.members || '[]' } : t);
  }

  async function updateMembersFromAvatarList(event, team) {
    const updatedMembers = event.detail;
    team.members = updatedMembers.map(member => member.email);
    await updateTeamMembers(team);
  }
</script>

<div class="row m-2">
  <div class="col-lg-4">
    <div class="card mb-2">
      <div class="empty">
        <p class="empty-title">Guild and teams</p>
        <p class="empty-subtitle text-secondary">
          Add a new team
        </p>
        <div class="empty-action">
          <button class="btn btn-primary" on:click={openNewTeamModal}>
            <Icon icon="new"/>
            New team
          </button>
        </div>
      </div>
    </div>
  </div>
  {#each teams as team, index }
    <div class="col-lg-4 mb-2">
      <div class="card h-100" class:card-inactive={team.archived}>
        <div class="card-header">
          <h3 class="card-title">{team.name} <span class="card-subtitle">{team.role}</span></h3>
          <ul class="nav nav-pills card-header-pills">
            <li class="nav-item ms-auto">
              <button class="nav-link" on:click|preventDefault|stopPropagation={(event) => toggleDropdown(event, index)}>
                <Icon icon="cog" />
              </button>
            </li>
          </ul>
          <Dropdown bind:show={team.showDropdown}>
            <a class="dropdown-item" href="#" on:click|preventDefault={() => editTeam(team)}>
              <Icon icon="edit" stroke="1" class="dropdown-item-icon"/>
              Edit
            </a>
            <a class="dropdown-item" href="#" on:click|preventDefault={() => archiveTeam(team)}>
              <Icon icon="archive" stroke="1" class="dropdown-item-icon"/>
              Archive
            </a>
            <a class="dropdown-item" href="#" on:click|preventDefault={() => deleteTeam(team)}>
              <Icon icon="delete" stroke="1" class="dropdown-item-icon"/>
              Delete
            </a>
          </Dropdown>
        </div>
        <div class="card-body">
          <Avatarlist
            hackers={Array.isArray(team.members) ? team.members.map(email => ({ email })) : JSON.parse(team.members || '[]').map(email => ({ email }))}
            on:updateHackers={(event) => updateMembersFromAvatarList(event, team)}
          />
        </div>
      </div>
    </div>
  {/each}
</div>

<Modal bind:showModal large={false} showHeader={false}>
  <div class="row m-3 align-items-end">
    <div class="col">
      <label class="form-label">Name</label>
      <input type="text" class="form-control" bind:value={selectedTeam.name}>
    </div>
    <div class="col">
      <select class="form-select" bind:value={selectedTeam.role}>
        {#each roles as role}
          <option value="{role}" class="text-capitalize">{role}</option>
        {/each}
      </select>
    </div>
  </div>

  <div class="modal-footer" slot="footer">
    <button type="button" class="btn me-auto" on:click={() => showModal = false}>Close</button>
    <button type="button" class="btn btn-primary" on:click={storeTeam}>Save</button>
  </div>
</Modal>

<style>
  .card-inactive {
    opacity: 0.6;
  }
</style>