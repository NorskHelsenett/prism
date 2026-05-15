<script>
  import { onMount, onDestroy } from 'svelte';
  import { slide } from 'svelte/transition';

  /**
   * @typedef {Object} Props
   * @property {boolean} [show]
   * @property {import('svelte').Snippet} [children]
   */

  /** @type {Props} */
  let { show = $bindable(false), children } = $props();
  let dropdownElement = $state();

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
      {@render children?.()}
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