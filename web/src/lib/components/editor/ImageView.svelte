<script>
	import AnnotationEditor from './AnnotationEditor.svelte';

	/**
	 * @typedef {Object} Props
	 * @property {any} node
	 * @property {any} updateAttributes
	 * @property {boolean} [selected]
	 * @property {boolean} [editable]
	 */

	/** @type {Props} */
	let {
		node,
		updateAttributes,
		selected = false,
		editable = true
	} = $props();

	let localNode = $state(node);
	let localSelected = $state(selected);
	let localEditable = $state(editable);
	let annotationOpen = $state(false);
	let altText = $state(localNode.attrs.alt || '');

	let src = $derived(localNode.attrs.src || '');
	let renderedSrc = $derived(localNode.attrs.renderedSrc || '');
	let displaySrc = $derived(renderedSrc || src);

	$effect(() => {
		altText = localNode.attrs.alt || '';
	});

	// Exported methods for TipTap node view integration (Svelte 5 replaces $$set)
	export function update(newNode, newEditable) {
		if (newNode) {
			localNode = newNode;
		}
		if (newEditable !== undefined) {
			localEditable = newEditable;
		}
	}

	export function select() {
		localSelected = true;
	}

	export function deselect() {
		localSelected = false;
	}

	export function destroy() {
		// Cleanup if needed
	}

	function handleImageClick(event) {
		if (!localEditable) {
			event.currentTarget?.dispatchEvent(new CustomEvent('rte-image-click', {
				detail: {
					src: displaySrc,
					originalSrc: src,
					renderedSrc,
					alt: altText,
					wrapperEl: event.currentTarget
				},
				bubbles: true
			}));
			return;
		}
		annotationOpen = true;
	}

	function saveAlt() {
		if (altText !== (localNode.attrs.alt || '')) {
			updateAttributes({ alt: altText });
		}
	}

	function handleDragStart(e) {
		const img = e.target;
		const ghost = img.cloneNode(true);
		ghost.style.width = '120px';
		ghost.style.height = 'auto';
		ghost.style.opacity = '0.8';
		ghost.style.position = 'absolute';
		ghost.style.top = '-9999px';
		document.body.appendChild(ghost);
		e.dataTransfer.setDragImage(ghost, 60, 40);
		setTimeout(() => document.body.removeChild(ghost), 0);
	}
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<figure class="image-node" class:localSelected>
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="image-wrapper" onclick={handleImageClick}>
		<img src={displaySrc} alt={altText} draggable={localEditable} ondragstart={handleDragStart} />
	</div>

	{#if localEditable}
		<figcaption class="image-caption-row">
			<input
				type="text"
				class="caption-input"
				class:has-text={!!altText}
				bind:value={altText}
				onblur={saveAlt}
				onkeydown={(e) => { if (e.key === 'Enter') e.target.blur(); }}
				placeholder="Click to add image caption"
			/>
		</figcaption>
	{:else if altText}
		<figcaption class="image-caption">{altText}</figcaption>
	{/if}
</figure>

{#if localEditable}
	<AnnotationEditor
		bind:open={annotationOpen}
		{src}
		annotations={localNode.attrs.annotations || []}
		crop={localNode.attrs.crop || null}
		on:save={(e) => {
			updateAttributes({
				annotations: e.detail.annotations,
				crop: e.detail.crop,
				renderedSrc: e.detail.renderedSrc || ''
			});
		}}
	/>
{/if}

<style>
	.image-node {
		position: relative;
		display: block;
		width: fit-content;
		max-width: 100%;
		margin: 0.5rem auto;
		padding: 0;
		cursor: default;
	}

	.image-wrapper {
		cursor: pointer;
		text-align: center;
	}

	.image-node img,
	.image-node :global(.annotation-svg),
	.image-node :global(.plain-image) {
		max-width: 100%;
		max-height: 500px;
		width: auto;
		height: auto;
		object-fit: contain;
		border-radius: 5px;
		display: block;
		transition: outline 0.15s;
	}

	.image-node.selected .image-wrapper,
	.image-node:hover .image-wrapper {
		outline-offset: 2px;
		border-radius: 5px;
	}

	.image-caption-row {
		margin-top: 0.25rem;
	}

	.caption-input {
		width: 100%;
		padding: 0.25rem 0.375rem;
		border: 1px solid transparent;
		border-radius: 4px;
		font-size: 0.8125rem;
		font-style: italic;
		text-align: center;
		outline: none;
		background: transparent;
		color: var(--rte-muted, #6c7a91);
		transition: border-color 0.15s, background 0.15s;
	}

	.caption-input::placeholder {
		color: var(--rte-placeholder, #b0b8c4);
	}

	.caption-input:hover {
		background: var(--rte-menu-hover, #f1f5f9);
	}

	.caption-input:focus {
		border-color: var(--rte-accent, #0054a6);
		background: var(--rte-bg, #fff);
		color: var(--rte-menu-color, #1d2939);
		box-shadow: 0 0 0 2px var(--rte-focus-ring, rgba(0, 84, 166, 0.15));
	}

	.caption-input.has-text {
		color: var(--rte-muted, #6c7a91);
	}

	.image-caption {
		margin-top: 0.375rem;
		font-size: 0.8125rem;
		color: var(--rte-muted, #6c7a91);
		font-style: italic;
		text-align: center;
	}
</style>
