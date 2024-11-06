import { baseUrl } from '$lib/stores';

export const ssr = false;

/** @type {import('@sveltejs/kit').Load} */
export async function load({ fetch }) {  // Add fetch parameter here
  try {
    sessionStorage.removeItem('vulnerabilities');

    const response = await fetch('/.well-known/config.json');
    if (!response.ok) {
      throw new Error('Failed to fetch API endpoint');
    }
    const config = await response.json();
    baseUrl.set(config.apiEndpoint);

    return {
      config  // Return the loaded data
    };
  } catch (error) {
    console.error('Could not fetch API endpoint:', error);
    return {
      error: 'Failed to load configuration'
    };
  }
}