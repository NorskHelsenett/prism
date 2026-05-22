<script>
	/**
	 * @typedef {Object} Props
	 * @property {number} [data]
	 */

	/** @type {Props} */
	let { data = 0 } = $props();
</script>

<div class="card-body">
	<div class="row align-items-center">
		<div class="col-auto">
			<span class="bg-green text-white avatar"
				><!-- Download SVG icon from http://tabler-icons.io/i/currency-dollar -->
					<svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-calendar-check" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M11.5 21h-5.5a2 2 0 0 1 -2 -2v-12a2 2 0 0 1 2 -2h12a2 2 0 0 1 2 2v6" /><path d="M16 3v4" /><path d="M8 3v4" /><path d="M4 11h16" /><path d="M15 19l2 2l4 -4" /></svg>
			</span>
		</div>
		<div class="col">
			<div class="font-weight-medium">
				{data}
			</div>
			<div class="text-secondary">Assessments</div>
		</div>
	</div>
</div>
