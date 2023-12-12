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
			<span class="bg-warning text-white avatar"
				><!-- Download SVG icon from http://tabler-icons.io/i/currency-dollar -->
<svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-checklist" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M9.615 20h-2.615a2 2 0 0 1 -2 -2v-12a2 2 0 0 1 2 -2h8a2 2 0 0 1 2 2v8" /><path d="M14 19l2 2l4 -4" /><path d="M9 8h4" /><path d="M9 12h2" /></svg>
			</span>
		</div>
		<div class="col">
			<div class="font-weight-medium">
				{$dashboardStore.unresolved}
			</div>
			<div class="text-secondary">Unresolved</div>
		</div>
	</div>
</div>
