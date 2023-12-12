<script>
	import { onMount, onDestroy } from 'svelte';
	import { dashboardStore } from '$lib/stores/dashboardStore';

	onMount(() => {
		dashboardStore.startPolling(); // Start polling when component mounts
	});

	onDestroy(() => {
		dashboardStore.stopPolling(); // Stop polling when component is destroyed
	});
</script>

<div class="card-body">
	<div class="row align-items-center">
		<div class="col-auto">
			<span class="bg-green text-white avatar"
				><!-- Download SVG icon from http://tabler-icons.io/i/currency-dollar -->
<svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-briefcase" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M3 7m0 2a2 2 0 0 1 2 -2h14a2 2 0 0 1 2 2v9a2 2 0 0 1 -2 2h-14a2 2 0 0 1 -2 -2z" /><path d="M8 7v-2a2 2 0 0 1 2 -2h4a2 2 0 0 1 2 2v2" /><path d="M12 12l0 .01" /><path d="M3 13a20 20 0 0 0 18 0" /></svg>
			</span>
		</div>
		<div class="col">
			<div class="font-weight-medium">
				{$dashboardStore.projects}
			</div>
			<div class="text-secondary">Projects</div>
		</div>
	</div>
</div>
