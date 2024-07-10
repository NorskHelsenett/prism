<script>
  import '../../app.css'
  import '@tabler/core/dist/css/tabler.min.css';
  import { initializeApiEndpoint, isLoading } from '$lib/stores/configStore';
	import { onMount } from 'svelte';

  onMount(async () => {
    const anon = true
    await initializeApiEndpoint(anon);
  });

  $: isInitialized = !$isLoading;
</script>

{#if isInitialized}
  <slot/>
{/if}
<style>
  @import url("@tabler/icons-webfont/tabler-icons.min.css");
</style>