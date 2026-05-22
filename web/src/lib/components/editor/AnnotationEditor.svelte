<script>
	import { run, createBubbler, stopPropagation, preventDefault } from 'svelte/legacy';

	const bubble = createBubbler();
	import { createEventDispatcher } from 'svelte';

	const dispatch = createEventDispatcher();

	
	
	
	
	/**
	 * @typedef {Object} Props
	 * @property {string} [src] - The image src to annotate
	 * @property {any} [annotations] - Existing annotations to load
	 * @property {any} [crop] - Existing crop to load
	 * @property {boolean} [open] - Whether the modal is open
	 */

	/** @type {Props} */
	let {
		src = '',
		annotations = [],
		crop = null,
		open = $bindable(false)
	} = $props();

	let canvas = $state();
	let ctx = $state();
	let imgEl = $state();
	let canvasWidth = $state(0);
	let canvasHeight = $state(0);
	let naturalWidth = 0;
	let naturalHeight = 0;
	let scale = $state(1);

	// Tool state
	let activeTool = $state('pointer'); // pointer | arrow | rect | filled-rect | solid-filled-rect | text | freehand | eraser-dot | eraser-all | crop
	let activeColor = $state('#ff0000');
	let strokeWidth = $state(12);
	let elements = $state([...annotations]);
	let currentElement = null;
	let isDrawing = false;
	let dragTarget = null;
	let dragOffset = { x: 0, y: 0 };
	let eraserTrail = [];
	let hoverPos = null;

	// Crop state
	let cropRect = $state(crop ? { ...crop } : null);
	let isCropping = false;
	let cropStart = null;

	// Text input state
	let textInputVisible = $state(false);
	let textInputX = $state(0);
	let textInputY = $state(0);
	let textInputValue = $state('');
	let textInput = $state();
	let editingTextIndex = -1; // index of text element being edited, -1 = new
	let strokeMenuOpen = $state(false);
	let shapeMenuOpen = $state(false);
	const MIN_TEXT_SIZE = 12;

	const colors = ['#ff0000', '#ff6600', '#ffcc00', '#00cc00', '#0066ff', '#9933ff', '#000000', '#ffffff'];
	const strokeOptions = [4, 8, 12, 16, 20];
	const tools = [
		{ id: 'pointer', label: 'Select', icon: 'M3 3l7.07 16.97 2.51-7.39 7.39-2.51L3 3z' },
		{ id: 'arrow', label: 'Arrow', icon: 'M5 12h14M12 5l7 7-7 7' },
		{ id: 'text', label: 'Text', icon: 'M5 4h14M12 4v16M9 20h6' },
		{ id: 'freehand', label: 'Freehand', icon: 'M12 20h9M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4 12.5-12.5z' },
		{ id: 'eraser-dot', label: 'Eraser', icon: 'M19 20H5M18 7L8.7 17.3c-.4.4-1 .4-1.4 0L4 14l10-10 4 3z' },
		{ id: 'eraser-all', label: 'Clear All', icon: 'M3 6h18M8 6V4h8v2M5 6v14h14V6M10 11v6M14 11v6' }
	];
	const shapeOptions = [
		{ id: 'rect', label: 'Square', filled: false },
		{ id: 'filled-rect', label: 'Filled square', filled: true, solid: false },
		{ id: 'solid-filled-rect', label: 'Solid filled square', filled: true, solid: true }
	];

	let canvasContainer = $state();

	// Set canvas context whenever the canvas element becomes available (after {#if open} renders)
	run(() => {
		if (canvas) {
			ctx = canvas.getContext('2d');
		}
	});

	function onImageLoad() {
		naturalWidth = imgEl.naturalWidth;
		naturalHeight = imgEl.naturalHeight;

		// Fit to the available container space
		const container = canvasContainer;
		const pad = 24; // 12px padding on each side
		const maxW = container ? container.clientWidth - pad : window.innerWidth - 120;
		const maxH = container ? container.clientHeight - pad : window.innerHeight - 200;
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

		// Draw all elements (skip the one being edited)
		for (let i = 0; i < elements.length; i++) {
			if (i === editingTextIndex && textInputVisible) continue;
			drawElement(elements[i]);
		}

		// Draw current in-progress element
		if (currentElement) {
			drawElement(currentElement);
		}

		if (eraserTrail.length > 0) {
			drawEraserTrail();
		}

		if ((activeTool === 'eraser-dot' || activeTool === 'freehand') && hoverPos) {
			drawEraserCursor();
		}

		// Draw crop overlay
		if (cropRect && activeTool === 'crop') {
			drawCropOverlay();
		}
	}

	function drawEraserTrail() {
		if (!ctx || eraserTrail.length === 0) return;
		ctx.save();
		ctx.strokeStyle = 'rgba(255,255,255,0.4)';
		ctx.fillStyle = 'rgba(255,255,255,0.16)';
		ctx.lineWidth = strokeWidth * scale;
		ctx.lineCap = 'round';
		ctx.lineJoin = 'round';

		if (eraserTrail.length === 1) {
			const pt = eraserTrail[0];
			ctx.beginPath();
			ctx.arc(pt.x * scale, pt.y * scale, (strokeWidth * scale) / 2, 0, Math.PI * 2);
			ctx.fill();
		} else if (eraserTrail.length === 2) {
			ctx.beginPath();
			ctx.moveTo(eraserTrail[0].x * scale, eraserTrail[0].y * scale);
			ctx.lineTo(eraserTrail[1].x * scale, eraserTrail[1].y * scale);
			ctx.stroke();
		} else {
			ctx.beginPath();
			ctx.moveTo(eraserTrail[0].x * scale, eraserTrail[0].y * scale);
			for (let i = 1; i < eraserTrail.length - 1; i++) {
				const cx = eraserTrail[i].x * scale;
				const cy = eraserTrail[i].y * scale;
				const nx = eraserTrail[i + 1].x * scale;
				const ny = eraserTrail[i + 1].y * scale;
				const mx = (cx + nx) / 2;
				const my = (cy + ny) / 2;
				ctx.quadraticCurveTo(cx, cy, mx, my);
			}
			const last = eraserTrail[eraserTrail.length - 1];
			ctx.lineTo(last.x * scale, last.y * scale);
			ctx.stroke();
		}
		ctx.restore();
	}

	function drawEraserCursor() {
		if (!ctx || !hoverPos) return;
		const radius = (Math.max(8, strokeWidth) * scale) / 2;
		ctx.save();
		ctx.strokeStyle = 'rgba(255,255,255,0.9)';
		ctx.fillStyle = 'rgba(255,255,255,0.12)';
		ctx.lineWidth = 1.5;
		ctx.beginPath();
		ctx.arc(hoverPos.x * scale, hoverPos.y * scale, radius, 0, Math.PI * 2);
		ctx.fill();
		ctx.stroke();
		ctx.restore();
	}

	function getScreenTextSize() {
		return Math.max(MIN_TEXT_SIZE, strokeWidth * 2);
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
				const angle = Math.atan2(ey - sy, ex - sx);
				const headLen = (10 + el.strokeWidth * 3) * scale;
				const headAngle = 0.45;
				// Shorten line so it meets the arrowhead base
				const lineEndX = ex - headLen * 0.6 * Math.cos(angle);
				const lineEndY = ey - headLen * 0.6 * Math.sin(angle);
				ctx.beginPath();
				ctx.moveTo(sx, sy);
				ctx.lineTo(lineEndX, lineEndY);
				ctx.stroke();
				// Arrowhead — filled triangle at the endpoint
				ctx.beginPath();
				ctx.moveTo(ex, ey);
				ctx.lineTo(ex - headLen * Math.cos(angle - headAngle), ey - headLen * Math.sin(angle - headAngle));
				ctx.lineTo(ex - headLen * Math.cos(angle + headAngle), ey - headLen * Math.sin(angle + headAngle));
				ctx.closePath();
				ctx.fill();
				break;
			}
			case 'rect':
			case 'filled-rect':
			case 'solid-filled-rect': {
				const x = Math.min(el.x1, el.x2) * scale;
				const y = Math.min(el.y1, el.y2) * scale;
				const w = Math.abs(el.x2 - el.x1) * scale;
				const h = Math.abs(el.y2 - el.y1) * scale;
				const r = Math.min(8 * scale, w / 3, h / 3);
				ctx.beginPath();
				ctx.moveTo(x + r, y);
				ctx.arcTo(x + w, y, x + w, y + h, r);
				ctx.arcTo(x + w, y + h, x, y + h, r);
				ctx.arcTo(x, y + h, x, y, r);
				ctx.arcTo(x, y, x + w, y, r);
				ctx.closePath();
				if (el.type === 'filled-rect' || el.type === 'solid-filled-rect') {
					const previousAlpha = ctx.globalAlpha;
					ctx.globalAlpha = el.type === 'solid-filled-rect' ? 1 : 0.22;
					ctx.fill();
					ctx.globalAlpha = previousAlpha;
				}
				if (el.type === 'rect' || el.type === 'filled-rect' || el.type === 'solid-filled-rect') {
					ctx.stroke();
				}
				break;
			}
			case 'text': {
				const fontSize = (el.fontSize || getScreenTextSize() / scale) * scale;
				ctx.font = `bold ${fontSize}px sans-serif`;
				ctx.textBaseline = 'top';
				ctx.fillStyle = el.color;
				ctx.fillText(el.text, el.x * scale, el.y * scale);
				break;
			}
			case 'freehand': {
				if (el.points.length === 1) {
					const pt = el.points[0];
					ctx.beginPath();
					ctx.arc(pt.x * scale, pt.y * scale, (el.strokeWidth * scale) / 2, 0, Math.PI * 2);
					ctx.fill();
					break;
				}
				if (el.points.length < 2) break;
				ctx.beginPath();
				ctx.moveTo(el.points[0].x * scale, el.points[0].y * scale);
				if (el.points.length === 2) {
					ctx.lineTo(el.points[1].x * scale, el.points[1].y * scale);
				} else {
					// Smooth curve through midpoints using quadratic bezier
					for (let i = 1; i < el.points.length - 1; i++) {
						const cx = el.points[i].x * scale;
						const cy = el.points[i].y * scale;
						const nx = el.points[i + 1].x * scale;
						const ny = el.points[i + 1].y * scale;
						const mx = (cx + nx) / 2;
						const my = (cy + ny) / 2;
						ctx.quadraticCurveTo(cx, cy, mx, my);
					}
					// Last point
					const last = el.points[el.points.length - 1];
					ctx.lineTo(last.x * scale, last.y * scale);
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
				case 'filled-rect':
				case 'solid-filled-rect':
					if (pos.x >= Math.min(el.x1, el.x2) - threshold &&
						pos.x <= Math.max(el.x1, el.x2) + threshold &&
						pos.y >= Math.min(el.y1, el.y2) - threshold &&
						pos.y <= Math.max(el.y1, el.y2) + threshold) return i;
					break;
				case 'text': {
					const fontSize = (el.fontSize || getScreenTextSize() / scale) * scale;
					ctx.save();
					ctx.font = `bold ${fontSize}px sans-serif`;
					ctx.textBaseline = 'top';
					const tw = ctx.measureText(el.text).width / scale;
					ctx.restore();
					const th = (el.fontSize || getScreenTextSize() / scale);
					if (pos.x >= el.x - threshold && pos.x <= el.x + tw + threshold &&
						pos.y >= el.y - threshold && pos.y <= el.y + th + threshold) return i;
					break;
				}
				case 'freehand':
					for (const pt of el.points) {
						if (Math.abs(pt.x - pos.x) < threshold && Math.abs(pt.y - pos.y) < threshold) return i;
					}
					break;
			}
		}
		return -1;
	}

	function pointToSegmentDistance(px, py, x1, y1, x2, y2) {
		const dx = x2 - x1;
		const dy = y2 - y1;
		if (dx === 0 && dy === 0) return Math.hypot(px - x1, py - y1);
		const t = Math.max(0, Math.min(1, ((px - x1) * dx + (py - y1) * dy) / (dx * dx + dy * dy)));
		const projX = x1 + t * dx;
		const projY = y1 + t * dy;
		return Math.hypot(px - projX, py - projY);
	}

	function pointHitsTrail(point, radius) {
		if (eraserTrail.length === 0) return false;
		if (eraserTrail.length === 1) {
			return Math.hypot(point.x - eraserTrail[0].x, point.y - eraserTrail[0].y) <= radius;
		}
		for (let i = 1; i < eraserTrail.length; i++) {
			const a = eraserTrail[i - 1];
			const b = eraserTrail[i];
			if (pointToSegmentDistance(point.x, point.y, a.x, a.y, b.x, b.y) <= radius) return true;
		}
		return false;
	}

	function segmentIntersectsTrail(a, b, radius) {
		if (eraserTrail.length === 0) return false;
		if (pointHitsTrail(a, radius) || pointHitsTrail(b, radius)) return true;
		if (eraserTrail.length === 1) {
			return pointToSegmentDistance(eraserTrail[0].x, eraserTrail[0].y, a.x, a.y, b.x, b.y) <= radius;
		}
		for (let i = 1; i < eraserTrail.length; i++) {
			const p = eraserTrail[i - 1];
			const q = eraserTrail[i];
			if (
				pointToSegmentDistance(a.x, a.y, p.x, p.y, q.x, q.y) <= radius ||
				pointToSegmentDistance(b.x, b.y, p.x, p.y, q.x, q.y) <= radius ||
				pointToSegmentDistance(p.x, p.y, a.x, a.y, b.x, b.y) <= radius ||
				pointToSegmentDistance(q.x, q.y, a.x, a.y, b.x, b.y) <= radius
			) {
				return true;
			}
		}
		return false;
	}

	function elementIntersectsTrail(el, radius) {
		switch (el.type) {
			case 'arrow': {
				return segmentIntersectsTrail({ x: el.x1, y: el.y1 }, { x: el.x2, y: el.y2 }, radius);
			}
			case 'rect':
			case 'filled-rect':
			case 'solid-filled-rect': {
				const minX = Math.min(el.x1, el.x2);
				const maxX = Math.max(el.x1, el.x2);
				const minY = Math.min(el.y1, el.y2);
				const maxY = Math.max(el.y1, el.y2);
				const corners = [
					{ x: minX, y: minY },
					{ x: maxX, y: minY },
					{ x: maxX, y: maxY },
					{ x: minX, y: maxY }
				];
				return (
					segmentIntersectsTrail(corners[0], corners[1], radius) ||
					segmentIntersectsTrail(corners[1], corners[2], radius) ||
					segmentIntersectsTrail(corners[2], corners[3], radius) ||
					segmentIntersectsTrail(corners[3], corners[0], radius) ||
					pointHitsTrail({ x: (minX + maxX) / 2, y: (minY + maxY) / 2 }, radius)
				);
			}
			case 'text': {
				const fontSize = el.fontSize || getScreenTextSize() / scale;
				ctx.save();
				ctx.font = `bold ${fontSize * scale}px sans-serif`;
				ctx.textBaseline = 'top';
				const textWidth = ctx.measureText(el.text).width / scale;
				ctx.restore();
				const minX = el.x;
				const maxX = el.x + textWidth;
				const minY = el.y;
				const maxY = el.y + fontSize;
				return (
					pointHitsTrail({ x: minX, y: minY }, radius) ||
					pointHitsTrail({ x: maxX, y: minY }, radius) ||
					pointHitsTrail({ x: maxX, y: maxY }, radius) ||
					pointHitsTrail({ x: minX, y: maxY }, radius) ||
					pointHitsTrail({ x: (minX + maxX) / 2, y: (minY + maxY) / 2 }, radius)
				);
			}
			case 'freehand':
				for (let i = 1; i < el.points.length; i++) {
					if (segmentIntersectsTrail(el.points[i - 1], el.points[i], radius)) return true;
				}
				return el.points.some((point) => pointHitsTrail(point, radius));
			default:
				return false;
		}
	}

	function eraseAtTrail() {
		const radius = Math.max(10, strokeWidth) / 2;
		const next = elements.filter((el) => !elementIntersectsTrail(el, radius));
		if (next.length !== elements.length) {
			elements = next;
		}
	}

	function handleMouseDown(e) {
		e.preventDefault();
		strokeMenuOpen = false;
		shapeMenuOpen = false;
		// Commit any pending text input before starting a new action
		if (textInputVisible) {
			commitText();
		}
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
			case 'filled-rect':
			case 'solid-filled-rect':
				currentElement = { type: activeTool, x1: pos.x, y1: pos.y, x2: pos.x, y2: pos.y, color: activeColor, strokeWidth };
				break;
			case 'freehand':
				currentElement = { type: 'freehand', points: [pos], color: activeColor, strokeWidth };
				break;
			case 'text': {
				const textIdx = hitTest(pos);
				if (textIdx >= 0 && elements[textIdx].type === 'text') {
					// Edit existing text element
					const el = elements[textIdx];
					editingTextIndex = textIdx;
					textInputX = el.x;
					textInputY = el.y;
					textInputValue = el.text;
					activeColor = el.color;
				} else {
					// New text element
					editingTextIndex = -1;
					textInputX = pos.x;
					textInputY = pos.y;
					textInputValue = '';
				}
				textInputVisible = true;
				setTimeout(() => textInput?.focus(), 0);
				break;
			}
			case 'eraser-dot': {
				eraserTrail = [pos];
				eraseAtTrail();
				redraw();
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
		hoverPos = getPos(e);
		if (!isDrawing) {
			if (activeTool === 'eraser-dot' || activeTool === 'freehand') redraw();
			return;
		}
		e.preventDefault();
		const pos = hoverPos;

		switch (activeTool) {
			case 'pointer':
				if (dragTarget !== null) {
					const el = elements[dragTarget];
					const dx = pos.x - dragOffset.x - (el.x1 ?? el.x ?? el.points?.[0]?.x ?? 0);
					const dy = pos.y - dragOffset.y - (el.y1 ?? el.y ?? el.points?.[0]?.y ?? 0);
					if (el.type === 'arrow' || el.type === 'rect' || el.type === 'filled-rect' || el.type === 'solid-filled-rect') {
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
			case 'filled-rect':
			case 'solid-filled-rect':
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
				eraserTrail = [...eraserTrail, pos];
				eraseAtTrail();
				redraw();
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
		eraserTrail = [];
		redraw();
	}

	function handleMouseEnter(e) {
		hoverPos = getPos(e);
		redraw();
	}

	function handleMouseLeave() {
		handleMouseUp();
		hoverPos = null;
		redraw();
	}

	function commitText() {
		if (textInputValue.trim()) {
			const naturalFontSize = getScreenTextSize() / scale;
			const newEl = {
				type: 'text',
				x: textInputX,
				y: textInputY,
				text: textInputValue,
				color: activeColor,
				fontSize: naturalFontSize,
				strokeWidth
			};
			if (editingTextIndex >= 0) {
				// Replace the existing element
				elements[editingTextIndex] = newEl;
				elements = elements;
			} else {
				elements = [...elements, newEl];
			}
			redraw();
		} else if (editingTextIndex >= 0) {
			// Cleared the text — remove the element
			elements = elements.filter((_, i) => i !== editingTextIndex);
			redraw();
		}
		editingTextIndex = -1;
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
		if (textInputVisible) commitText();
		save();
	}

	run(() => {
		if (open) {
			elements = [...annotations];
			cropRect = crop ? { ...crop } : null;
		}
	});

	let backdropMouseDown = false;

	function onBackdropMouseDown(e) {
		strokeMenuOpen = false;
		shapeMenuOpen = false;
		if (e.target === e.currentTarget) backdropMouseDown = true;
	}

	function onBackdropMouseUp(e) {
		if (backdropMouseDown && e.target === e.currentTarget) close();
		backdropMouseDown = false;
	}

	function consumePointerEvent(e) {
		e.stopPropagation();
	}

	function setStrokeWidth(value) {
		strokeWidth = value;
		strokeMenuOpen = false;
	}

	function selectShapeTool(toolId) {
		if (textInputVisible) commitText();
		activeTool = toolId;
		shapeMenuOpen = false;
	}

	function toggleShapeMenu() {
		if (textInputVisible) commitText();
		if (activeTool !== 'rect' && activeTool !== 'filled-rect' && activeTool !== 'solid-filled-rect') {
			activeTool = 'rect';
			shapeMenuOpen = false;
			return;
		}
		shapeMenuOpen = !shapeMenuOpen;
	}

	function isFilledShape(toolId) {
		return toolId === 'filled-rect' || toolId === 'solid-filled-rect';
	}

	function isSolidShape(toolId) {
		return toolId === 'solid-filled-rect';
	}

	let selectedShapeTool = $derived(shapeOptions.some((shape) => shape.id === activeTool) ? activeTool : 'rect');

	run(() => {
		if (canvas && canvasWidth && canvasHeight) {
			requestAnimationFrame(redraw);
		}
	});
</script>

{#if open}
<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="annotation-backdrop" onmousedown={onBackdropMouseDown} onmouseup={onBackdropMouseUp}>
	<div class="annotation-modal" onmousedown={stopPropagation(bubble('mousedown'))} ondragstart={preventDefault(bubble('dragstart'))}>
		<div class="annotation-toolbar" onpointerdown={stopPropagation(bubble('pointerdown'))}>
			{#each tools as tool}
				<button
					class="tool-btn"
					class:active={activeTool === tool.id}
					onclick={() => { if (textInputVisible) commitText(); activeTool = tool.id; shapeMenuOpen = false; }}
					title={tool.label}
				>
					<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"
						stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round">
						<path d={tool.icon} />
					</svg>
				</button>

				{#if tool.id === 'freehand'}
					<div class="shape-control" onpointerdown={stopPropagation(bubble('pointerdown'))}>
						<button
							class="shape-trigger"
							class:active={activeTool === 'rect' || activeTool === 'filled-rect' || activeTool === 'solid-filled-rect' || shapeMenuOpen}
							onclick={toggleShapeMenu}
							title="Shapes"
							type="button"
						>
							<span
								class="shape-trigger-preview"
								class:filled={isFilledShape(selectedShapeTool)}
								class:soft-filled={isFilledShape(selectedShapeTool) && !isSolidShape(selectedShapeTool)}
								style={`border-color:${activeColor};background:${isFilledShape(selectedShapeTool) ? activeColor : 'transparent'};`}
							></span>
						</button>

						{#if shapeMenuOpen}
							<div class="shape-menu">
								{#each shapeOptions as shape}
									<button
										class="shape-option"
										class:active={activeTool === shape.id}
										onclick={() => selectShapeTool(shape.id)}
										title={shape.label}
										type="button"
									>
										<span
											class="shape-option-preview"
											class:filled={shape.filled}
											class:soft-filled={shape.filled && !shape.solid}
											style={`border-color:${activeColor};background:${shape.filled ? activeColor : 'transparent'};`}
										></span>
									</button>
								{/each}
							</div>
						{/if}
					</div>
				{/if}
			{/each}

			<div class="toolbar-divider"></div>

			{#each colors as c}
				<button
					class="color-btn"
					class:active={activeColor === c}
					style="background: {c}; {c === '#ffffff' ? 'border: 1px solid #ccc;' : ''}"
					onclick={() => activeColor = c}
					title={c}
				></button>
			{/each}

			<div class="toolbar-divider"></div>

			<div class="stroke-control" onpointerdown={stopPropagation(bubble('pointerdown'))}>
				<button
					class="stroke-trigger"
					class:active={strokeMenuOpen}
					onclick={() => strokeMenuOpen = !strokeMenuOpen}
					title="Stroke width"
					type="button"
				>
					<span class="stroke-preview-dot" style={`width: ${strokeWidth}px; height: ${strokeWidth}px; background: ${activeColor};`}></span>
					<span class="stroke-value">{strokeWidth}</span>
				</button>

				{#if strokeMenuOpen}
					<div class="stroke-menu">
						{#each strokeOptions as option}
							<button
								class="stroke-option"
								class:active={strokeWidth === option}
								onclick={() => setStrokeWidth(option)}
								title={`Stroke width ${option}`}
								type="button"
							>
								<span class="stroke-option-dot" style={`width: ${option}px; height: ${option}px; background: ${activeColor};`}></span>
							</button>
						{/each}
					</div>
				{/if}
			</div>

			<div class="toolbar-spacer"></div>

			<button class="tool-btn save-btn" onclick={close} title="Save & Close">
				<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"
					stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round">
					<path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z" />
					<polyline points="17 21 17 13 7 13 7 21" />
					<polyline points="7 3 7 8 15 8" />
				</svg>
			</button>
		</div>

		<div class="annotation-canvas-container" bind:this={canvasContainer}>
			<img bind:this={imgEl} {src} alt="" onload={onImageLoad} class="hidden-img" />
			<canvas
				bind:this={canvas}
				width={canvasWidth}
				height={canvasHeight}
				class="annotation-canvas"
				class:cursor-crosshair={activeTool !== 'pointer' && activeTool !== 'eraser-dot' && activeTool !== 'freehand'}
				class:cursor-pointer={activeTool === 'pointer'}
				class:cursor-eraser={activeTool === 'eraser-dot' || activeTool === 'freehand'}
				onmousedown={handleMouseDown}
				onmouseenter={handleMouseEnter}
				onmousemove={handleMouseMove}
				onmouseup={handleMouseUp}
				onmouseleave={handleMouseLeave}
			></canvas>

			{#if textInputVisible}
				<input
					bind:this={textInput}
					type="text"
					class="text-overlay-input"
					style="left: {textInputX * scale}px; top: {textInputY * scale}px; color: {activeColor}; font-size: {getScreenTextSize()}px;"
					bind:value={textInputValue}
					onkeydown={handleTextKeydown}
					onblur={commitText}
					placeholder="Type text..."
				/>
			{/if}
		</div>
	</div>
</div>
{/if}

<style>
	/* Dark theme variable mappings */
	:global([data-bs-theme="dark"]) .annotation-backdrop {
		--rte-menu-bg: var(--tblr-bg-surface, #1e2a3a);
		--rte-menu-color: var(--tblr-body-color, #dcdfe4);
		--rte-code-bg: var(--tblr-bg-surface-secondary, #1a2234);
		--rte-border: var(--tblr-border-color, #3a4658);
		--rte-menu-hover: rgba(255, 255, 255, 0.06);
		--rte-accent: var(--tblr-primary, #206bc4);
		--rte-muted: #8a95a5;
	}

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
		overflow: visible;
		width: 90vw;
		max-width: 1400px;
		height: calc(90vw * 9 / 16);
		max-height: 90vh;
		display: flex;
		flex-direction: row;
		user-select: none;
		-webkit-user-drag: none;
	}

	.annotation-toolbar {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
		padding: 10px 8px;
		background: var(--rte-code-bg, #f1f5f9);
		overflow: visible;
		flex-shrink: 0;
		position: relative;
		z-index: 3;
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
		display: block;
		flex: 0 0 22px;
		width: 22px;
		height: 22px;
		min-width: 22px;
		min-height: 22px;
		aspect-ratio: 1 / 1;
		box-sizing: border-box;
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

	.stroke-control {
		position: relative;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 6px;
		width: 34px;
		overflow: visible;
	}

	.shape-control {
		position: relative;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 6px;
		width: 34px;
		overflow: visible;
	}

	.stroke-trigger,
	.stroke-option,
	.shape-trigger,
	.shape-option {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border: none;
		background: transparent;
		color: var(--rte-menu-color, #1d2939);
		cursor: pointer;
		padding: 0;
	}

	.stroke-trigger:hover,
	.stroke-trigger.active,
	.stroke-option:hover,
	.stroke-option.active,
	.shape-trigger:hover,
	.shape-trigger.active,
	.shape-option:hover,
	.shape-option.active {
		background: var(--rte-menu-hover, #e2e8f0);
	}

	.stroke-trigger {
		flex-direction: column;
		gap: 6px;
		width: 34px;
		min-height: 54px;
		padding: 8px 0;
		border-radius: 10px;
	}

	.shape-trigger {
		width: 34px;
		height: 34px;
		padding: 0;
		border-radius: 10px;
	}

	.stroke-menu {
		position: absolute;
		left: calc(100% + 10px);
		top: 50%;
		transform: translateY(-50%);
		display: flex;
		flex-direction: row;
		gap: 6px;
		padding: 8px;
		border-radius: 12px;
		background: var(--rte-menu-bg, #fff);
		border: 1px solid var(--rte-border, #e6e7e9);
		box-shadow: 0 10px 24px rgba(0, 0, 0, 0.18);
		z-index: 2;
	}

	.shape-menu {
		position: absolute;
		left: calc(100% + 10px);
		top: 50%;
		transform: translateY(-50%);
		display: flex;
		flex-direction: row;
		gap: 6px;
		padding: 8px;
		border-radius: 12px;
		background: var(--rte-menu-bg, #fff);
		border: 1px solid var(--rte-border, #e6e7e9);
		box-shadow: 0 10px 24px rgba(0, 0, 0, 0.18);
		z-index: 2;
	}

	.stroke-preview-dot {
		display: block;
		flex: 0 0 auto;
		border-radius: 999px;
		max-width: 20px;
		max-height: 20px;
	}

	.stroke-value {
		font-size: 0.7rem;
		font-weight: 600;
		color: var(--rte-muted, #6c7a91);
	}

	.stroke-option {
		width: 32px;
		height: 32px;
		min-width: 32px;
		padding: 0;
		border-radius: 8px;
	}

	.stroke-option-dot {
		display: block;
		flex: 0 0 auto;
		border-radius: 999px;
		max-width: 14px;
		max-height: 14px;
	}

	.shape-trigger-preview,
	.shape-option-preview {
		display: block;
		width: 16px;
		height: 16px;
		border: 2px solid currentColor;
		border-radius: 4px;
		box-sizing: border-box;
	}

	.shape-trigger-preview.filled,
	.shape-option-preview.filled {
		opacity: 0.9;
	}

	.shape-trigger-preview.soft-filled,
	.shape-option-preview.soft-filled {
		opacity: 0.22;
	}

	.shape-option {
		width: 32px;
		height: 32px;
		min-width: 32px;
		padding: 0;
		border-radius: 8px;
	}

	.annotation-canvas-container {
		position: relative;
		overflow: auto;
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 12px;
		background: var(--tblr-bg-surface-dark, #2a3a4e);
	}

	:global([data-bs-theme=\"dark\"]) .annotation-canvas-container {
		background: #0a0c11;
	}
	:global([data-bs-theme="dark"]) .annotation-canvas {
		background: #0a0c11;
	}
	.annotation-canvas {
		border-radius: 4px;
		box-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
		background: var(--tblr-body-bg, #fff);
	}

	.cursor-crosshair { cursor: crosshair; }
	.cursor-pointer { cursor: default; }
	.cursor-eraser { cursor: none; }

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
		padding: 4px 6px;
		font-weight: bold;
		outline: none;
		min-width: 120px;
		text-align: left;
		line-height: 1.2;
	}


</style>
