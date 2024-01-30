<script>
	import { accessLevels } from '$lib/userStore';


import DOMPurify from 'dompurify'
import { marked } from 'marked';
import { createEventDispatcher } from 'svelte';
import { onMount } from 'svelte';

const dispatch = createEventDispatcher();

let container;
export let markdown = ""

const renderer = new marked.Renderer();
let todo = 0
let checked = 0

export let writeAccess = false

renderer.listitem = function(text) {
  if (text.includes("type=\"checkbox\"") == false){
    return "<li>" + text + "</li>"
  }
  // Assuming the text is in the format: '[ ] Item text' or '[x] Item text'
  const isChecked = text.includes('checked');
  if (isChecked) {
    checked++
  }
  const isDisabled = !writeAccess; // Modify as needed
  let itemText = text.replace(/^\[\s?x?\]\s?/, '');
  itemText = itemText.replace(/<input [^>]*type="checkbox"[^>]*>\s*/, '');

  return `
    <label class="form-check" data-index="${todo++}">
      <input class="form-check-input" type="checkbox" ${isChecked ? 'checked' : ''} ${isDisabled ? 'disabled' : ''}>
      <span class="form-check-label">${itemText}</span>
    </label>
  `;
};

let renderedMarkdown = ""

 $: {
    checked = 0;
    todo = 0; // Reset todo before each rendering
    renderedMarkdown = DOMPurify.sanitize(marked.parse(markdown, { renderer }));
  }

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
<div class="" bind:this={container}>{@html renderedMarkdown || 'N/A'}</div>

<style>
  :global(.form-check-input:disabled){
    opacity: 1 !important;
  }
  :global(form-check-input:disabled ~ .form-check-label, .form-check-input[disabled] ~ .form-check-label) {
    opacity: 1 !important;
  }
</style>