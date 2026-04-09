<script>
	import { createEventDispatcher } from 'svelte';
	import AnnotationEditor from './AnnotationEditor.svelte';

	export let node;
	export let updateAttributes;
	export let editor;
	export let selected = false;
	export let editable = true;

	let showOverlay = false;
	let annotationOpen = false;
	let altText = node.attrs.alt || '';
	let editingAlt = false;
	let altInput;

	$: src = node.attrs.src || '';
	$: altText = node.attrs.alt || '';

	function handleClick() {
		if (!editable) return;
		showOverlay = !showOverlay;
	}

	function handleClickOutside(event) {
		if (showOverlay && !event.target.closest('.image-overlay-container')) {
			showOverlay = false;
			editingAlt = false;
		}
	}

	function startEditAlt() {
		editingAlt = true;
		setTimeout(() => altInput?.focus(), 0);
	}

	function saveAlt() {
		updateAttributes({ alt: altText });
		editingAlt = false;
	}

	function handleAltKeydown(e) {
		if (e.key === 'Enter') {
			saveAlt();
		} else if (e.key === 'Escape') {
			altText = node.attrs.alt || '';
			editingAlt = false;
		}
	}

	function handleDragStart(e) {
		// Create a small thumbnail as drag ghost instead of the full-size image
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

<svelte:window on:click={handleClickOutside} />

<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-static-element-interactions -->
<div class="image-node" class:selected on:click|stopPropagation={handleClick}>
	<img src={src} alt={altText} draggable={editable} on:dragstart={handleDragStart} />

	{#if altText && !editable}
		<figcaption class="image-caption">{altText}</figcaption>
	{/if}

	{#if showOverlay && editable}
		<div class="image-overlay-container">
			<div class="image-overlay">
				{#if editingAlt}
					<div class="overlay-field">
						<input
							bind:this={altInput}
							type="text"
							bind:value={altText}
							on:keydown={handleAltKeydown}
							on:blur={saveAlt}
							placeholder="Describe this image..."
							class="alt-input"
						/>
					</div>
				{:else}
					<button class="overlay-btn" on:click|stopPropagation={startEditAlt} title="Add caption / alt text">
						<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M4 20h4l10.5 -10.5a2.828 2.828 0 1 0 -4 -4l-10.5 10.5v4" /><path d="M13.5 6.5l4 4" /></svg>
						{altText || 'Add caption'}
					</button>
					<button class="overlay-btn" on:click|stopPropagation={() => { annotationOpen = true; showOverlay = false; }} title="Annotate image">
						<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M3 19a2 2 0 0 0 2 2h14a2 2 0 0 0 2 -2v-14a2 2 0 0 0 -2 -2h-14a2 2 0 0 0 -2 2v14z" /><path d="M3 16l5 -5c.928 -.893 2.072 -.893 3 0l5 5" /><path d="M14 14l1 -1c.928 -.893 2.072 -.893 3 0l3 3" /><path d="M14 8m-1.5 0a1.5 1.5 0 1 0 3 0a1.5 1.5 0 1 0 -3 0" /></svg>
						Edit
					</button>
				{/if}
			</div>
		</div>
	{/if}

	{#if altText && editable && !showOverlay}
		<figcaption class="image-caption">{altText}</figcaption>
	{/if}
</div>

{#if editable}
	<AnnotationEditor
		bind:open={annotationOpen}
		{src}
		annotations={node.attrs.annotations || []}
		crop={node.attrs.crop || null}
		on:save={(e) => {
			updateAttributes({
				annotations: e.detail.annotations,
				crop: e.detail.crop
			});
		}}
	/>
{/if}

<style>
	.image-node {
		position: relative;
		display: inline-block;
		max-width: 100%;
		margin: 0.5rem 0;
		cursor: default;
	}

	.image-node img {
		max-width: 100%;
		max-height: 500px;
		width: auto;
		height: auto;
		object-fit: contain;
		border-radius: 5px;
		display: block;
		transition: outline 0.15s;
	}

	.image-node.selected img,
	.image-node:hover img {
		outline: 2px solid var(--rte-accent, #0054a6);
		outline-offset: 2px;
	}

	.image-overlay-container {
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		padding: 0.5rem;
	}

	.image-overlay {
		display: flex;
		gap: 0.375rem;
		padding: 0.375rem;
		border-radius: 10px;
		background: var(--rte-menu-bg, #fff);
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15), 0 1px 3px rgba(0, 0, 0, 0.1);
		border: 1px solid var(--rte-menu-border, #f0f0f0);
	}

	.overlay-btn {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		padding: 0.375rem 0.625rem;
		border: none;
		background: transparent;
		color: var(--rte-menu-color, #1d2939);
		border-radius: 7px;
		cursor: pointer;
		font-size: 0.8125rem;
		white-space: nowrap;
		transition: background 0.15s;
	}

	.overlay-btn:hover:not(:disabled) {
		background: var(--rte-menu-hover, #f1f5f9);
	}

	.overlay-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.overlay-field {
		flex: 1;
		min-width: 0;
	}

	.alt-input {
		width: 100%;
		padding: 0.375rem 0.5rem;
		border: 1px solid var(--rte-border, #e6e7e9);
		border-radius: 6px;
		font-size: 0.8125rem;
		outline: none;
		background: var(--rte-bg, #fff);
		color: var(--rte-menu-color, #1d2939);
	}

	.alt-input:focus {
		border-color: var(--rte-accent, #0054a6);
		box-shadow: 0 0 0 2px var(--rte-focus-ring, rgba(0, 84, 166, 0.15));
	}

	.image-caption {
		margin-top: 0.375rem;
		font-size: 0.8125rem;
		color: var(--rte-muted, #6c7a91);
		font-style: italic;
		text-align: center;
	}
</style>
