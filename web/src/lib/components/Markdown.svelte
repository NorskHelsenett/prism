<script>
  import { run } from 'svelte/legacy';

import DOMPurify from 'dompurify'
import { marked } from 'marked';
import { createEventDispatcher } from 'svelte';
import { onMount } from 'svelte';

const dispatch = createEventDispatcher();

let container = $state();

const renderer = $state(new marked.Renderer());
let todo = $state(0)
let checked = $state(0)

  /**
   * @typedef {Object} Props
   * @property {string} [markdown]
   * @property {boolean} [writeAccess]
   */

  /** @type {Props} */
  let { markdown = $bindable(""), writeAccess = false } = $props();

// marked v5+ passes token objects (not strings) to renderer methods.
// Task-list checkboxes are emitted as a separate `checkbox` token inside the
// list item's tokens; suppress the built-in one so our custom input is the only one.
renderer.checkbox = function() {
  return '';
};

renderer.listitem = function(item) {
  const body = this.parser.parse(item.tokens);
  if (!item.task) {
    return `<li>${body}</li>`;
  }
  if (item.checked) {
    checked++;
  }
  const isDisabled = !writeAccess; // Modify as needed
  return `
    <label class="form-check" data-index="${todo++}">
      <input class="form-check-input" type="checkbox" ${item.checked ? 'checked' : ''} ${isDisabled ? 'disabled' : ''}>
      <span class="form-check-label">${body.trim()}</span>
    </label>
  `;
};

// Add a custom renderer for links
renderer.link = function({ href, title, tokens }) {
  const text = this.parser.parseInline(tokens);
  // Ensure the href has the 'https://' prefix
  if (!href.startsWith('http://') && !href.startsWith('https://')) {
    href = 'https://' + href;
  }
  const target = '_blank'; // Open link in a new tab
  const rel = 'noopener noreferrer'; // Security attributes
  const titleAttr = title ? `title="${title}"` : '';
  return `<a href="${href}" target="${target}" rel="${rel}" ${titleAttr}>${text}</a>`;
};

// Configure DOMPurify to allow 'target' and 'rel' attributes on 'a' tags
const domPurifyConfig = {
  ALLOWED_TAGS: [
    'a', 'b', 'i', 'em', 'strong', 'p', 'ul', 'ol', 'li',
    'table', 'thead', 'tbody', 'tr', 'th', 'td', 'img',
    'label', 'input', 'span', 'div', 'hr', 'h1', 'h2',
    'h3', 'h4', 'h5', 'h6', 'pre', 'code', 'blockquote'
  ],
  ALLOWED_ATTR: [
    'href', 'title', 'target', 'rel', 'class', 'src',
    'alt', 'data-index', 'type', 'checked', 'disabled'
  ]
};

let renderedMarkdown = $state("")

 run(() => {
    checked = 0;
    todo = 0; // Reset todo before each rendering
    renderedMarkdown = DOMPurify.sanitize(marked.parse(markdown, { renderer }), domPurifyConfig);
  });

onMount(() => {
  if (!writeAccess) { return }
    container.addEventListener('click', (event) => {
      const label = event.target.closest('.form-check');
      if (label) {
        const index = label.getAttribute('data-index');
        updateText(index);
      }
    });
  });

function updateText(index) {
  const lines = markdown.split('\n');
  let checkboxCount = 0;

  for (let i = 0; i < lines.length; i++) {
    // Check if the line is a checkbox item
    if (/\[([\sxX])\]/.test(lines[i])) {
      if (checkboxCount == index) {
        // Toggle the checkbox state
        lines[i] = lines[i].replace(/\[([\sxX])\]/, (_, checkboxState) => {
          return checkboxState.trim() === 'x' || checkboxState.trim() === 'X' ? '[ ]' : '[x]';
        });
        break;
      }
      checkboxCount++;
    }
  }

  markdown = lines.join('\n');
  dispatch('markdownChanged', { updatedMarkdown: markdown });
}
</script>

<div>
{#if todo > 0 && todo != checked}
<span class="badge text-azure mb-3"><svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-checklist" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M9.615 20h-2.615a2 2 0 0 1 -2 -2v-12a2 2 0 0 1 2 -2h8a2 2 0 0 1 2 2v8" /><path d="M14 19l2 2l4 -4" /><path d="M9 8h4" /><path d="M9 12h2" /></svg>
  {checked}/{todo} tasks done
</span>
{:else if todo > 0}
<span class="badge text-green mb-3">
  <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-discount-check" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M5 7.2a2.2 2.2 0 0 1 2.2 -2.2h1a2.2 2.2 0 0 0 1.55 -.64l.7 -.7a2.2 2.2 0 0 1 3.12 0l.7 .7c.412 .41 .97 .64 1.55 .64h1a2.2 2.2 0 0 1 2.2 2.2v1c0 .58 .23 1.138 .64 1.55l.7 .7a2.2 2.2 0 0 1 0 3.12l-.7 .7a2.2 2.2 0 0 0 -.64 1.55v1a2.2 2.2 0 0 1 -2.2 2.2h-1a2.2 2.2 0 0 0 -1.55 .64l-.7 .7a2.2 2.2 0 0 1 -3.12 0l-.7 -.7a2.2 2.2 0 0 0 -1.55 -.64h-1a2.2 2.2 0 0 1 -2.2 -2.2v-1a2.2 2.2 0 0 0 -.64 -1.55l-.7 -.7a2.2 2.2 0 0 1 0 -3.12l.7 -.7a2.2 2.2 0 0 0 .64 -1.55v-1" /><path d="M9 12l2 2l4 -4" /></svg>
  {checked}/{todo} all done
</span>
{/if}
</div>
<div class="markdown-content" bind:this={container}>{@html renderedMarkdown || ''}</div>

<style>
  :global(.form-check-input:disabled){
    opacity: 1 !important;
  }
  :global(form-check-input:disabled ~ .form-check-label, .form-check-input[disabled] ~ .form-check-label) {
    opacity: 1 !important;
  }

  :global(.markdown-content > *){
    margin-bottom: 1.5rem!important;
  }

  :global(.markdown-content > hr){
    margin-bottom: 1.2rem!important;
    margin-top: -0.5rem!important;
  }

  :global(td,th){
    padding: 0 1rem 0.4rem 0;
  }

  :global(table) {
    width: 100%;
    background-color: var(--tblr-table-bg);
  }

  :global(tr){
    padding: 1rem 1rem;
  }

  :global(td){
    vertical-align: top;
  }

  :global(.markdown-content img){
    border-radius: 5px;
    display: block;
    margin-left: auto;
    margin-right: auto;
    max-width: 100%;
  }
</style>
