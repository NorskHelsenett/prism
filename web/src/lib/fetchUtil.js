// src/lib/fetchUtil.js
import { get } from 'svelte/store';
import { apiEndpoint } from '$lib/stores/configStore';
// import { goto } from '$app/navigation';

export async function Fetch(endpoint, options = {}) {
  const apiUrl = get(apiEndpoint);
  const url = `${apiUrl}${endpoint}`;

  try {
    const response = await fetch(url, { credentials: 'include', ...options });

    if (!response.ok) {
      console.error(`Error: ${response.status}`);
      if (response.status === 401) {
        window.location.href = '/login';
        // goto('/login');
        // Return a resolved promise to prevent throwing an error
        return Promise.resolve(null);
      }
      return null;
    }

    return await response.json();
  } catch (error) {
    console.error('Error fetching data:', error);
    // Optionally, handle network errors differently here
    return null;
  }
}
