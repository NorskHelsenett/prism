<script>
  import { createEventDispatcher, onDestroy, onMount } from 'svelte';
  import { Fetch } from '$lib/fetchUtil';
  import Avatar from '$lib/components/Avatar.svelte';
  import { fade, slide } from 'svelte/transition';
  import { scale } from 'svelte/transition';
  import { quintOut } from 'svelte/easing';

  const dispatch = createEventDispatcher();

  let allUsers = [];
  let availableUsers = $state([]);
  let { hackers = $bindable([]) } = $props();
  let allTeams = $state([]);
  let filterText = $state('');

  let dropdownPosition = $state({ top: 0, left: 0 });
  let addButtonElement = $state();

  onMount(async () => {
    try {
      const response = await Fetch('/api/profile/all');
      allUsers = response.users || [];
      allTeams = response.teams || [];
      updateAvailableUsers();
    } catch (error) {
      console.error('Failed to fetch users:', error);
    }
    window.addEventListener('click', handleClickOutside);
  });

  function updateAvailableUsers() {
      availableUsers = allUsers.filter(user => !hackers.some(hacker => hacker.email === user.email));
  }

  $effect(() => {
    if (hackers) {
      const deduplicated = hackers.filter((hacker, index, self) => 
        index === self.findIndex(h => h.email === hacker.email)
      );
      if (deduplicated.length !== hackers.length) {
        hackers = deduplicated;
      }
      updateAvailableUsers();
    }
  });

  let filteredUsers = $derived(availableUsers.filter(user =>
    user.email.toLowerCase().includes(filterText.toLowerCase()) ||
    (user.name && user.name.toLowerCase().includes(filterText.toLowerCase()))
  ));

  let filteredTeams = $derived(allTeams.filter(team =>
    team.name.toLowerCase().includes(filterText.toLowerCase()) ||
    team.members.some(email => email.toLowerCase().includes(filterText.toLowerCase()))
  ));

  function handleClickOutside(event) {
    const cardElement = document.getElementById('hackersDropdownList');
    if (cardElement && !cardElement.contains(event.target)) {
      showHackersList = false;
    }
  }

  onDestroy(() => {
    window.removeEventListener('click', handleClickOutside);
  });

  let showRemoveHacker = $state([]);
  let showHackersList = $state(false);

  function addHacker(userOrEmail) {
    let email;

    if (typeof userOrEmail === 'string') {
      email = userOrEmail;
    } else if (userOrEmail && typeof userOrEmail === 'object' && userOrEmail.email) {
      email = userOrEmail.email;
    } else {
      console.error('Invalid input to addHacker function');
      return;
    }

    if (!hackers.some(hacker => hacker.email === email)) {
      hackers = [...hackers, { email }];
      updateAvailableUsers();
      dispatch('updateHackers', hackers);
    }
    showHackersList = false;
    filterText = '';
  }

  function removeHacker(hacker) {
    hackers = hackers.filter(h => h.email !== hacker.email);
    updateAvailableUsers();
    dispatch('updateHackers', hackers);
  }

  function isUserAvailable(email) {
    return !hackers.some(hacker => hacker.email === email);
  }

  function hasAvailableMembers(team) {
    return team.members.some(email => isUserAvailable(email));
  }

  function toggleHackersList() {
    showHackersList = !showHackersList;
    
    if (showHackersList && addButtonElement) {
      const rect = addButtonElement.getBoundingClientRect();
      console.log(rect);
      console.log(window.scrollY);
      console.log(window.scrollX);
      dropdownPosition = {
        top: rect.bottom + 5,
        left: rect.left + window.scrollX
      };
      console.log(dropdownPosition);
    }
  }
</script>

<div class="avatar-list" style="position:relative">
  {#each hackers as hacker, index (hacker.email)}
    {#if hacker && hacker.email}
      <div class="avatar-container"
           onmouseenter={() => showRemoveHacker[index] = true}
           onmouseleave={() => showRemoveHacker[index] = false}
           transition:scale={{ duration: 300, delay: 0, opacity: 0.5, start: 0.0, easing: quintOut }}>
        <Avatar email="{hacker.email}" option={{ showName: false, size: "sm", emptyFields: false, circle: true, tooltipEnabled: false}}/>
        {#if showRemoveHacker[index]}
          <i class="overlay ti ti-x rounded-circle"
             transition:fade={{ delay: 50, duration: 500 }}
             onclick={() => removeHacker(hacker)}></i>
        {/if}
      </div>
    {/if}
  {/each}
  <span class="avatar rounded-circle avatar-sm cursor-pointer"
        bind:this={addButtonElement}
        onclick={(e) => { e.stopPropagation(); toggleHackersList(); }}>
    <i class="ti ti-plus"></i>
  </span>
  {#if showHackersList}
    <div id="hackersDropdownList" 
         class="card" 
         style="top: {dropdownPosition.top}px; left: {dropdownPosition.left}px;"
         transition:slide="{{ duration: 100, axis: 'y' }}">
      <div class="filter-input">
        <input type="text" bind:value={filterText} placeholder="Filter users..." />
      </div>
      <div class="">
        {#each filteredTeams as team}
          {#if team.members !== null && team.members.length > 0 && hasAvailableMembers(team)}
            <h5 class="list-header">{team.name}</h5>
            <ul class="team-list">
              {#each team.members as email}
                {#if isUserAvailable(email) && email.toLowerCase().includes(filterText.toLowerCase())}
                  <li class="option selected p-2" onclick={() => addHacker(email)}>
                    <Avatar email={email} option={{ showName: true, emptyFields: false, circle: true, tooltipEnabled: false}}/>
                  </li>
                {/if}
              {/each}
            </ul>
            <hr />
          {/if}
        {/each}
        <ul>
          {#each filteredUsers as user (user.email)}
            <li class="option selected p-2" onclick={() => addHacker(user)}>
              <Avatar email={user.email} option={{ showName: true, emptyFields: false, circle: true, tooltipEnabled: false}}/>
            </li>
          {/each}
        </ul>
      </div>
      {#if filteredUsers.length === 0 && filteredTeams.every(team => !hasAvailableMembers(team))}
        <div class="centered">
          No users match the filter
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .centered {
    text-align: center;
    padding: 10px;
    color: #888;
  }
  .team-list{
    margin-left: 0em;
  }

  #hackersDropdownList{
    position: fixed;
    z-index: 10000;
    max-height: 25em;
    overflow-y: auto;
    min-width: 250px;
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
  }

  .avatar-container {
    cursor: pointer;
    position: relative;
    display: inline-block;
  }

  .overlay {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    justify-content: center;
    align-items: center;
    color: white;
    font-size: 16px;
  }

  ul {
    list-style-type: none;
    padding: 0;
    margin: 0;
  }

  li.option:hover {
    cursor: pointer;
    background-color: rgba(var(--tblr-secondary-rgb), .08);
    color: inherit;
  }

  .list-header{
    margin-left: 5%;
    margin-top: 5%;
    margin-bottom: 0;
  }

  hr{
    margin: 0;
  }

  .filter-input {
    padding: 10px;
  }

  .filter-input input {
    width: 100%;
    padding: 5px;
    border: 1px solid #ccc;
    border-radius: 4px;
  }
</style>