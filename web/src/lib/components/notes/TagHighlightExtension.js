import { Extension } from '@tiptap/core';
import { Plugin, PluginKey } from '@tiptap/pm/state';
import { Decoration, DecorationSet } from '@tiptap/pm/view';

const TAG_REGEX = /(?:^|\s)(#[a-zA-Z][\w\-\/]*)/g;

function buildDecorations(doc) {
  const decorations = [];
  doc.descendants((node, pos) => {
    if (!node.isText) return;
    if (node.marks.some((m) => m.type.name === 'code' || m.type.name === 'link')) return;

    const text = node.text || '';
    let match;
    TAG_REGEX.lastIndex = 0;
    while ((match = TAG_REGEX.exec(text)) !== null) {
      const tag = match[1];
      const offset = match.index + match[0].indexOf(tag);
      const from = pos + offset;
      const to = from + tag.length;
      decorations.push(
        Decoration.inline(from, to, { class: 'note-tag-highlight' })
      );
    }
  });
  return DecorationSet.create(doc, decorations);
}

export const TagHighlight = Extension.create({
  name: 'tagHighlight',
  addProseMirrorPlugins() {
    const key = new PluginKey('tagHighlight');
    return [
      new Plugin({
        key,
        state: {
          init(_, { doc }) {
            return buildDecorations(doc);
          },
          apply(tr, old) {
            if (!tr.docChanged) return old;
            return buildDecorations(tr.doc);
          },
        },
        props: {
          decorations(state) {
            return this.getState(state);
          },
        },
      }),
    ];
  },
});
