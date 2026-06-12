<script>
  import { run } from 'svelte/legacy';

  import { onMount, onDestroy, tick } from 'svelte';
  import RichTextEditor from '$lib/components/editor/RichTextEditor.svelte';
  import NotesSidebar from '$lib/components/notes/NotesSidebar.svelte';
  import SavePill from '$lib/components/notes/SavePill.svelte';
  import { TagHighlight } from '$lib/components/notes/TagHighlightExtension.js';
  import {
    selectedNoteId,
    selectedNote,
    loadNotes,
    openNote,
    createNote,
    queueSave,
    flushSave,
    subscribeToStream,
    unsubscribeFromStream,
    remoteUpdatedConflict,
  } from '$lib/stores/notesStore';
  import { Fetch } from '$lib/fetchUtil';
  import { toast } from 'svelte-sonner';
  import { fileToMarkdownImage } from '$lib/utils/inlineImage';

  const pretitle = 'Personal';
  const title = 'Notes';

  let editorRef = $state();
  let editorContent = $state('');
  let lastLoadedId = $state(null);
  // Snapshot of the content we last loaded from the server, used to suppress
  // "saving…" pill flashes when the editor roundtrips markdown->html->markdown
  // on open without any real user edit.
  let lastLoadedContent = $state('');

  function normalizeForCompare(s) {
    return (s || '')
      .replace(/\r\n/g, '\n')
      .replace(/\n{3,}/g, '\n\n')
      .trim();
  }

  run(() => {
    if ($selectedNote && $selectedNote.id !== lastLoadedId) {
      lastLoadedId = $selectedNote.id;
      lastLoadedContent = $selectedNote.content || '';
      editorContent = lastLoadedContent;
      tick().then(() => editorRef?.focusEditor(lastLoadedContent ? 'end' : 'start'));
    } else if (!$selectedNote && lastLoadedId !== null) {
      lastLoadedId = null;
      lastLoadedContent = '';
      editorContent = '';
    }
  });

  // Clicking anywhere in the editor pane (including the empty padding area
  // outside the ProseMirror element) should focus the editor so the user
  // can start typing immediately.
  function handlePaneClick(event) {
    if (!editorRef || event.target.closest('.ProseMirror')) return;
    editorRef.focusEditor('end');
  }

  function handleEditorChange(event) {
    const md = event.detail.markdown ?? '';
    editorContent = md;
    if (!$selectedNoteId) return;
    // Ignore no-op change events fired while opening a note.
    if (normalizeForCompare(md) === normalizeForCompare(lastLoadedContent)) return;
    queueSave($selectedNoteId, md);
  }

  async function handleFilePaste(event) {
    const blob = event.detail?.blob;
    if (!blob) return;
    await uploadAndInsert(blob, null);
  }

  async function handleFileDrop(event) {
    const { files, pos } = event.detail || {};
    if (!files?.length) return;
    for (const file of files) {
      await uploadAndInsert(file, pos);
    }
  }

  async function uploadAndInsert(file, pos) {
    // Notes have no attachment store, so only images (inlined as data URIs
    // in the note content) are supported here.
    if (!file.type?.startsWith('image/')) {
      toast.error('Only images can be embedded in notes');
      return;
    }
    const link = await fileToMarkdownImage(file, file.name || 'image');
    await tick();
    editorRef?.insertAttachment?.(link, pos ?? null);
  }

  async function reloadCurrentNote() {
    const id = $selectedNoteId;
    if (!id) return;
    const note = await Fetch(`/api/notes/${id}`);
    if (note && !note.error) {
      selectedNote.set(note);
      editorContent = note.content || '';
      lastLoadedId = id;
    }
    remoteUpdatedConflict.set(null);
  }

  function dismissConflict() {
    remoteUpdatedConflict.set(null);
  }

  onMount(async () => {
    await loadNotes();
    subscribeToStream();
  });

  onDestroy(() => {
    flushSave();
    unsubscribeFromStream();
  });
</script>

<svelte:head>
  <title>Notes · PRISM</title>
</svelte:head>

<div class="row g-2 align-items-center">
  <div class="col">
    <div class="page-pretitle">{pretitle}</div>
    <h2 class="page-title">{title}</h2>
  </div>
  <div class="col-auto ms-auto d-print-non">
    <div class="btn-list">
      <button type="button" class="btn btn-primary d-none d-sm-inline-block" onclick={createNote}>
        <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
          <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
          <path d="M12 5l0 14" />
          <path d="M5 12l14 0" />
        </svg>
        New note
      </button>
    </div>
  </div>
</div>

<div class="page-body">
  <div class="notes-app card">
    <div class="editor-pane" onclick={handlePaneClick} role="presentation">
      {#if $selectedNote}
        {#if $remoteUpdatedConflict && $remoteUpdatedConflict.id === $selectedNoteId}
          <div class="conflict-banner">
            This note was updated on another device.
            <button type="button" class="link-btn" onclick={reloadCurrentNote}>Reload</button>
            <button type="button" class="link-btn dismiss" onclick={dismissConflict}>Dismiss</button>
          </div>
        {/if}

        <div class="editor-container">
          <RichTextEditor
            bind:this={editorRef}
            bind:value={editorContent}
            placeholder="Start writing…  use #tag anywhere to tag this note"
            minHeight="100%"
            extraExtensions={[TagHighlight]}
            on:change={handleEditorChange}
            on:filepaste={handleFilePaste}
            on:filedrop={handleFileDrop}
          />
        </div>
      {:else}
        <div class="empty-editor">
          <p>Select a note on the right, or create a new one.</p>
          <button type="button" class="btn btn-primary" onclick={createNote}>New note</button>
        </div>
      {/if}
    </div>

    <NotesSidebar />
  </div>
</div>

<SavePill />

<style>
  .notes-app {
    display: flex;
    flex-direction: row-reverse;
    height: calc(100vh - 220px);
    min-height: 480px;
    padding: 0;
    overflow: hidden;
  }

  .editor-pane {
    flex: 1 1 auto;
    min-width: 0;
    overflow-y: auto;
    position: relative;
    padding: 24px clamp(20px, 6vw, 80px) 48px;
    background: transparent;
    cursor: text;
  }

  .editor-container {
    max-width: 820px;
    margin: 0 auto;
    min-height: 100%;
    cursor: text;
  }

  .empty-editor {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: 12px;
    color: var(--tblr-secondary, #6c757d);
  }

  .conflict-banner {
    max-width: 820px;
    margin: 0 auto 16px;
    padding: 8px 12px;
    border-radius: 6px;
    background: var(--tblr-warning-lt, #fff4e0);
    color: var(--tblr-warning, #f59f00);
    font-size: 0.875rem;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .link-btn {
    background: none;
    border: none;
    color: inherit;
    text-decoration: underline;
    cursor: pointer;
    padding: 0;
  }

  .link-btn.dismiss {
    margin-left: auto;
    opacity: 0.7;
  }

  :global(.note-tag-highlight) {
    color: var(--tblr-primary, #206bc4);
    font-weight: 500;
  }
</style>
