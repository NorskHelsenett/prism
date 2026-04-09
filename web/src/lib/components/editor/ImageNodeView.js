import Image from '@tiptap/extension-image';
import ImageView from './ImageView.svelte';

export const ImageWithView = Image.extend({
	draggable: true,

	addAttributes() {
		return {
			...this.parent?.(),
			annotations: {
				default: [],
				parseHTML: (el) => {
					try {
						return JSON.parse(el.getAttribute('data-annotations') || '[]');
					} catch { return []; }
				},
				renderHTML: (attrs) => {
					if (!attrs.annotations?.length) return {};
					return { 'data-annotations': JSON.stringify(attrs.annotations) };
				}
			},
			crop: {
				default: null,
				parseHTML: (el) => {
					try {
						return JSON.parse(el.getAttribute('data-crop') || 'null');
					} catch { return null; }
				},
				renderHTML: (attrs) => {
					if (!attrs.crop) return {};
					return { 'data-crop': JSON.stringify(attrs.crop) };
				}
			},
			renderedSrc: {
				default: '',
				parseHTML: (el) => el.getAttribute('data-rendered-src') || '',
				renderHTML: (attrs) => {
					if (!attrs.renderedSrc) return {};
					return { 'data-rendered-src': attrs.renderedSrc };
				}
			}
		};
	},

	addNodeView() {
		return ({ node, getPos, editor }) => {
			const dom = document.createElement('div');
			dom.classList.add('image-nodeview');

			let component;

			function mount() {
				component = new ImageView({
					target: dom,
					props: {
						node,
						editable: editor.isEditable,
						selected: false,
						updateAttributes: (attrs) => {
							if (typeof getPos !== 'function') return;
							const pos = getPos();
							const currentNode = editor.view.state.doc.nodeAt(pos);
							if (!currentNode) return;
							editor.view.dispatch(
								editor.view.state.tr.setNodeMarkup(pos, undefined, {
									...currentNode.attrs,
									...attrs
								})
							);
						}
					}
				});
			}

			mount();

			return {
				dom,
				// stopEvent: return true = Svelte handles it, false = TipTap handles it
				stopEvent(event) {
					// All events inside figcaption or inputs belong to Svelte
					const target = event.target;
					if (
						target.closest('figcaption') ||
						target.closest('input') ||
						target.closest('button') ||
						target.closest('.annotation-backdrop')
					) {
						return true;
					}
					// Keyboard events: if an input is focused, Svelte handles it
					if (event instanceof KeyboardEvent) {
						const active = document.activeElement;
						if (active && (active.tagName === 'INPUT' || active.closest('.annotation-backdrop'))) {
							return true;
						}
						// Otherwise TipTap handles (undo, delete, etc.)
						return false;
					}
					// Everything else (clicks on img, drags) -> TipTap
					return false;
				},
				update(updatedNode) {
					if (updatedNode.type.name !== 'image') return false;
					node = updatedNode;
					component.$set({
						node: updatedNode,
						editable: editor.isEditable
					});
					return true;
				},
				selectNode() {
					component.$set({ selected: true });
				},
				deselectNode() {
					component.$set({ selected: false });
				},
				destroy() {
					component.$destroy();
				}
			};
		};
	}
}).configure({
	inline: false,
	allowBase64: true
});
