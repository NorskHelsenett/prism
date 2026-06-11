/**
 * Turn a pasted/dropped file into a markdown reference.
 *
 * Images are base64-inlined as data: URIs — they work before the parent
 * vulnerability exists, and the API normaliser (database.
 * MigrateVulnAttachments) extracts them into per-vuln attachment rows on
 * save. Deliberately NOT downscaled or transcoded client-side: the server is
 * authoritative on storage shape and generates its own proxy.
 *
 * Videos and other files are too big to round-trip as base64 through the
 * editor and the vulnerability JSON, so they upload straight to a scoped
 * attachment endpoint (`/api/vulnerability/<id>` for saved vulnerabilities,
 * `/api/drafts/<id>` for drafts) and the markdown references the returned
 * URL. On publish the server rewrites draft URLs to vulnerability URLs.
 *
 * Either way the server sniffs + magic-byte verifies the bytes; nothing
 * declared here is trusted for security.
 */

import { get } from 'svelte/store';
import { apiEndpoint } from '$lib/stores/configStore';

// Video types every evergreen browser can play in a <video> element. Other
// video containers (e.g. video/quicktime) are attached as generic file links
// instead of broken inline players.
const PLAYABLE_VIDEO_MIMES = ['video/mp4', 'video/webm'];

/** @param {Blob} file */
export function isPlayableVideoFile(file) {
  return PLAYABLE_VIDEO_MIMES.includes(file?.type);
}

/**
 * Whether a markdown media src should render as <video> rather than <img>.
 * Scoped attachment URLs carry a cosmetic .mp4/.webm suffix exactly so this
 * works without a metadata round-trip.
 * @param {string} src
 */
export function isVideoSource(src) {
  return /^data:video\//i.test(src || '') || /\.(mp4|webm)([?#]|$)/i.test(src || '');
}

/**
 * Display src for a <video> element. With preload="metadata" browsers
 * (Safari in particular) paint no poster frame — the player sits grey until
 * play. The #t=0.001 media fragment makes them seek and paint the first
 * frame while still only fetching a chunk. Display-only: never write this
 * back into markdown or node attrs.
 * @param {string} src
 */
export function videoPreviewSrc(src) {
  if (!src || src.includes('#') || src.startsWith('data:')) return src;
  return `${src}#t=0.001`;
}

/**
 * Upload a file as a scoped attachment under the given owner base
 * (`/api/vulnerability/<id>` or `/api/drafts/<id>`).
 * Uses raw fetch (not the Fetch util) on purpose: the util navigates the
 * whole page to /404 on a 404 response, which must never happen mid-edit.
 * @param {string} base - e.g. `/api/drafts/7`
 * @param {Blob} file
 * @param {string} [name]
 * @returns {Promise<{key: string, url: string, filename: string, mime: string, kind: string}>}
 */
export async function uploadAttachment(base, file, name) {
  const form = new FormData();
  form.append('file', file, name || file.name || 'attachment');
  const response = await fetch(`${get(apiEndpoint)}${base}/attachments`, {
    method: 'POST',
    credentials: 'include',
    body: form
  });
  if (!response.ok) {
    let message = 'upload failed';
    try {
      message = (await response.json())?.error || message;
    } catch {}
    throw new Error(message);
  }
  return await response.json();
}

/**
 * Markdown reference for an uploaded attachment summary: media renders
 * inline (`![name](url)`), everything else becomes a link (`[name](url)`).
 * @param {{url: string, filename?: string, kind?: string}} summary
 */
export function attachmentSummaryToMarkdown(summary) {
  const name = String(summary.filename || 'attachment').replace(/[\[\]]/g, '_');
  if (summary.kind === 'image' || summary.kind === 'video') {
    return `![${name}](${summary.url})`;
  }
  return `[${name}](${summary.url})`;
}

/**
 * Delete a scoped attachment under the given owner base. Raw fetch for the
 * same reason as uploadAttachment. Returns whether the server confirmed it.
 * @param {string} base - e.g. `/api/drafts/7`
 * @param {string} key
 */
export async function deleteAttachment(base, key) {
  const response = await fetch(`${get(apiEndpoint)}${base}/attachments/${key}`, {
    method: 'DELETE',
    credentials: 'include'
  });
  return response.ok;
}

/**
 * Markdown for any pasted/dropped file: images inline as data URIs, videos
 * and other files upload to the scoped attachment endpoint when uploadBase
 * is given. Returns null when the file needs an upload but no base could be
 * resolved (caller decides the messaging).
 * @param {Blob} file
 * @param {{uploadBase?: string|null, name?: string}} [options]
 * @returns {Promise<string|null>}
 */
export async function fileToMarkdown(file, { uploadBase, name } = {}) {
  const displayName = name || file.name || 'file';
  if (file.type?.startsWith('image/')) {
    return fileToMarkdownImage(file, displayName);
  }
  if (!uploadBase) return null;
  const summary = await uploadAttachment(uploadBase, file, displayName);
  return attachmentSummaryToMarkdown(summary);
}

/**
 * @param {Blob} file
 * @param {string} alt - alt text for the image (typically the original filename)
 * @returns {Promise<string>} markdown like `![alt](data:image/<mime>;base64,...)`
 */
export async function fileToMarkdownImage(file, alt = 'image') {
  const dataUri = await blobToDataURI(file);
  const safeAlt = String(alt).replace(/[\[\]]/g, '_');
  return `![${safeAlt}](${dataUri})`;
}

function blobToDataURI(blob) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result));
    reader.onerror = () => reject(reader.error);
    reader.readAsDataURL(blob);
  });
}
