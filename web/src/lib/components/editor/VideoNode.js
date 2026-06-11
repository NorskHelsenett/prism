import { Node } from '@tiptap/core';
import { videoPreviewSrc } from '$lib/utils/inlineImage';

/**
 * Block-level <video> node so inline videos survive the markdown ⇄ HTML
 * round-trip. Markdown shape is image syntax (`![title](src)`) — the marked
 * renderer in RichTextEditor picks <video> over <img> via isVideoSource().
 */
export const Video = Node.create({
	name: 'video',
	group: 'block',
	atom: true,
	draggable: true,

	addAttributes() {
		return {
			src: { default: null },
			title: { default: null }
		};
	},

	parseHTML() {
		return [{ tag: 'video' }];
	},

	renderHTML({ HTMLAttributes }) {
		return ['video', { ...HTMLAttributes, controls: 'true', preload: 'metadata' }];
	},

	addNodeView() {
		return ({ node }) => {
			const video = document.createElement('video');
			video.controls = true;
			video.preload = 'metadata';
			// videoPreviewSrc adds a display-only #t fragment so the browser
			// paints a poster frame; node attrs keep the clean src.
			let currentSrc = node.attrs.src || '';
			video.src = videoPreviewSrc(currentSrc);
			if (node.attrs.title) video.title = node.attrs.title;

			return {
				dom: video,
				// Let the native controls own mouse/touch/keyboard interaction;
				// keep drag events for ProseMirror so the node stays draggable.
				stopEvent(event) {
					return !event.type.startsWith('drag') && event.type !== 'drop';
				},
				update(updatedNode) {
					if (updatedNode.type.name !== 'video') return false;
					if (updatedNode.attrs.src !== currentSrc) {
						currentSrc = updatedNode.attrs.src || '';
						video.src = videoPreviewSrc(currentSrc);
					}
					return true;
				}
			};
		};
	}
});
