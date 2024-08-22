
import {baseUrl} from '$lib/stores';
export const ssr = false;
export async function load() {
  try {
    sessionStorage.removeItem('vulnerabilities');

    const response = await fetch('/.well-known/config.json');
    if (!response.ok) {
      throw new Error('Failed to fetch API endpoint');
    }
    const config = await response.json();
    baseUrl.set(config.apiEndpoint);
  } catch (error) {
    console.error('Could not fetch API endpoint:', error);
  }
}