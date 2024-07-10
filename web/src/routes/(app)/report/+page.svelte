<script>
  import { pageMeta } from '$lib/stores/pageMeta';
  import { onMount } from 'svelte';

    onMount(() => {
        pageMeta.set({ pretitle: 'Report',title: 'Overview over vulnerabilites' });
    });
</script>
<h1>Reports</h1>