<script>
	import { run } from 'svelte/legacy';

	let { severityData = $bindable({ critical: 0, high: 0, medium: 0, low: 0, information: 0 }), vulnerabilities = [] } = $props();
	let progressData = $state([]);

	run(() => {
		if (severityData) {
	        const total = (severityData.information || 0) +
	                      (severityData.low || 0) +
	                      (severityData.medium || 0) +
	                      (severityData.high || 0) +
	                      (severityData.critical || 0);

	        const toPercentage = (value) => total > 0 ? (value / total * 100).toFixed(2) : 0;

	        progressData = [
	            { label: 'Information', value: toPercentage(severityData.information || 0), class: 'bg-info' },
	            { label: 'Low', value: toPercentage(severityData.low || 0), class: 'bg-primary' },
	            { label: 'Medium', value: toPercentage(severityData.medium || 0), class: 'bg-yellow' },
	            { label: 'High', value: toPercentage(severityData.high || 0), class: 'bg-warning' },
	            { label: 'Critical', value: toPercentage(severityData.critical || 0), class: 'bg-danger' }
	        ];
	    }
	});

	let tooltipVisible = $state(false);
	let tooltipX = $state(0);
	let tooltipY = $state(0);

	function showTooltip(event) {
		tooltipVisible = true;
		tooltipX = event.clientX;
		tooltipY = event.clientY;
	}

	function hideTooltip() {
		tooltipVisible = false;
	}

  run(() => {
		if(vulnerabilities?.length > 0){
	    // Reset counts
	    severityData = { critical: 0, high: 0, medium: 0, low: 0, information: 0 };

	    vulnerabilities.forEach(vulnerability => {
	      const criticality = (vulnerability.Vulnerability?.criticality || '').toLowerCase();

	      if (criticality === 'critical') {
	        severityData.critical += 1;
	      } else if (criticality === 'high') {
	        severityData.high += 1;
	      } else if (criticality === 'medium') {
	        severityData.medium += 1;
	      } else if (criticality === 'low') {
	        severityData.low += 1;
	      } else if (criticality === 'information') {
	        severityData.information += 1;
	      }
	    });
	  }
	});
</script>

<div class="progress mb-2" role="presentation" onmousemove={showTooltip} onmouseleave={hideTooltip}>
    {#each progressData as segment}
        <div
            class="progress-bar {segment.class}"
            style="width: {segment.value}%"
            role="progressbar"
            aria-valuenow={segment.value}
            aria-valuemin="0"
            aria-valuemax="100"
        >
            <span class="visually-hidden">{segment.value}% {segment.label}</span>
        </div>
    {/each}
</div>

{#if tooltipVisible}
	<div class="progress-tooltip" style="position: absolute; left: {tooltipX}px; top: {tooltipY}px;">
		<div class="card">
			<div class="card-stamp">
				<div class="card-stamp-icon bg-blue">
					<!-- Download SVG icon from http://tabler-icons.io/i/bell -->
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
				</div>
			</div>
			<div class="card-body">
				<h3 class="card-title">Vulnerabilities</h3>
				<div class="d-flex justify-content-between mb-1 text-secondary">
					<span>Critical:</span>
					<span>{severityData.critical || 0}</span>
				</div>
				<div class="d-flex justify-content-between mb-1 text-secondary">
					<span>High:</span>
					<span>{severityData.high || 0}</span>
				</div>
				<div class="d-flex justify-content-between mb-1 text-secondary">
					<span>Medium:</span>
					<span>{severityData.medium || 0}</span>
				</div>
				<div class="d-flex justify-content-between mb-1 text-secondary">
					<span>Low:</span>
					<span>{severityData.low || 0}</span>
				</div>
				<div class="d-flex justify-content-between mb-1 text-secondary">
					<span>Information:</span>
					<span>{severityData.information || 0}</span>
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
	.progress-tooltip {
		/* background-color: black;
    color: white; */
		max-width: 20em;
		min-width: 15em;
		padding: 5px;
		border-radius: 4px;
		z-index: 1000;
		/* Adjust the positioning as needed */
		/* transform: translate(-50%, -100%); */
	}
</style>
