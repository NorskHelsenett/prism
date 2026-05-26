<script>
	import { Fetch } from '$lib/fetchUtil';
	import { toast } from 'svelte-sonner';

	let { settings = $bindable({}) } = $props();

	const tiers = [
		{ id: 'low', label: 'Low (1280)', value: 1280, hint: '~720p — smallest payloads' },
		{ id: 'medium', label: 'Medium (1920)', value: 1920, hint: 'Full HD — balanced default' },
		{ id: 'high', label: 'High (2560)', value: 2560, hint: '1440p — best for screenshots' },
		{ id: 'custom', label: 'Custom', value: null, hint: 'Specify the long edge in pixels' },
	];

	function tierFor(value) {
		const match = tiers.find((t) => t.value === Number(value));
		return match ? match.id : 'custom';
	}

	let selectedTier = $state(tierFor(settings.attachmentMaxEdge ?? 2560));
	let customValue = $state(Number(settings.attachmentMaxEdge ?? 2560));
	let applying = $state(false);

	async function pickTier(id) {
		if (applying) return;
		selectedTier = id;
		if (id === 'custom') return; // wait for Save click
		const tier = tiers.find((t) => t.id === id);
		settings.attachmentMaxEdge = tier.value;
		customValue = tier.value;
		await saveAndRegen();
	}

	async function applyCustom() {
		if (applying) return;
		const v = Math.max(64, Math.min(8192, Number(customValue) || 2560));
		customValue = v;
		settings.attachmentMaxEdge = v;
		await saveAndRegen();
	}

	async function saveAndRegen() {
		applying = true;
		try {
			const saveRes = await Fetch('/api/settings', {
				method: 'POST',
				body: JSON.stringify(settings),
				headers: { 'Content-Type': 'application/json' },
			});
			if (saveRes?.error) {
				toast.error('Failed to save resolution');
				return;
			}
			const regenRes = await Fetch('/api/settings/attachments/regenerate-proxies', {
				method: 'POST',
			});
			if (regenRes?.error) {
				toast.error('Regeneration failed');
			} else {
				toast.success(
					`Saved · regenerated ${regenRes.regenerated}/${regenRes.total} proxies`,
				);
			}
		} finally {
			applying = false;
		}
	}
</script>

<div class="card-body">
	<h2 class="card-title">Attachment Resolution</h2>
	<p class="text-secondary">
		Long-edge cap for the inline-rendered proxy of each uploaded image. Originals are always
		stored at full resolution; only the proxy is downscaled. Picking a tier saves and
		regenerates every existing proxy.
	</p>

	<div class="form-selectgroup">
		{#each tiers as tier (tier.id)}
			<label class="form-selectgroup-item" class:disabled={applying}>
				<input
					type="radio"
					name="attachment-tier"
					value={tier.id}
					checked={selectedTier === tier.id}
					disabled={applying}
					onchange={() => pickTier(tier.id)}
					class="form-selectgroup-input"
				/>
				<span class="form-selectgroup-label">
					<strong>{tier.label}</strong>
					<span class="d-block text-secondary small">{tier.hint}</span>
				</span>
			</label>
		{/each}
	</div>

	{#if selectedTier === 'custom'}
		<div class="mt-3" style="max-width: 32rem;">
			<label class="form-label" for="attachment-custom">Long edge (pixels)</label>
			<div class="d-flex gap-2">
				<input
					id="attachment-custom"
					type="number"
					class="form-control"
					min="64"
					max="8192"
					step="1"
					bind:value={customValue}
					disabled={applying}
				/>
				<button class="btn btn-primary text-nowrap px-6" disabled={applying} onclick={applyCustom}>
					{applying ? 'Saving…' : 'Save & regenerate'}
				</button>
			</div>
			<small class="form-hint">Between 64 and 8192.</small>
		</div>
	{/if}

	{#if applying}
		<small class="form-hint mt-2 d-block">
			Re-encoding existing proxies from their stored originals — this can take a while on
			larger databases.
		</small>
	{/if}
</div>
