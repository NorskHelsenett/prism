import Image from '@tiptap/extension-image';
import ImageView from './ImageView.svelte';
import { mount as mountSvelte } from 'svelte';

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

			function mountImageView() {
				component = mountSvelte(ImageView, {
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

			mountImageView();

			return {
				dom,
				stopEvent(event) {
					const target = event.target;
					if (
						target.closest('figcaption') ||
						target.closest('input') ||
						target.closest('button') ||
						target.closest('.annotation-backdrop')
					) {
						return true;
					}
					if (event instanceof KeyboardEvent) {
						const active = document.activeElement;
						if (active && (active.tagName === 'INPUT' || active.closest('.annotation-backdrop'))) {
							return true;
						}
						return false;
					}
					return false;
				},
				update(updatedNode) {
					if (updatedNode.type.name !== 'image') return false;
					component.update(updatedNode, editor.isEditable);
					return true;
				},
				selectNode() {
					component.select();
				},
				deselectNode() {
					component.deselect();
				},
				destroy() {
					component.destroy();
				}
			};
		};
	}
}).configure({
	inline: false,
	allowBase64: true
});
