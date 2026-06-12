/**
 * Client for server-side vulnerability drafts. Drafts replace the old
 * localStorage draft: they survive browser changes and give attachments
 * (videos, files) a parent to upload against before the vulnerability
 * itself exists. On publish the server creates the vulnerability, claims
 * the draft's attachments, rewrites attachment URLs in the markdown, and
 * consumes the draft.
 *
 * Raw fetch (not the Fetch util) on purpose: the util navigates the whole
 * page to /404 or /login on error responses, which must never happen from a
 * background autosave.
 */

import { get } from 'svelte/store';
import { apiEndpoint } from '$lib/stores/configStore';

async function draftFetch(path, options = {}) {
  const response = await fetch(`${get(apiEndpoint)}${path}`, {
    credentials: 'include',
    ...options
  });
  if (!response.ok) {
    let message = `draft request failed (${response.status})`;
    try {
      message = (await response.json())?.error || message;
    } catch {}
    throw new Error(message);
  }
  if (response.status === 204) return null;
  return await response.json();
}

/**
 * @param {{vulnerability?: object, projectID?: number}} [payload]
 * @returns {Promise<number>} the new draft id
 */
export async function createDraft(payload = {}) {
  const created = await draftFetch('/api/drafts', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  return created.id;
}

/** Newest-first summaries: [{id, projectID, createdAt, updatedAt}] */
export async function listDrafts() {
  return await draftFetch('/api/drafts');
}

/** Full draft: {id, vulnerability, projectID, createdAt, updatedAt} */
export async function getDraft(id) {
  return await draftFetch(`/api/drafts/${id}`);
}

export async function updateDraft(id, payload) {
  return await draftFetch(`/api/drafts/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
}

export async function deleteDraft(id) {
  return await draftFetch(`/api/drafts/${id}`, { method: 'DELETE' });
}

/** @returns {Promise<number>} the published vulnerability id */
export async function publishDraft(id) {
  const published = await draftFetch(`/api/drafts/${id}/publish`, { method: 'POST' });
  return published.id;
}
