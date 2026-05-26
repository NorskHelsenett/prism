/**
 * Base64-inline a pasted/dropped file into markdown image syntax.
 *
 * Deliberately does NOT downscale or transcode. The server is authoritative
 * on storage shape — it stores the original bytes untouched and generates
 * its own proxy at the resolution configured in admin settings. Letting the
 * browser fiddle with pixels here only introduced quality bugs (1080p cap
 * making screenshot text unreadable) without buying any real protection.
 *
 * Used by every editor that supports pasting/dropping images into markdown.
 * On save, the API normaliser (database.MigrateVulnAttachments) extracts the
 * data: URI into a per-vuln attachment row.
 */

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
