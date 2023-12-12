// src/lib/config.js
export async function getApiEndpoint() {
  const response = await fetch('/.well-known/config.json');
  if (!response.ok) {
    throw new Error('Failed to fetch API endpoint');
  }
  const config = await response.json();
  return config.apiEndpoint;
}
