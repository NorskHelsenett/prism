<script>
	import { createEventDispatcher } from 'svelte';

	const dispatch = createEventDispatcher();

	/** The image src to annotate */
	export let src = '';
	/** Existing annotations to load */
	export let annotations = [];
	/** Existing crop to load */
	export let crop = null;
	/** Whether the modal is open */
	export let open = false;

	let canvas;
	let ctx;
	let imgEl;
	let canvasWidth = 0;
	let canvasHeight = 0;
	let naturalWidth = 0;
	let naturalHeight = 0;
	let scale = 1;

	// Tool state
	let activeTool = 'pointer'; // pointer | arrow | rect | text | freehand | eraser-dot | eraser-all | crop
	let activeColor = '#ff0000';
	let strokeWidth = 3;
	let elements = [...annotations];
	let currentElement = null;
	let isDrawing = false;
	let dragTarget = null;
	let dragOffset = { x: 0, y: 0 };

	// Crop state
	let cropRect = crop ? { ...crop } : null;
	let isCropping = false;
	let cropStart = null;

	// Text input state
	let textInputVisible = false;
	let textInputX = 0;
	let textInputY = 0;
	let textInputValue = '';
	let textInput;

	const colors = ['#ff0000', '#ff6600', '#ffcc00', '#00cc00', '#0066ff', '#9933ff', '#000000', '#ffffff'];
	const tools = [
		{ id: 'pointer', label: 'Select', icon: 'M3 3l7.07 16.97 2.51-7.39 7.39-2.51L3 3z' },
		{ id: 'arrow', label: 'Arrow', icon: 'M5 12h14M12 5l7 7-7 7' },
		{ id: 'rect', label: 'Rectangle', icon: 'M3 3h18v18H3z' },
		{ id: 'text', label: 'Text', icon: 'M5 4h14M12 4v16M9 20h6' },
		{ id: 'freehand', label: 'Freehand', icon: 'M3 17c3.333-3.333 5-6.667 8-8 3-1.333 5.667 1.333 7 0' },
		{ id: 'eraser-dot', label: 'Eraser', icon: 'M19 20H5M18 7L8.7 17.3c-.4.4-1 .4-1.4 0L4 14l10-10 4 3z' },
		{ id: 'eraser-all', label: 'Clear All', icon: 'M3 6h18M8 6V4h8v2M5 6v14h14V6M10 11v6M14 11v6' },
		{ id: 'crop', label: 'Crop', icon: 'M6 2v4H2v2h4v14h2V8h10v10h4v-2h-2V6H8V2H6z' }
	];

	// Set canvas context whenever the canvas element becomes available (after {#if open} renders)
	$: if (canvas) {
		ctx = canvas.getContext('2d');
	}

	function onImageLoad() {
		naturalWidth = imgEl.naturalWidth;
		naturalHeight = imgEl.naturalHeight;

		// Fit to container (max 900px wide)
		const maxW = Math.min(900, window.innerWidth - 120);
		const maxH = window.innerHeight - 200;
		scale = Math.min(maxW / naturalWidth, maxH / naturalHeight, 1);
		canvasWidth = Math.round(naturalWidth * scale);
		canvasHeight = Math.round(naturalHeight * scale);

		requestAnimationFrame(redraw);
	}

	function redraw() {
		if (!ctx || !imgEl) return;
		ctx.clearRect(0, 0, canvasWidth, canvasHeight);

		// Draw image (or cropped region)
		ctx.drawImage(imgEl, 0, 0, canvasWidth, canvasHeight);

		// Draw all elements
		for (const el of elements) {
			drawElement(el);
		}

		// Draw current in-progress element
		if (currentElement) {
			drawElement(currentElement);
		}

		// Draw crop overlay
		if (cropRect && activeTool === 'crop') {
			drawCropOverlay();
		}
	}

	function drawElement(el) {
		ctx.save();
		ctx.strokeStyle = el.color;
		ctx.fillStyle = el.color;
		ctx.lineWidth = el.strokeWidth * scale;
		ctx.lineCap = 'round';
		ctx.lineJoin = 'round';

		switch (el.type) {
			case 'arrow': {
				const sx = el.x1 * scale, sy = el.y1 * scale;
				const ex = el.x2 * scale, ey = el.y2 * scale;
				ctx.beginPath();
				ctx.moveTo(sx, sy);
				ctx.lineTo(ex, ey);
				ctx.stroke();
				// Arrowhead
				const angle = Math.atan2(ey - sy, ex - sx);
				const headLen = 12 * scale;
				ctx.beginPath();
				ctx.moveTo(ex, ey);
				ctx.lineTo(ex - headLen * Math.cos(angle - 0.4), ey - headLen * Math.sin(angle - 0.4));
				ctx.moveTo(ex, ey);
				ctx.lineTo(ex - headLen * Math.cos(angle + 0.4), ey - headLen * Math.sin(angle + 0.4));
				ctx.stroke();
				break;
			}
			case 'rect': {
				const x = Math.min(el.x1, el.x2) * scale;
				const y = Math.min(el.y1, el.y2) * scale;
				const w = Math.abs(el.x2 - el.x1) * scale;
				const h = Math.abs(el.y2 - el.y1) * scale;
				ctx.strokeRect(x, y, w, h);
				break;
			}
			case 'text': {
				const fontSize = Math.max(14, el.fontSize || 16) * scale;
				ctx.font = `bold ${fontSize}px sans-serif`;
				// Background
				const metrics = ctx.measureText(el.text);
				const pad = 4 * scale;
				ctx.fillStyle = 'rgba(0,0,0,0.5)';
				ctx.fillRect(
					el.x * scale - pad,
					el.y * scale - fontSize + pad,
					metrics.width + pad * 2,
					fontSize + pad
				);
				ctx.fillStyle = el.color;
				ctx.fillText(el.text, el.x * scale, el.y * scale);
				break;
			}
			case 'freehand': {
				if (el.points.length < 2) break;
				ctx.beginPath();
				ctx.moveTo(el.points[0].x * scale, el.points[0].y * scale);
				for (let i = 1; i < el.points.length; i++) {
					ctx.lineTo(el.points[i].x * scale, el.points[i].y * scale);
				}
				ctx.stroke();
				break;
			}
		}
		ctx.restore();
	}

	function drawCropOverlay() {
		if (!cropRect) return;
		ctx.save();
		ctx.fillStyle = 'rgba(0,0,0,0.5)';
		ctx.fillRect(0, 0, canvasWidth, canvasHeight);
		// Clear the crop area
		const x = cropRect.x * scale;
		const y = cropRect.y * scale;
		const w = cropRect.width * scale;
		const h = cropRect.height * scale;
		ctx.clearRect(x, y, w, h);
		ctx.drawImage(imgEl, 0, 0, canvasWidth, canvasHeight);
		// Re-darken outside crop
		ctx.fillStyle = 'rgba(0,0,0,0.5)';
		ctx.fillRect(0, 0, canvasWidth, y); // top
		ctx.fillRect(0, y + h, canvasWidth, canvasHeight - y - h); // bottom
		ctx.fillRect(0, y, x, h); // left
		ctx.fillRect(x + w, y, canvasWidth - x - w, h); // right
		// Border
		ctx.strokeStyle = '#fff';
		ctx.lineWidth = 2;
		ctx.setLineDash([6, 3]);
		ctx.strokeRect(x, y, w, h);
		ctx.restore();
	}

	function getPos(e) {
		const rect = canvas.getBoundingClientRect();
		return {
			x: (e.clientX - rect.left) / scale,
			y: (e.clientY - rect.top) / scale
		};
	}

	function hitTest(pos) {
		// Simple hit test for pointer/eraser-dot
		for (let i = elements.length - 1; i >= 0; i--) {
			const el = elements[i];
			const threshold = 10;
			switch (el.type) {
				case 'arrow':
				case 'rect':
					if (pos.x >= Math.min(el.x1, el.x2) - threshold &&
						pos.x <= Math.max(el.x1, el.x2) + threshold &&
						pos.y >= Math.min(el.y1, el.y2) - threshold &&
						pos.y <= Math.max(el.y1, el.y2) + threshold) return i;
					break;
				case 'text':
					if (pos.x >= el.x - threshold && pos.x <= el.x + 150 &&
						pos.y >= el.y - 20 && pos.y <= el.y + threshold) return i;
					break;
				case 'freehand':
					for (const pt of el.points) {
						if (Math.abs(pt.x - pos.x) < threshold && Math.abs(pt.y - pos.y) < threshold) return i;
					}
					break;
			}
		}
		return -1;
	}

	function handleMouseDown(e) {
		e.preventDefault();
		const pos = getPos(e);
		isDrawing = true;

		switch (activeTool) {
			case 'pointer': {
				const idx = hitTest(pos);
				if (idx >= 0) {
					dragTarget = idx;
					const el = elements[idx];
					dragOffset = { x: pos.x - (el.x1 ?? el.x ?? el.points?.[0]?.x ?? 0), y: pos.y - (el.y1 ?? el.y ?? el.points?.[0]?.y ?? 0) };
				}
				break;
			}
			case 'arrow':
			case 'rect':
				currentElement = { type: activeTool, x1: pos.x, y1: pos.y, x2: pos.x, y2: pos.y, color: activeColor, strokeWidth };
				break;
			case 'freehand':
				currentElement = { type: 'freehand', points: [pos], color: activeColor, strokeWidth };
				break;
			case 'text':
				textInputX = pos.x;
				textInputY = pos.y;
				textInputValue = '';
				textInputVisible = true;
				setTimeout(() => textInput?.focus(), 0);
				break;
			case 'eraser-dot': {
				const idx = hitTest(pos);
				if (idx >= 0) {
					elements = elements.filter((_, i) => i !== idx);
					redraw();
				}
				break;
			}
			case 'eraser-all':
				elements = [];
				redraw();
				break;
			case 'crop':
				isCropping = true;
				cropStart = pos;
				cropRect = { x: pos.x, y: pos.y, width: 0, height: 0 };
				break;
		}
	}

	function handleMouseMove(e) {
		if (!isDrawing) return;
		e.preventDefault();
		const pos = getPos(e);

		switch (activeTool) {
			case 'pointer':
				if (dragTarget !== null) {
					const el = elements[dragTarget];
					const dx = pos.x - dragOffset.x - (el.x1 ?? el.x ?? el.points?.[0]?.x ?? 0);
					const dy = pos.y - dragOffset.y - (el.y1 ?? el.y ?? el.points?.[0]?.y ?? 0);
					if (el.type === 'arrow' || el.type === 'rect') {
						el.x1 += dx; el.y1 += dy; el.x2 += dx; el.y2 += dy;
					} else if (el.type === 'text') {
						el.x += dx; el.y += dy;
					} else if (el.type === 'freehand') {
						el.points = el.points.map(p => ({ x: p.x + dx, y: p.y + dy }));
					}
					dragOffset = { x: pos.x - (el.x1 ?? el.x ?? el.points?.[0]?.x ?? 0), y: pos.y - (el.y1 ?? el.y ?? el.points?.[0]?.y ?? 0) };
					elements = elements;
					redraw();
				}
				break;
			case 'arrow':
			case 'rect':
				if (currentElement) {
					currentElement.x2 = pos.x;
					currentElement.y2 = pos.y;
					redraw();
				}
				break;
			case 'freehand':
				if (currentElement) {
					currentElement.points.push(pos);
					redraw();
				}
				break;
			case 'eraser-dot': {
				const idx = hitTest(pos);
				if (idx >= 0) {
					elements = elements.filter((_, i) => i !== idx);
					redraw();
				}
				break;
			}
			case 'crop':
				if (isCropping && cropStart) {
					cropRect = {
						x: Math.min(cropStart.x, pos.x),
						y: Math.min(cropStart.y, pos.y),
						width: Math.abs(pos.x - cropStart.x),
						height: Math.abs(pos.y - cropStart.y)
					};
					redraw();
				}
				break;
		}
	}

	function handleMouseUp() {
		if (currentElement) {
			elements = [...elements, currentElement];
			currentElement = null;
		}
		isDrawing = false;
		dragTarget = null;
		isCropping = false;
		redraw();
	}

	function commitText() {
		if (textInputValue.trim()) {
			elements = [...elements, {
				type: 'text',
				x: textInputX,
				y: textInputY,
				text: textInputValue,
				color: activeColor,
				fontSize: 16,
				strokeWidth
			}];
			redraw();
		}
		textInputVisible = false;
		textInputValue = '';
	}

	function handleTextKeydown(e) {
		if (e.key === 'Enter') commitText();
		else if (e.key === 'Escape') { textInputVisible = false; textInputValue = ''; }
	}

	function save() {
		// Render final image with annotations burned in
		const hasAnnotations = elements.length > 0 || cropRect;
		let renderedSrc = '';

		if (hasAnnotations && imgEl && naturalWidth && naturalHeight) {
			const offscreen = document.createElement('canvas');
			const offCtx = offscreen.getContext('2d');

			// Work at full natural resolution
			const prevScale = scale;
			const prevW = canvasWidth;
			const prevH = canvasHeight;

			scale = 1;
			canvasWidth = naturalWidth;
			canvasHeight = naturalHeight;

			offscreen.width = naturalWidth;
			offscreen.height = naturalHeight;

			// Draw image
			offCtx.drawImage(imgEl, 0, 0, naturalWidth, naturalHeight);

			// Draw all annotation elements at full resolution
			const prevCtx = ctx;
			ctx = offCtx;
			for (const el of elements) {
				drawElement(el);
			}
			ctx = prevCtx;

			// Apply crop by copying the cropped region
			if (cropRect) {
				const cropped = document.createElement('canvas');
				cropped.width = cropRect.width;
				cropped.height = cropRect.height;
				const cropCtx = cropped.getContext('2d');
				cropCtx.drawImage(offscreen, cropRect.x, cropRect.y, cropRect.width, cropRect.height, 0, 0, cropRect.width, cropRect.height);
				renderedSrc = cropped.toDataURL('image/png');
			} else {
				renderedSrc = offscreen.toDataURL('image/png');
			}

			// Restore
			scale = prevScale;
			canvasWidth = prevW;
			canvasHeight = prevH;
		}

		dispatch('save', {
			annotations: elements,
			crop: cropRect,
			renderedSrc: renderedSrc || ''
		});
		open = false;
	}

	function close() {
		save();
	}

	$: if (open) {
		elements = [...annotations];
		cropRect = crop ? { ...crop } : null;
	}

	$: if (canvas && canvasWidth && canvasHeight) {
		requestAnimationFrame(redraw);
	}
</script>

{#if open}
<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-static-element-interactions -->
<div class="annotation-backdrop" on:click|self={close}>
	<div class="annotation-modal">
		<div class="annotation-toolbar">
			{#each tools as tool}
				<button
					class="tool-btn"
					class:active={activeTool === tool.id}
					on:click={() => { activeTool = tool.id; }}
					title={tool.label}
				>
					<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"
						stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round">
						<path d={tool.icon} />
					</svg>
				</button>
			{/each}

			<div class="toolbar-divider"></div>

			{#each colors as c}
				<button
					class="color-btn"
					class:active={activeColor === c}
					style="background: {c}; {c === '#ffffff' ? 'border: 1px solid #ccc;' : ''}"
					on:click={() => activeColor = c}
					title={c}
				></button>
			{/each}

			<div class="toolbar-divider"></div>

			<label class="stroke-label">
				<input type="range" min="1" max="10" bind:value={strokeWidth} class="stroke-range" />
			</label>

			<div class="toolbar-spacer"></div>

			<button class="tool-btn save-btn" on:click={close} title="Save & Close">
				<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"
					stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round">
					<path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z" />
					<polyline points="17 21 17 13 7 13 7 21" />
					<polyline points="7 3 7 8 15 8" />
				</svg>
			</button>
		</div>

		<div class="annotation-canvas-container">
			<img bind:this={imgEl} {src} alt="" on:load={onImageLoad} class="hidden-img" />
			<canvas
				bind:this={canvas}
				width={canvasWidth}
				height={canvasHeight}
				class="annotation-canvas"
				class:cursor-crosshair={activeTool !== 'pointer' && activeTool !== 'eraser-dot'}
				class:cursor-pointer={activeTool === 'pointer'}
				class:cursor-eraser={activeTool === 'eraser-dot'}
				on:mousedown={handleMouseDown}
				on:mousemove={handleMouseMove}
				on:mouseup={handleMouseUp}
				on:mouseleave={handleMouseUp}
			></canvas>

			{#if textInputVisible}
				<input
					bind:this={textInput}
					type="text"
					class="text-overlay-input"
					style="left: {textInputX * scale}px; top: {textInputY * scale - 20}px; color: {activeColor};"
					bind:value={textInputValue}
					on:keydown={handleTextKeydown}
					on:blur={commitText}
					placeholder="Type text..."
				/>
			{/if}
		</div>
	</div>
</div>
{/if}

<style>
	.annotation-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.7);
		backdrop-filter: blur(4px);
		z-index: 10000;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.annotation-modal {
		background: var(--rte-menu-bg, #fff);
		border-radius: 12px;
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
		overflow: hidden;
		max-width: 95vw;
		max-height: 95vh;
		display: flex;
		flex-direction: row;
	}

	.annotation-toolbar {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
		padding: 10px 8px;
		background: var(--rte-code-bg, #f1f5f9);
		border-right: 1px solid var(--rte-border, #e6e7e9);
		overflow-y: auto;
		flex-shrink: 0;
	}

	.tool-btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 34px;
		height: 34px;
		border: none;
		background: transparent;
		border-radius: 8px;
		cursor: pointer;
		color: var(--rte-menu-color, #1d2939);
		transition: background 0.15s;
	}

	.tool-btn:hover { background: var(--rte-menu-hover, #e2e8f0); }
	.tool-btn.active { background: var(--rte-accent, #0054a6); color: #fff; }

	.toolbar-spacer {
		flex: 1;
	}

	.save-btn {
		background: var(--rte-accent, #0054a6);
		color: #fff;
	}
	.save-btn:hover { opacity: 0.85; }

	.color-btn {
		width: 22px;
		height: 22px;
		border-radius: 50%;
		border: 2px solid transparent;
		cursor: pointer;
		transition: transform 0.15s;
	}

	.color-btn:hover { transform: scale(1.2); }
	.color-btn.active { border-color: var(--rte-accent, #0054a6); transform: scale(1.2); }

	.toolbar-divider {
		height: 1px;
		width: 28px;
		background: var(--rte-border, #e6e7e9);
		margin: 4px 0;
	}

	.stroke-range {
		width: 28px;
		accent-color: var(--rte-accent, #0054a6);
		writing-mode: vertical-lr;
		direction: rtl;
		height: 80px;
	}

	.stroke-label {
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.75rem;
		color: var(--rte-muted, #6c7a91);
	}

	.annotation-canvas-container {
		position: relative;
		overflow: auto;
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 12px;
		background: #1a1a2e;
	}

	.annotation-canvas {
		border-radius: 4px;
	}

	.cursor-crosshair { cursor: crosshair; }
	.cursor-pointer { cursor: default; }
	.cursor-eraser { cursor: pointer; }

	.hidden-img {
		position: absolute;
		opacity: 0;
		pointer-events: none;
		width: 0;
		height: 0;
	}

	.text-overlay-input {
		position: absolute;
		background: rgba(0, 0, 0, 0.6);
		border: 1px solid rgba(255, 255, 255, 0.3);
		border-radius: 4px;
		padding: 4px 8px;
		font-size: 16px;
		font-weight: bold;
		outline: none;
		min-width: 120px;
	}


</style>
