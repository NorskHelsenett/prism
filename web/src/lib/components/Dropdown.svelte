<script>
  import { onMount, onDestroy } from 'svelte';
  import { slide } from 'svelte/transition';

  export let show = false;
  let dropdownElement;

  function handleClickOutside(event) {
    if (show && dropdownElement && !dropdownElement.contains(event.target)) {
      show = false;
    }
  }

  onMount(() => {
    document.addEventListener('click', handleClickOutside);
  });

  onDestroy(() => {
    document.removeEventListener('click', handleClickOutside);
  });
</script>

{#if show}
  <div class="dropdown" bind:this={dropdownElement}>
    <div class="dropdown-menu dropdown-menu-right show" transition:slide="{{ duration: 100, axis: 'y' }}">
      <slot />
    </div>
  </div>
{/if}

<style>
  .dropdown-menu-right {
    top: 10px;
    right: -15px;
  }

  .dropdown {
    position: relative;
    right: 20px;
  }
</style>