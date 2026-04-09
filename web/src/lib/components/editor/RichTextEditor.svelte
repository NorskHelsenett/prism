<script>
	import { onMount, onDestroy, createEventDispatcher } from 'svelte';
	import { marked } from 'marked';
	import DOMPurify from 'dompurify';
	import TurndownService from 'turndown';
	import { gfm } from '@joplin/turndown-plugin-gfm';

	import { Editor } from '@tiptap/core';
	import StarterKit from '@tiptap/starter-kit';
	import Placeholder from '@tiptap/extension-placeholder';
	import Underline from '@tiptap/extension-underline';
	import Highlight from '@tiptap/extension-highlight';
	import TaskList from '@tiptap/extension-task-list';
	import TaskItem from '@tiptap/extension-task-item';
	import Image from '@tiptap/extension-image';
	import BubbleMenu from '@tiptap/extension-bubble-menu';

	/** Markdown string (two-way bindable) */
	export let value = '';
	/** Placeholder text shown when editor is empty */
	export let placeholder = 'Start writing...';
	/** Whether the editor is editable */
	export let editable = true;
	/** Minimum height of the editor area */
	export let minHeight = '200px';

	const dispatch = createEventDispatcher();

	let element;
	let editor;
	let internalUpdate = false;
	let bubbleMenuElement;

	// -- Turndown: HTML -> Markdown -------------------------------------------

	const turndownService = new TurndownService({
		codeBlockStyle: 'fenced',
		headingStyle: 'atx'
	});
	turndownService.escape = (s) => s;
	turndownService.use(gfm);

	// Override after GFM plugin so our rules take priority

	turndownService.addRule('taskList', {
		filter: (node) =>
			node.nodeName === 'UL' &&
			node.getAttribute('data-type') === 'taskList',
		replacement(content) {
			return '\n' + content + '\n';
		}
	});

	turndownService.addRule('taskListItems', {
		filter: (node) =>
			node.nodeName === 'LI' &&
			(node.getAttribute('data-type') === 'taskItem' ||
				node.getAttribute('data-checked') === 'true' ||
				node.getAttribute('data-checked') === 'false'),
		replacement(content, node) {
			const checked = node.getAttribute('data-checked') === 'true';
			// Strip any checkbox inputs that TipTap may render inside the <li>
			content = content
				.replace(/^\s*\[[ x]\]\s*/i, '')
				.replace(/^\s+/, '')
				.replace(/\n+$/, '');
			return `- [${checked ? 'x' : ' '}] ${content}\n`;
		}
	});

	turndownService.addRule('fencedCodeBlock', {
		filter: (node) =>
			node.nodeName === 'PRE' && node.firstChild && node.firstChild.nodeName === 'CODE',
		replacement(content, node) {
			// TipTap always appends a trailing \n inside <code> — strip it so it doesn't grow on each save
			const code = node.firstChild.textContent.replace(/\n$/, '');
			const lang = (node.firstChild.getAttribute('class') || '').replace('language-', '');
			const fence = '```';
			return `\n\n${fence}${lang}\n${code}\n${fence}\n\n`;
		}
	});

	// -- Marked: Markdown -> HTML (for initial load) --------------------------

	marked.use({
		breaks: true,
		gfm: true,
		renderer: {
			list(body, ordered, start) {
				if (body.includes('data-checked=')) {
					return `<ul data-type="taskList">${body}</ul>`;
				}
				const type = ordered ? 'ol' : 'ul';
				const startAttr = ordered && start !== 1 ? ` start="${start}"` : '';
				return `<${type}${startAttr}>${body}</${type}>`;
			},
			listitem(text, task, checked) {
				if (task) {
					// Strip the <input> checkbox that marked's GFM inserts — TipTap uses data-checked instead
					text = text.replace(/<input[^>]*type="checkbox"[^>]*>\s*/i, '');
					return `<li data-type="taskItem" data-checked="${checked ? 'true' : 'false'}">${text}</li>`;
				}
				return `<li>${text}</li>`;
			}
		}
	});

	function markdownToHtml(md) {
		if (!md) return '';
		return DOMPurify.sanitize(marked.parse(md), {
			ADD_ATTR: ['data-type', 'data-checked']
		});
	}

	function htmlToMarkdown(html) {
		if (!html || html === '<p></p>') return '';
		return turndownService
			.turndown(html.replace(/<p><\/p>/g, '<br/>'))
			.replace(/\u00a0/g, ' ');
	}

	// -- Editor setup ---------------------------------------------------------

	onMount(() => {
		const content = markdownToHtml(value);

		editor = new Editor({
			element: element,
			extensions: [
				StarterKit.configure({
					heading: { levels: [1, 2, 3, 4] },
					codeBlock: { HTMLAttributes: { class: 'code-block' } }
				}),
				Placeholder.configure({ placeholder }),
				Underline,
				Highlight.configure({ multicolor: false }),
				TaskList,
				TaskItem.configure({ nested: true }),
				Image.configure({ inline: false, allowBase64: true }),
				...(editable
					? [
						BubbleMenu.configure({
							element: bubbleMenuElement,
							tippyOptions: {
								placement: 'bottom-start',
								offset: [0, 4]
							},
							shouldShow: ({ editor: e, view, from, to }) => {
								if (!e || !e.view || e.isDestroyed) return false;
								return view.hasFocus() && from !== to;
							}
						})
					]
					: [])
			],
			content,
			editable,
			onTransaction: () => {
				editor = editor;
			},
			onUpdate: ({ editor: e }) => {
				internalUpdate = true;
				const html = e.getHTML();
				const md = htmlToMarkdown(html);
				value = md;
				dispatch('change', { markdown: md, html });
				internalUpdate = false;
			}
		});
	});

	onDestroy(() => {
		if (editor) {
			editor.destroy();
		}
	});

	function normalizeMd(s) {
		return (s || '')
			.replace(/\r\n/g, '\n')
			.replace(/\n{3,}/g, '\n\n')
			// Normalize trailing newlines inside fenced code blocks
			.replace(/(```\w*\n[\s\S]*?)\n+(```)/g, '$1\n$2')
			.trim();
	}

	$: if (editor && !internalUpdate && value !== undefined) {
		const currentMd = htmlToMarkdown(editor.getHTML());
		if (normalizeMd(value) !== normalizeMd(currentMd)) {
			const html = markdownToHtml(value);
			editor.commands.setContent(html, false);
		}
	}

	$: if (editor) {
		editor.setEditable(editable);
	}
</script>

<!-- Bubble menu: appears when selecting text -->
{#if editable}
<div
	bind:this={bubbleMenuElement}
	class="formatting-menu"
	style="visibility: hidden; position: absolute; z-index: 9999;"
>
	<div class="menu-buttons">
		<button type="button" class:active={editor?.isActive('heading', { level: 1 })}
			on:click={() => editor?.chain().focus().toggleHeading({ level: 1 }).run()} title="Heading 1">
			<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5" class="menu-icon"><path stroke-linecap="round" stroke-linejoin="round" d="M2.243 4.493v7.5m0 0v7.502m0-7.501h10.5m0-7.5v7.5m0 0v7.501m4.501-8.627 2.25-1.5v10.126m0 0h-2.25m2.25 0h2.25" /></svg>
		</button>
		<button type="button" class:active={editor?.isActive('heading', { level: 2 })}
			on:click={() => editor?.chain().focus().toggleHeading({ level: 2 }).run()} title="Heading 2">
			<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5" class="menu-icon"><path stroke-linecap="round" stroke-linejoin="round" d="M21.75 19.5H16.5v-1.609a2.25 2.25 0 0 1 1.244-2.012l2.89-1.445c.651-.326 1.116-.955 1.116-1.683 0-.498-.04-.987-.118-1.463-.135-.825-.835-1.422-1.668-1.489a15.202 15.202 0 0 0-3.464.12M2.243 4.492v7.5m0 0v7.502m0-7.501h10.5m0-7.5v7.5m0 0v7.501" /></svg>
		</button>
		<button type="button" class:active={editor?.isActive('heading', { level: 3 })}
			on:click={() => editor?.chain().focus().toggleHeading({ level: 3 }).run()} title="Heading 3">
			<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5" class="menu-icon"><path stroke-linecap="round" stroke-linejoin="round" d="M20.905 14.626a4.52 4.52 0 0 1 .738 3.603c-.154.695-.794 1.143-1.504 1.208a15.194 15.194 0 0 1-3.639-.104m4.405-4.707a4.52 4.52 0 0 0 .738-3.603c-.154-.696-.794-1.144-1.504-1.209a15.19 15.19 0 0 0-3.639.104m4.405 4.708H18M2.243 4.493v7.5m0 0v7.502m0-7.501h10.5m0-7.5v7.5m0 0v7.501" /></svg>
		</button>
		<button type="button" class:active={editor?.isActive('bulletList')}
			on:click={() => editor?.chain().focus().toggleBulletList().run()} title="Bullet List">
			<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5" class="menu-icon"><path stroke-linecap="round" stroke-linejoin="round" d="M8.25 6.75h12M8.25 12h12m-12 5.25h12M3.75 6.75h.007v.008H3.75V6.75Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0ZM3.75 12h.007v.008H3.75V12Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm-.375 5.25h.007v.008H3.75v-.008Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z" /></svg>
		</button>
		<button type="button" class:active={editor?.isActive('orderedList')}
			on:click={() => editor?.chain().focus().toggleOrderedList().run()} title="Numbered List">
			<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5" class="menu-icon"><path stroke-linecap="round" stroke-linejoin="round" d="M8.242 5.992h12m-12 6.003H20.24m-12 5.999h12M4.117 7.495v-3.75H2.99m1.125 3.75H2.99m1.125 0H5.24m-1.92 2.577a1.125 1.125 0 1 1 1.591 1.59l-1.83 1.83h2.16M2.99 15.745h1.125a1.125 1.125 0 0 1 0 2.25H3.74m0-.002h.375a1.125 1.125 0 0 1 0 2.25H2.99" /></svg>
		</button>
		<button type="button" class:active={editor?.isActive('taskList')}
			on:click={() => editor?.chain().focus().toggleTaskList().run()} title="Task List">
			<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5" class="menu-icon"><path d="M3 20.4V3.6C3 3.26863 3.26863 3 3.6 3H20.4C20.7314 3 21 3.26863 21 3.6V20.4C21 20.7314 20.7314 21 20.4 21H3.6C3.26863 21 3 20.7314 3 20.4Z" stroke-width="1.5" /><path d="M7 12.5L10 15.5L17 8.5" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" /></svg>
		</button>
		<button type="button" class:active={editor?.isActive('bold')}
			on:click={() => editor?.chain().focus().toggleBold().run()} title="Bold">
			<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5" class="menu-icon"><path stroke-linejoin="round" d="M6.75 3.744h-.753v8.25h7.125a4.125 4.125 0 0 0 0-8.25H6.75Zm0 0v.38m0 16.122h6.747a4.5 4.5 0 0 0 0-9.001h-7.5v9h.753Zm0 0v-.37m0-15.751h6a3.75 3.75 0 1 1 0 7.5h-6m0-7.5v7.5m0 0v8.25m0-8.25h6.375a4.125 4.125 0 0 1 0 8.25H6.75m.747-15.38h4.875a3.375 3.375 0 0 1 0 6.75H7.497v-6.75Zm0 7.5h5.25a3.75 3.75 0 0 1 0 7.5h-5.25v-7.5Z" /></svg>
		</button>
		<button type="button" class:active={editor?.isActive('italic')}
			on:click={() => editor?.chain().focus().toggleItalic().run()} title="Italic">
			<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5" class="menu-icon"><path stroke-linecap="round" stroke-linejoin="round" d="M5.248 20.246H9.05m0 0h3.696m-3.696 0 5.893-16.502m0 0h-3.697m3.697 0h3.803" /></svg>
		</button>
		<button type="button" class:active={editor?.isActive('underline')}
			on:click={() => editor?.chain().focus().toggleUnderline().run()} title="Underline">
			<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5" class="menu-icon"><path stroke-linecap="round" stroke-linejoin="round" d="M17.995 3.744v7.5a6 6 0 1 1-12 0v-7.5m-2.25 16.502h16.5" /></svg>
		</button>
		<button type="button" class:active={editor?.isActive('strike')}
			on:click={() => editor?.chain().focus().toggleStrike().run()} title="Strikethrough">
			<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5" class="menu-icon"><path stroke-linecap="round" stroke-linejoin="round" d="M12 12a8.912 8.912 0 0 1-.318-.079c-1.585-.424-2.904-1.247-3.76-2.236-.873-1.009-1.265-2.19-.968-3.301.59-2.2 3.663-3.29 6.863-2.432A8.186 8.186 0 0 1 16.5 5.21M6.42 17.81c.857.99 2.176 1.812 3.761 2.237 3.2.858 6.274-.23 6.863-2.431.233-.868.044-1.779-.465-2.617M3.75 12h16.5" /></svg>
		</button>
		<button type="button" class:active={editor?.isActive('codeBlock')}
			on:click={() => editor?.chain().focus().toggleCodeBlock().run()} title="Code Block">
			<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5" class="menu-icon"><path stroke-linecap="round" stroke-linejoin="round" d="M17.25 6.75 22.5 12l-5.25 5.25m-10.5 0L1.5 12l5.25-5.25m7.5-3-4.5 16.5" /></svg>
		</button>
	</div>
</div>

{/if}

<div class="editor-wrapper" class:editor-editable={editable} class:editor-readonly={!editable} style="--editor-min-height: {minHeight}">
	<div bind:this={element}></div>
</div>

<style>
	/* Floating / Bubble menu pill */
	.formatting-menu {
		pointer-events: auto;
	}

	.menu-buttons {
		display: flex;
		gap: 2px;
		padding: 3px;
		border-radius: 12px;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12), 0 1px 3px rgba(0, 0, 0, 0.08);
		background: var(--rte-menu-bg, #fff);
		color: var(--rte-menu-color, #1d2939);
		border: 1px solid var(--rte-menu-border, #f0f0f0);
		min-width: max-content;
	}

	.menu-buttons button {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 6px;
		border: none;
		background: transparent;
		color: inherit;
		border-radius: 8px;
		cursor: pointer;
		transition: background 0.15s;
	}

	.menu-buttons button:hover {
		background: var(--rte-menu-hover, #f1f5f9);
	}

	.menu-buttons button.active {
		background: var(--rte-menu-active-bg, #e2e8f0);
	}

	:global(.menu-icon) {
		width: 16px;
		height: 16px;
	}

	/* Editor wrapper */
	.editor-wrapper {
		background: var(--rte-bg, #fff);
	}

	.editor-editable {
		border: 1px solid var(--rte-border, #e6e7e9);
		border-radius: var(--rte-radius, 4px);
	}

	.editor-editable:focus-within {
		border-color: var(--rte-focus-border, #0054a6);
		box-shadow: 0 0 0 2px var(--rte-focus-ring, rgba(0, 84, 166, 0.15));
	}

	.editor-readonly {
		border: none;
		background: transparent;
	}

	/* ProseMirror editor */
	:global(.editor-wrapper .ProseMirror) {
		padding: 0.75rem;
		min-height: var(--editor-min-height, 200px);
		outline: none;
		font-size: 0.9375rem;
		line-height: 1.6;
	}

	:global(.editor-readonly .ProseMirror) {
		padding: 0;
		min-height: auto;
	}

	/* Placeholder */
	:global(.editor-wrapper .ProseMirror p.is-editor-empty:first-child::before) {
		content: attr(data-placeholder);
		float: left;
		color: var(--rte-placeholder, #9ca3af);
		pointer-events: none;
		height: 0;
	}

	/* Headings */
	:global(.editor-wrapper .ProseMirror h1) {
		font-size: 1.75rem;
		font-weight: 700;
		margin-top: 1.5rem;
		margin-bottom: 0.75rem;
		line-height: 1.3;
	}

	:global(.editor-wrapper .ProseMirror h2) {
		font-size: 1.375rem;
		font-weight: 600;
		margin-top: 1.25rem;
		margin-bottom: 0.5rem;
		line-height: 1.3;
	}

	:global(.editor-wrapper .ProseMirror h3) {
		font-size: 1.125rem;
		font-weight: 600;
		margin-top: 1rem;
		margin-bottom: 0.5rem;
		line-height: 1.4;
	}

	:global(.editor-wrapper .ProseMirror h4) {
		font-size: 1rem;
		font-weight: 600;
		margin-top: 0.75rem;
		margin-bottom: 0.5rem;
	}

	/* Paragraphs */
	:global(.editor-wrapper .ProseMirror p) {
		margin-bottom: 0.5rem;
	}

	/* Lists */
	:global(.editor-wrapper .ProseMirror ul),
	:global(.editor-wrapper .ProseMirror ol) {
		padding-left: 1.5rem;
		margin-bottom: 0.75rem;
	}

	:global(.editor-wrapper .ProseMirror li) {
		margin-bottom: 0.25rem;
	}

	:global(.editor-wrapper .ProseMirror li p) {
		margin-bottom: 0;
	}

	/* Task list */
	:global(.editor-wrapper .ProseMirror ul[data-type="taskList"]) {
		list-style: none;
		padding-left: 0;
	}

	:global(.editor-wrapper .ProseMirror ul[data-type="taskList"] li) {
		display: flex;
		align-items: flex-start;
		gap: 0.5rem;
	}

	:global(.editor-wrapper .ProseMirror ul[data-type="taskList"] li label) {
		flex: 0 0 auto;
		margin-top: 0.1rem;
	}

	:global(.editor-wrapper .ProseMirror ul[data-type="taskList"] li label input[type="checkbox"]) {
		margin: 0;
		vertical-align: middle;
	}

	:global(.editor-wrapper .ProseMirror ul[data-type="taskList"] li div) {
		flex: 1 1 auto;
	}

	:global(.editor-wrapper .ProseMirror ul[data-type="taskList"] li[data-checked="true"] > div > p) {
		text-decoration: line-through;
		opacity: 0.6;
	}

	/* Inline code */
	:global(.editor-wrapper .ProseMirror code) {
		background: var(--rte-code-bg, #f1f5f9);
		border-radius: 3px;
		padding: 0.15rem 0.4rem;
		font-size: 0.85em;
		font-family: var(--rte-font-mono, 'JetBrains Mono', 'Fira Code', monospace);
	}

	/* Code block */
	:global(.editor-wrapper .ProseMirror pre) {
		background: var(--rte-code-bg, #f1f5f9);
		color: var(--rte-codeblock-color, #1e293b);
		border-radius: var(--rte-radius, 4px);
		padding: 0.75rem 1rem;
		margin-bottom: 0.75rem;
		overflow-x: auto;
	}

	:global(.editor-wrapper .ProseMirror pre code) {
		background: none;
		padding: 0;
		font-size: 0.85rem;
		color: inherit;
	}

	/* Blockquote */
	:global(.editor-wrapper .ProseMirror blockquote) {
		border-left: 3px solid var(--rte-accent, #0054a6);
		padding-left: 1rem;
		margin-left: 0;
		margin-bottom: 0.75rem;
		color: var(--rte-muted, #6c7a91);
	}

	/* Horizontal rule */
	:global(.editor-wrapper .ProseMirror hr) {
		border: none;
		border-top: 1px solid var(--rte-border, #e6e7e9);
		margin: 1.5rem 0;
	}

	/* Images */
	:global(.editor-wrapper .ProseMirror img) {
		max-width: 100%;
		height: auto;
		border-radius: 5px;
		margin: 0.5rem 0;
	}

	/* Table */
	:global(.editor-wrapper .ProseMirror table) {
		width: 100%;
		border-collapse: collapse;
		margin-bottom: 0.75rem;
	}

	:global(.editor-wrapper .ProseMirror th),
	:global(.editor-wrapper .ProseMirror td) {
		border: 1px solid var(--rte-border, #e6e7e9);
		padding: 0.5rem;
		text-align: left;
	}

	:global(.editor-wrapper .ProseMirror th) {
		background: var(--rte-code-bg, #f1f5f9);
		font-weight: 600;
	}

	/* Highlight mark */
	:global(.editor-wrapper .ProseMirror mark) {
		background-color: #fff3bf;
		padding: 0.1rem 0.2rem;
		border-radius: 2px;
	}

	/* Strikethrough */
	:global(.editor-wrapper .ProseMirror s) {
		text-decoration: line-through;
	}
</style>
