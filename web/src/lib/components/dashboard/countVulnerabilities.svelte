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
			<span class="bg-twitter text-white avatar"
				><!-- Download SVG icon from http://tabler-icons.io/i/currency-dollar -->
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="icon icon-tabler icon-tabler-shield-half-filled"
					width="24"
					height="24"
					viewBox="0 0 24 24"
					stroke-width="2"
					stroke="currentColor"
					fill="none"
					stroke-linecap="round"
					stroke-linejoin="round"
					><path stroke="none" d="M0 0h24v24H0z" fill="none" /><path
						d="M12 3a12 12 0 0 0 8.5 3a12 12 0 0 1 -8.5 15a12 12 0 0 1 -8.5 -15a12 12 0 0 0 8.5 -3"
					/><path d="M12 3v18" /><path d="M12 11h8.9" /><path d="M12 8h8.9" /><path
						d="M12 5h3.1"
					/><path d="M12 17h6.2" /><path d="M12 14h8" /></svg
				>
			</span>
		</div>
		<div class="col">
			<div class="font-weight-medium">
				{data}
			</div>
			<div class="text-secondary">Vulnerabilities</div>
		</div>
	</div>
</div>