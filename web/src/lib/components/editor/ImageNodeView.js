import Image from '@tiptap/extension-image';
import ImageView from './ImageView.svelte';

export const ImageWithView = Image.extend({
	addNodeView() {
		return ({ node, getPos, editor, HTMLAttributes }) => {
			const dom = document.createElement('div');
			dom.style.display = 'contents';

			let component;

			function mount() {
				component = new ImageView({
					target: dom,
					props: {
						node,
						editor,
						editable: editor.isEditable,
						selected: false,
						updateAttributes: (attrs) => {
							if (typeof getPos === 'function') {
								const pos = getPos();
								const tr = editor.view.state.tr;
								for (const [key, value] of Object.entries(attrs)) {
									tr.setNodeAttribute(pos, key, value);
								}
								editor.view.dispatch(tr);
							}
						}
					}
				});
			}

			mount();

			return {
				dom,
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
