<script>
  import { onDestroy, onMount } from 'svelte';
  import { clickOutside } from './clickOutside.js';
  import { userStore } from '$lib/userStore.js';
  import { goto } from '$app/navigation';

  function navigate(url) {
    return async (event) => {
      event.preventDefault(); // Prevent the default anchor navigation
      await closeDropdown();  // If closeDropdown is async, wait for it to complete
      goto(url);              // Use goto for navigation
    };
  }

  let isHidden = true;

  function toggleHidden() {
      isHidden = !isHidden;
  }

  function closeDropdown() {
      isHidden = true;
  }

  let user = {
      Image: "",
      Role: "visitor",
      Name: ""
  }

  // Subscribe to the user store
  const unsubscribe = userStore.subscribe(storeUser => {
      if (!storeUser.loading) {
          user.Image = storeUser.Picture;
          user.Role = storeUser.Role;
          user.Name = storeUser.Name;
      }
  });

  // Remember to unsubscribe when the component is destroyed
  onDestroy(unsubscribe);
</script>

<a href="#" on:click|preventDefault={toggleHidden} class="nav-link d-flex lh-1 text-reset p-0">
    <span class="avatar avatar-sm" style="background-image: url({user.Image})"></span>
    <div class="d-none d-xl-block ps-2">
        <div>{user.Name}</div>
        <div class="mt-1 small text-secondary text-capitalize">{user.Role}</div>
    </div>
</a>
<div
    use:clickOutside on:outsideClick={closeDropdown}
    hidden={isHidden}
    class="dropdown-menu dropdown-menu-end dropdown-menu-arrow show"
    data-bs-popper="static"
>
    <a href="/" class="dropdown-item" on:click={navigate("/")}>Status</a>
    <a href="/settings" class="dropdown-item" on:click={navigate("/settings")}>Settings</a>
    <div class="dropdown-divider"></div>
    <a href="/api/logout" class="dropdown-item" on:click={navigate("/api/logout")}>Logout</a>
</div>
