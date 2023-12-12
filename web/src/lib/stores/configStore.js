import { writable } from 'svelte/store';

export const apiEndpoint = writable('');
export const isLoading = writable(true);

export async function initializeApiEndpoint() {
  isLoading.set(true);
  try {
    const response = await fetch('/.well-known/config.json');
    if (response.ok) {
      const config = await response.json();
      apiEndpoint.set(config.apiEndpoint);
    } else {
      console.error('Failed to load API endpoint config:', response.status);
    }
  } catch (error) {
    console.error('Error fetching API endpoint config:', error);
  } finally {
    isLoading.set(false);
  }
}
