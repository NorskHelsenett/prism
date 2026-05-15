<script>
	
	
	
	
	/**
	 * @typedef {Object} Props
	 * @property {string} [src] - Image src
	 * @property {any} [annotations] - Annotation elements array
	 * @property {any} [crop] - Crop rect { x, y, width, height } or null
	 * @property {string} [alt] - Alt text / caption
	 */

	/** @type {Props} */
	let {
		src = '',
		annotations = [],
		crop = null,
		alt = ''
	} = $props();

	let imgWidth = $state(0);
	let imgHeight = $state(0);
	let sizeProbe = $state();

	function onLoad() {
		imgWidth = sizeProbe.naturalWidth;
		imgHeight = sizeProbe.naturalHeight;
	}

	// Compute viewBox for crop
	let viewBox = $derived(crop
		? `${crop.x} ${crop.y} ${crop.width} ${crop.height}`
		: `0 0 ${imgWidth} ${imgHeight}`);
</script>

<figure class="annotated-image">
	<img bind:this={sizeProbe} {src} alt="" onload={onLoad} class="size-probe" />
	{#if annotations.length > 0 || crop}
		<svg
			viewBox={viewBox}
			class="annotation-svg"
			xmlns="http://www.w3.org/2000/svg"
		>
			<image href={src} x="0" y="0" width={imgWidth} height={imgHeight} />

			{#each annotations as el}
				{#if el.type === 'arrow'}
					<line x1={el.x1} y1={el.y1} x2={el.x2} y2={el.y2}
						stroke={el.color} stroke-width={el.strokeWidth} stroke-linecap="round" />
					<!-- arrowhead -->
					{@const angle = Math.atan2(el.y2 - el.y1, el.x2 - el.x1)}
					{@const hl = 12}
					<line x1={el.x2} y1={el.y2}
						x2={el.x2 - hl * Math.cos(angle - 0.4)} y2={el.y2 - hl * Math.sin(angle - 0.4)}
						stroke={el.color} stroke-width={el.strokeWidth} stroke-linecap="round" />
					<line x1={el.x2} y1={el.y2}
						x2={el.x2 - hl * Math.cos(angle + 0.4)} y2={el.y2 - hl * Math.sin(angle + 0.4)}
						stroke={el.color} stroke-width={el.strokeWidth} stroke-linecap="round" />
				{:else if el.type === 'rect'}
					<rect x={Math.min(el.x1, el.x2)} y={Math.min(el.y1, el.y2)}
						width={Math.abs(el.x2 - el.x1)} height={Math.abs(el.y2 - el.y1)}
						stroke={el.color} stroke-width={el.strokeWidth} fill="none" />
				{:else if el.type === 'text'}
					<rect x={el.x - 4} y={el.y - (el.fontSize || 16)}
						width="200" height={(el.fontSize || 16) + 4}
						fill="rgba(0,0,0,0.5)" rx="2" />
					<text x={el.x} y={el.y} fill={el.color}
						font-size="{el.fontSize || 16}px" font-weight="bold" font-family="sans-serif">
						{el.text}
					</text>
				{:else if el.type === 'freehand' && el.points?.length >= 2}
					<polyline
						points={el.points.map(p => `${p.x},${p.y}`).join(' ')}
						stroke={el.color} stroke-width={el.strokeWidth}
						fill="none" stroke-linecap="round" stroke-linejoin="round" />
				{/if}
			{/each}
		</svg>
	{:else}
		<img {src} {alt} class="plain-image" />
	{/if}

	{#if alt}
		<figcaption class="image-caption">{alt}</figcaption>
	{/if}
</figure>

<style>
	.annotated-image {
		display: inline-block;
		margin: 0.5rem auto;
		max-width: 100%;
		text-align: center;
	}

	.size-probe {
		position: absolute;
		width: 0;
		height: 0;
		overflow: hidden;
		pointer-events: none;
		opacity: 0;
	}

	.annotation-svg {
		max-width: 100%;
		height: auto;
		border-radius: 5px;
		display: block;
		margin-left: auto;
		margin-right: auto;
	}

	.plain-image {
		max-width: 100%;
		max-height: 500px;
		height: auto;
		border-radius: 5px;
		display: block;
		margin-left: auto;
		margin-right: auto;
	}

	.image-caption {
		margin-top: 0.375rem;
		font-size: 0.8125rem;
		color: var(--rte-muted, #6c7a91);
		font-style: italic;
		text-align: center;
	}
</style>
