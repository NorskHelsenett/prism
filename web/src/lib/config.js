// src/lib/config.js
import { dev } from '$app/environment';

/** @param {string} apiEndpoint */
export function normalizeApiEndpoint(apiEndpoint) {
  if (dev) return '';
  return apiEndpoint;
}

export async function getApiEndpoint() {
  const response = await fetch('/.well-known/config.json');
  if (!response.ok) {
    throw new Error('Failed to fetch API endpoint');
  }
  const config = await response.json();
  return normalizeApiEndpoint(config.apiEndpoint);
}
