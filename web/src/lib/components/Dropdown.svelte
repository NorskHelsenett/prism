<script>
  import { onMount, onDestroy } from 'svelte';

  export let show = false;
  let justOpened = true

  function handleClickOutside(event) {
    const dropdownElement = event.target.closest('.dropdown');
    if (!dropdownElement && !justOpened) {
      show = false;
    }
    justOpened = !justOpened
  }

  onMount(() => {
    window.addEventListener('click', handleClickOutside);
  });

  onDestroy(() => {
    window.removeEventListener('click', handleClickOutside);
  });
</script>

{#if show}
  <div class="dropdown" on:click|stopPropagation>
    <div class="dropdown-menu dropdown-menu-right show">
      <slot />
    </div>
  </div>
{/if}


<style>
  .dropdown-menu-right{
    top: 100%;
    right: -20px;
  }

  .dropdown {
    position: relative;
    right: 20px;
  }
</style>