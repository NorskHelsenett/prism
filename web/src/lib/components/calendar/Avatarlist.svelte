<script>
	import { onDestroy, onMount } from 'svelte';
  import 'tom-select/dist/css/tom-select.bootstrap5.min.css';
	import { Fetch } from '$lib/fetchUtil';
	import Avatar from '$lib/components/Avatar.svelte';
  import { fade } from 'svelte/transition';
	import { scale } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
  import { createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher();

  let usersOriginal = []
  let users = []
  export let hackers = []

	onMount(async () => {
		usersOriginal = await Fetch('/api/profile/all');
    users = usersOriginal
    users = usersOriginal.filter(user => !hackers.some(hacker => hacker.email === user.email));

    window.addEventListener('click', handleClickOutside);
  });

  function handleClickOutside(event) {
    const cardElement = document.getElementById('hackersDropdownList');
    if (cardElement && !cardElement.contains(event.target)) {
      showHackersList = false;
    }
  }

  $: if(hackers) {
    users = usersOriginal.filter(user => !hackers.some(hacker => hacker.email === user.email));
  }

  onDestroy(() => {
    window.removeEventListener('click', handleClickOutside);
  });

  let showRemoveHacker = []
  let showHackersList = false

function addHacker(user) {
    user.email = user.email; // Ensure the email property is standardized

    // Check if the user is already in the hackers list based on email
    if (!hackers.some(hacker => hacker.email === user.email)) {
        hackers = [...hackers, user]; // Use spread syntax for immutability
    }

    // Filter the users list to exclude users that are now in the hackers list
    users = users.filter(u => !hackers.some(hacker => hacker.email === u.email));
    dispatch('updateHackers', hackers);
}

function removeHacker(user) {
    user.email = user.email; // Ensure the email property is standardized

    // Filter out the user from the hackers list
    hackers = hackers.filter(hacker => hacker.email !== user.email);

    // Add the user back to the users list if they're not already present
    if (!users.some(u => u.email === user.email)) {
        users = [...users, user]; // Use spread syntax for immutability
    }
    dispatch('updateHackers', hackers);
}

</script>

<div class="avatar-list" style="position:relative">
  {#each hackers as hacker, index (hacker.email)}
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div class="avatar-container" on:mouseenter={() => showRemoveHacker[index] = true} on:mouseleave={() => showRemoveHacker[index] = false} transition:scale={{ duration: 300, delay: 0, opacity: 0.5, start: 0.0, easing: quintOut }}>
      <Avatar email="{hacker.email}" option={{ showName: false, size: "sm", emptyFields: false, circle: true}}/>
      {#if showRemoveHacker[index]}
        <i class="overlay ti ti-x rounded-circle" transition:fade={{ delay: 50, duration: 500 }} on:click="{removeHacker(hacker)}"></i>
      {/if}
    </div>
  {/each}
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <span class="avatar rounded-circle avatar-sm cursor-pointer" on:click|stopPropagation="{() => showHackersList = !showHackersList}"><i class="ti ti-plus"></i></span>
  <div id="hackersDropdownList" class="card" style="position:absolute;margin-top: 42px;" hidden={!showHackersList}>
    <div class="">
      <ul>
        {#each users as user}
          <li class="option selected p-2" on:click="{addHacker(user)}"><Avatar email={user.email}/></li>
        {/each}
      </ul>
    </div>
  </div>
</div>

<style>
  #hackersDropdownList {
    z-index: 1000;
  }
  .avatar-container {
    cursor: pointer;
    position: relative;
    display: inline-block; /* Or 'block' depending on your layout */
  }

  .overlay {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5); /* 50% black overlay */
    display: flex;
    justify-content: center; /* Center horizontally */
    align-items: center; /* Center vertically */
    color: white; /* Text color */
    font-size: 16px; /* Adjust as needed */
  }

  ul {
    list-style-type: none;
    padding: 0;
    margin: 0;
  }
  :global(.item) {
    background: var(--tblr-modal-bg);
    color: var(--tblr-body-color);
  }

  :global(input) {
    background: var(--tblr-modal-bg);
    color: var(--tblr-body-color);
  }

  li.option:hover{
    cursor: pointer;
    background-color: rgba(var(--tblr-secondary-rgb),.08);
    color: inherit;
  }
</style>