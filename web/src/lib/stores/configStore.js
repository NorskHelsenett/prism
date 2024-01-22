import { goto } from '$app/navigation';
import { writable } from 'svelte/store';

export const apiEndpoint = writable('');
export const isLoading = writable(true);
export const isAuthenticated = writable(false);

export async function initializeApiEndpoint() {
  isLoading.set(true);
  try {
    const response = await fetch('/.well-known/config.json');
    if (response.ok) {
      const config = await response.json();
      apiEndpoint.set(config.apiEndpoint);

      const url = `${config.apiEndpoint}/api/user`;
      const responseAuth = await fetch(url, { credentials: 'include' });

      if (responseAuth.ok) {
        isAuthenticated.set(true)
        isLoading.set(false);
      }

      if (responseAuth.status === 401) {
        isLoading.set(false);
        goto("/login")
      }

      if (responseAuth.status === 403) {
        const errormessage = await responseAuth.json()

        if (errormessage?.initiateOTP === true) {
          isLoading.set(false);
          goto("/auth")
        }
      }

    } else {
      console.error('Failed to load API endpoint config:', response.status);
    }
  } catch (error) {
    console.error('Error fetching API endpoint config:', error);
  } finally {
    isLoading.set(false);
  }
}
