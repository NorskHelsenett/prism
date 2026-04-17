<script>
  import {
    notes,
    trashedNotes,
    selectedNoteId,
    searchQuery,
    activeTag,
    showTrash,
    availableTags,
    loadNotes,
    openNote,
    trashNote,
    restoreNote,
    purgeNote,
    emptyTrash,
  } from '$lib/stores/notesStore';

  let debounce;

  function onSearchInput() {
    clearTimeout(debounce);
    debounce = setTimeout(loadNotes, 250);
  }

  function clearSearch() {
    searchQuery.set('');
    clearTimeout(debounce);
    loadNotes();
  }

  function pickTag(tag) {
    activeTag.set($activeTag === tag ? null : tag);
    loadNotes();
  }

  function toggleTrash() {
    showTrash.update((v) => !v);
    loadNotes();
  }

  function formatDate(iso) {
    if (!iso) return '';
    const d = new Date(iso);
    const now = new Date();
    const diffMs = now - d;
    const dayMs = 86400000;
    if (diffMs < dayMs && d.getDate() === now.getDate()) {
      return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    }
    if (diffMs < 7 * dayMs) {
      return d.toLocaleDateString([], { weekday: 'short' });
    }
    return d.toLocaleDateString([], { month: 'short', day: 'numeric' });
  }

  async function confirmPurge(id) {
    if (confirm('Delete this note forever? This cannot be undone.')) {
      await purgeNote(id);
    }
  }

  async function confirmEmpty() {
    if (confirm('Empty the trash? All notes in trash will be permanently deleted.')) {
      await emptyTrash();
    }
  }

  $: list = $showTrash ? $trashedNotes : $notes;
</script>

<aside class="notes-sidebar">
  <div class="sidebar-header">
    <div class="search-row">
      <div class="search-input-wrap">
        <input
          type="search"
          class="form-control form-control-sm"
          placeholder={$showTrash ? 'Search trash…' : 'Search notes…'}
          bind:value={$searchQuery}
          on:input={onSearchInput}
        />
        {#if $searchQuery}
          <button type="button" class="search-clear" on:click={clearSearch} title="Clear search" aria-label="Clear search">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6L6 18" /><path d="M6 6l12 12" /></svg>
          </button>
        {/if}
      </div>
    </div>

    {#if !$showTrash && $availableTags.length > 0}
      <div class="tag-chips">
        {#each $availableTags as tag}
          <button
            type="button"
            class="tag-chip"
            class:active={$activeTag === tag}
            on:click={() => pickTag(tag)}
          >
            #{tag}
          </button>
        {/each}
      </div>
    {/if}

    {#if $showTrash && list.length > 0}
      <button type="button" class="empty-trash-btn" on:click={confirmEmpty}>
        Empty trash
      </button>
    {/if}
  </div>

  <div class="notes-list">
    {#if list.length === 0}
      <div class="empty-state">
        {#if $showTrash}
          Trash is empty.
        {:else if $searchQuery}
          No matching notes.
        {:else}
          No notes yet. Use "New note" above to create one.
        {/if}
      </div>
    {:else}
      {#each list as note (note.id)}
        <button
          type="button"
          class="note-card"
          class:active={$selectedNoteId === note.id}
          on:click={() => $showTrash ? null : openNote(note.id)}
        >
          <div class="note-title">{note.title || 'Untitled'}</div>
          <div class="note-meta">
            <span class="note-date">{formatDate($showTrash ? note.deletedAt : note.updatedAt)}</span>
            <span class="note-preview">{note.preview || ''}</span>
          </div>
          {#if note.tags && note.tags.length > 0}
            <div class="note-tags">
              {#each note.tags.slice(0, 3) as tag}
                <span class="note-tag">#{tag}</span>
              {/each}
            </div>
          {/if}
          <div class="trash-actions">
            {#if $showTrash}
              <button type="button" class="trash-btn" title="Restore" on:click|stopPropagation={() => restoreNote(note.id)}>
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 14l-4-4 4-4" /><path d="M5 10h11a4 4 0 1 1 0 8h-1" /></svg>
              </button>
              <button type="button" class="trash-btn danger" title="Delete forever" on:click|stopPropagation={() => confirmPurge(note.id)}>
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18" /><path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" /><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6" /></svg>
              </button>
            {:else}
              <button type="button" class="trash-btn" title="Move to trash" on:click|stopPropagation={() => trashNote(note.id)}>
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18" /><path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" /><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6" /></svg>
              </button>
            {/if}
          </div>
        </button>
      {/each}
    {/if}
  </div>

  <div class="sidebar-footer">
    <button type="button" class="trash-toggle" class:active={$showTrash} on:click={toggleTrash}>
      <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18" /><path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" /><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6" /></svg>
      <span>{$showTrash ? 'Back to notes' : 'Trash'}</span>
      {#if !$showTrash && $trashedNotes.length > 0}
        <span class="trash-count">{$trashedNotes.length}</span>
      {/if}
    </button>
  </div>
</aside>

<style>
  .notes-sidebar {
    display: flex;
    flex-direction: column;
    width: 320px;
    min-width: 320px;
    border-left: 1px solid var(--tblr-border-color, #e6e7e9);
    background: var(--tblr-bg-surface, #fff);
    height: 100%;
    overflow: hidden;
  }

  .sidebar-header {
    padding: 12px;
  }

  .search-row {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .search-input-wrap {
    position: relative;
    flex: 1 1 auto;
  }

  .search-input-wrap input {
    padding-right: 28px;
  }

  /* Hide the native WebKit clear button so we can render our own. */
  .search-input-wrap input::-webkit-search-cancel-button {
    -webkit-appearance: none;
    appearance: none;
  }

  .search-clear {
    position: absolute;
    top: 50%;
    right: 6px;
    transform: translateY(-50%);
    width: 20px;
    height: 20px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: transparent;
    color: var(--tblr-secondary, #6c757d);
    border-radius: 999px;
    cursor: pointer;
  }
  .search-clear:hover {
    background: var(--tblr-bg-surface-secondary, #f6f8fb);
    color: var(--tblr-body-color, #232e3c);
  }

  .tag-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    margin-top: 8px;
  }

  .tag-chip {
    font-size: 0.75rem;
    padding: 2px 8px;
    border-radius: 999px;
    border: 1px solid var(--tblr-border-color, #e6e7e9);
    background: transparent;
    color: var(--tblr-secondary, #6c757d);
    cursor: pointer;
  }
  .tag-chip.active {
    background: var(--tblr-primary, #206bc4);
    color: white;
    border-color: var(--tblr-primary, #206bc4);
  }

  .empty-trash-btn {
    margin-top: 8px;
    font-size: 0.8125rem;
    padding: 4px 8px;
    background: none;
    border: 1px solid var(--tblr-border-color, #e6e7e9);
    border-radius: 4px;
    color: var(--tblr-danger, #d63939);
    cursor: pointer;
  }

  .notes-list {
    flex: 1 1 auto;
    overflow-y: auto;
    padding: 4px 8px;
  }

  .empty-state {
    padding: 24px 16px;
    text-align: center;
    color: var(--tblr-secondary, #6c757d);
    font-size: 0.875rem;
  }

  .note-card {
    display: block;
    width: 100%;
    padding: 8px 10px;
    margin-bottom: 2px;
    text-align: left;
    background: transparent;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    position: relative;
    transition: background 0.15s;
  }
  .note-card:hover {
    background: var(--tblr-bg-surface-secondary, #f6f8fb);
  }
  .note-card.active {
    background: var(--tblr-bg-surface-secondary, #f6f8fb);
  }
  .note-card.active .note-title {
    color: var(--tblr-primary, #206bc4);
  }

  .note-title {
    font-weight: 600;
    font-size: 0.9375rem;
    color: var(--tblr-body-color, #232e3c);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .note-meta {
    display: flex;
    align-items: baseline;
    gap: 8px;
    margin-top: 2px;
    font-size: 0.8125rem;
    color: var(--tblr-secondary, #6c757d);
  }

  .note-date {
    flex: 0 0 auto;
    font-size: 0.75rem;
    white-space: nowrap;
  }

  .note-preview {
    flex: 1 1 auto;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .note-tags {
    margin-top: 4px;
    display: flex;
    gap: 4px;
    flex-wrap: wrap;
  }

  .note-tag {
    font-size: 0.7rem;
    color: var(--tblr-primary, #206bc4);
  }

  .trash-actions {
    position: absolute;
    top: 8px;
    right: 8px;
    display: flex;
    gap: 4px;
    opacity: 0;
    transition: opacity 0.15s;
  }
  .note-card:hover .trash-actions {
    opacity: 1;
  }
  .trash-btn {
    width: 24px;
    height: 24px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--tblr-border-color, #e6e7e9);
    background: var(--tblr-bg-surface, #fff);
    border-radius: 4px;
    cursor: pointer;
    color: var(--tblr-secondary, #6c757d);
  }
  .trash-btn.danger {
    color: var(--tblr-danger, #d63939);
  }
  .trash-btn:hover {
    background: var(--tblr-bg-surface-secondary, #f6f8fb);
  }

  .sidebar-footer {
    padding: 8px;
  }

  .trash-toggle {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 6px 10px;
    background: transparent;
    border: none;
    border-radius: 6px;
    color: var(--tblr-secondary, #6c757d);
    cursor: pointer;
    font-size: 0.875rem;
    width: 100%;
  }
  .trash-toggle:hover {
    background: var(--tblr-bg-surface-secondary, #f6f8fb);
  }
  .trash-toggle.active {
    color: var(--tblr-primary, #206bc4);
    font-weight: 500;
  }
  .trash-count {
    margin-left: auto;
    background: var(--tblr-secondary-lt, #edeef0);
    color: var(--tblr-secondary, #6c757d);
    padding: 1px 6px;
    border-radius: 999px;
    font-size: 0.75rem;
  }
</style>
