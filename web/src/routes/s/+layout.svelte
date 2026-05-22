<script>
  import '../../app.css'
  import '@tabler/core/dist/css/tabler.min.css';
  import { initializeApiEndpoint, isLoading } from '$lib/stores/configStore';
	import { onMount } from 'svelte';
  /**
   * @typedef {Object} Props
   * @property {import('svelte').Snippet} [children]
   */

  /** @type {Props} */
  let { children } = $props();

  onMount(async () => {
    const anon = true
    await initializeApiEndpoint(anon);
  });

  let isInitialized = $derived(!$isLoading);
</script>

{#if isInitialized}
  {@render children?.()}
{/if}
<style>
  @import url("@tabler/icons-webfont/dist/tabler-icons.min.css");
</style>