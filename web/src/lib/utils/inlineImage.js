/**
 * Client-side image downscale + base64 inlining helpers.
 *
 * Used by every editor that supports pasting/dropping images into markdown.
 * The bytes ride along inside the markdown as a data: URI; on save the API
 * normaliser (database.MigrateVulnAttachments) extracts them into per-vuln
 * attachment rows. The editor itself never talks to the attachment endpoint.
 */

const DEFAULT_MAX_EDGE = 1080;
const DEFAULT_WEBP_QUALITY = 0.85;
const DEFAULT_JPEG_QUALITY = 0.85;

/**
 * Decode a file (Blob or File) into a downscaled data: URI.
 *
 * @param {Blob} file
 * @param {number} maxEdge - maximum dimension on the long edge in pixels
 * @returns {Promise<string>} data: URI suitable for use in a markdown image
 */
export async function resizeToDataUri(file, maxEdge = DEFAULT_MAX_EDGE) {
  const bitmap = await createImageBitmap(file).catch(() => null);
  if (!bitmap) {
    // Couldn't decode (animated/proprietary format). Fall back to embedding
    // the raw bytes so something shows up; the server will reject if it's
    // outside the allowed MIME list, but a stale upload UX is preferable to
    // a silent drop.
    return await blobToDataURI(file);
  }

  const { width, height } = bitmap;
  const longEdge = Math.max(width, height);
  const scale = longEdge > maxEdge ? maxEdge / longEdge : 1;
  const w = Math.max(1, Math.round(width * scale));
  const h = Math.max(1, Math.round(height * scale));

  const canvas = document.createElement('canvas');
  canvas.width = w;
  canvas.height = h;
  const ctx = canvas.getContext('2d');
  ctx.drawImage(bitmap, 0, 0, w, h);
  bitmap.close?.();

  return await canvasToDataURI(canvas);
}

function canvasToDataURI(canvas) {
  // Prefer WebP for size; fall back to JPEG if the browser can't encode WebP.
  return new Promise((resolve) => {
    canvas.toBlob(
      (blob) => {
        if (blob) {
          blobToDataURI(blob).then(resolve);
        } else {
          // WebP unsupported — try JPEG.
          canvas.toBlob(
            (jpgBlob) => {
              if (jpgBlob) {
                blobToDataURI(jpgBlob).then(resolve);
              } else {
                resolve(canvas.toDataURL('image/jpeg', DEFAULT_JPEG_QUALITY));
              }
            },
            'image/jpeg',
            DEFAULT_JPEG_QUALITY,
          );
        }
      },
      'image/webp',
      DEFAULT_WEBP_QUALITY,
    );
  });
}

function blobToDataURI(blob) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result));
    reader.onerror = () => reject(reader.error);
    reader.readAsDataURL(blob);
  });
}

/**
 * Decode + downscale a file and build a markdown image reference for insertion.
 *
 * @param {Blob} file
 * @param {string} alt - alt text for the image (typically the original filename)
 * @returns {Promise<string>} markdown like `![alt](data:image/webp;base64,...)`
 */
export async function fileToMarkdownImage(file, alt = 'image') {
  const dataUri = await resizeToDataUri(file);
  const safeAlt = String(alt).replace(/[\[\]]/g, '_');
  return `![${safeAlt}](${dataUri})`;
}
