import { writable, get } from 'svelte/store';
import { Fetch } from '$lib/fetchUtil';
import { apiEndpoint } from '$lib/stores/configStore';

// Unique id for this tab. Echoed back in SSE `source` so we ignore our own
// updates (otherwise an autosave would stomp the active editor buffer).
export const tabId = `tab_${Math.random().toString(36).slice(2, 10)}`;

// Individual stores rather than one mega-object — simpler subscriptions.
export const notes = writable([]);
export const trashedNotes = writable([]);
export const selectedNoteId = writable(null);
export const selectedNote = writable(null);
export const searchQuery = writable('');
export const activeTag = writable(null);
export const showTrash = writable(false);
export const availableTags = writable([]);

// 'idle' | 'saving' | 'saved' | 'error'
export const saveState = writable('idle');

let saveTimer = null;
let saveStateTimer = null;
let pendingContent = null;
let pendingId = null;
let sse = null;

const SAVE_DEBOUNCE_MS = 800;

async function fetchTags() {
  const tags = await Fetch('/api/notes/tags');
  if (Array.isArray(tags)) availableTags.set(tags);
}

export async function loadNotes() {
  const trash = get(showTrash);
  const params = new URLSearchParams();
  const q = get(searchQuery).trim();
  const tag = get(activeTag);
  if (q) params.set('q', q);
  if (tag) params.set('tag', tag);
  if (trash) params.set('trash', 'true');
  const list = await Fetch(`/api/notes${params.toString() ? `?${params}` : ''}`);
  if (Array.isArray(list)) {
    if (trash) trashedNotes.set(list);
    else notes.set(list);
  }
  fetchTags();
}

export async function openNote(id) {
  if (!id) {
    selectedNoteId.set(null);
    selectedNote.set(null);
    return;
  }
  // Cancel any pending autosave for the previous note so we don't stomp it
  // after switching.
  await flushSave();
  selectedNoteId.set(id);
  const note = await Fetch(`/api/notes/${id}`);
  if (note && !note.error) {
    selectedNote.set(note);
  } else {
    selectedNote.set(null);
    selectedNoteId.set(null);
  }
}

export async function createNote() {
  await flushSave();
  const note = await Fetch('/api/notes', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ content: '', source: tabId }),
  });
  if (note && !note.error) {
    selectedNoteId.set(note.id);
    selectedNote.set(note);
    // Prepend to the sidebar list optimistically.
    notes.update((list) => [
      {
        id: note.id,
        title: note.title || '',
        preview: note.preview || '',
        tags: [],
        updatedAt: note.updatedAt,
      },
      ...list,
    ]);
    return note;
  }
  return null;
}

export function queueSave(id, content) {
  if (!id) return;
  pendingId = id;
  pendingContent = content;
  saveState.set('saving');
  if (saveTimer) clearTimeout(saveTimer);
  saveTimer = setTimeout(() => {
    flushSave();
  }, SAVE_DEBOUNCE_MS);
}

export async function flushSave() {
  if (saveTimer) {
    clearTimeout(saveTimer);
    saveTimer = null;
  }
  if (pendingId == null || pendingContent == null) return;
  const id = pendingId;
  const content = pendingContent;
  pendingId = null;
  pendingContent = null;

  const result = await Fetch(`/api/notes/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ content, source: tabId }),
  });
  if (!result || result.error) {
    saveState.set('error');
    return;
  }
  saveState.set('saved');
  if (saveStateTimer) clearTimeout(saveStateTimer);
  saveStateTimer = setTimeout(() => saveState.set('idle'), 1200);

  // Update sidebar preview/title without refetching the whole list.
  const patch = {
    id: result.id,
    title: result.title,
    preview: result.preview,
    tags: [],
    updatedAt: result.updatedAt,
  };
  notes.update((list) => {
    const idx = list.findIndex((n) => n.id === id);
    if (idx === -1) return [patch, ...list];
    const copy = list.slice();
    copy.splice(idx, 1);
    return [patch, ...copy];
  });
  fetchTags();
}

export async function trashNote(id) {
  await flushSave();
  const result = await Fetch(`/api/notes/${id}?source=${tabId}`, { method: 'DELETE' });
  if (!result || result.error) return false;
  notes.update((list) => list.filter((n) => n.id !== id));
  if (get(selectedNoteId) === id) {
    selectedNoteId.set(null);
    selectedNote.set(null);
  }
  return true;
}

export async function restoreNote(id) {
  const result = await Fetch(`/api/notes/${id}/restore`, { method: 'POST' });
  if (!result || result.error) return false;
  trashedNotes.update((list) => list.filter((n) => n.id !== id));
  await loadNotes();
  return true;
}

export async function purgeNote(id) {
  const result = await Fetch(`/api/notes/${id}/permanent`, { method: 'DELETE' });
  if (!result || result.error) return false;
  trashedNotes.update((list) => list.filter((n) => n.id !== id));
  return true;
}

export async function emptyTrash() {
  const result = await Fetch('/api/notes/trash', { method: 'DELETE' });
  if (!result || result.error) return false;
  trashedNotes.set([]);
  return true;
}

// -- SSE ----------------------------------------------------------------

export function subscribeToStream() {
  const endpoint = get(apiEndpoint);
  if (!endpoint) return;
  if (sse) return;
  try {
    sse = new EventSource(`${endpoint}/api/notes/events`, { withCredentials: true });
  } catch (e) {
    return;
  }

  const handle = async (type, data) => {
    if (data?.source === tabId) return; // Ignore echoes of our own writes.
    switch (type) {
      case 'note.updated': {
        const currentId = get(selectedNoteId);
        if (currentId === data.id) {
          // Another device edited the note we're on — refresh silently only
          // if our local buffer matches server state; otherwise show a
          // banner (handled in the page via `remoteUpdatedConflict`).
          remoteUpdatedConflict.set({ id: data.id, at: data.updatedAt });
        }
        await loadNotes();
        break;
      }
      case 'note.created':
      case 'note.restored':
      case 'note.deleted':
      case 'note.purged':
        await loadNotes();
        break;
    }
  };

  for (const t of ['note.updated', 'note.created', 'note.deleted', 'note.restored', 'note.purged']) {
    sse.addEventListener(t, (e) => {
      try { handle(t, JSON.parse(e.data)); } catch {}
    });
  }
  sse.onerror = () => {
    // EventSource auto-reconnects; just let it.
  };
}

export function unsubscribeFromStream() {
  if (sse) {
    sse.close();
    sse = null;
  }
}

export const remoteUpdatedConflict = writable(null);
