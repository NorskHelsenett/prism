<script>
  import { preventDefault } from 'svelte/legacy';

  import { onDestroy, onMount } from 'svelte';
  import { clickOutside } from './clickOutside.js';
  import { userStore } from '$lib/userStore.js';
  import { goto } from '$app/navigation';
  import { slide } from 'svelte/transition'

  function navigate(url) {
    return (event) => {
      event.preventDefault(); // Prevent the default anchor navigation
      closeDropdown();  // If closeDropdown is async, wait for it to complete
      goto(url);              // Use goto for navigation
    };
  }

  let isHidden = $state(true);

  function toggleHidden() {
      isHidden = !isHidden;
  }

  function closeDropdown() {
      isHidden = true;
  }

  let user = $state({
      image: "",
      role: "visitor",
      name: ""
  })

  // Subscribe to the user store
  const unsubscribe = userStore.subscribe(storeUser => {
      if (!storeUser.loading) {
          user.image = storeUser.picture;
          user.role = storeUser.role;
          user.name = storeUser.name;
      }
  });

  // Remember to unsubscribe when the component is destroyed
  onDestroy(unsubscribe);
</script>

<a href="#" onclick={preventDefault(toggleHidden)} class="nav-link d-flex lh-1 text-reset p-0">
    <span class="avatar avatar-sm" style="background-image: url({user.image})"></span>
    <div class="d-none d-xl-block ps-2">
        <div>{user.name}</div>
        <div class="mt-1 small text-secondary text-capitalize">{user.role}</div>
    </div>
</a>
{#if !isHidden}
<div
    use:clickOutside onoutsideClick={closeDropdown}
    class="dropdown-menu dropdown-menu-end dropdown-menu-arrow show"
    data-bs-popper="static"
    transition:slide
>
    <a href="/" class="dropdown-item" onclick={navigate("/")}>Status</a>
    <a href="/settings" class="dropdown-item" onclick={navigate("/settings")}>Settings</a>
    <div class="dropdown-divider"></div>
    <a href="/api/logout" class="dropdown-item" onclick={navigate("/api/logout")}>Logout</a>
</div>
{/if}