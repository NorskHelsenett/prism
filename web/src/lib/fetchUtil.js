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

export async function FetchFile(endpoint, filename = "") {
  const apiUrl = get(apiEndpoint);
  const url = `${apiUrl}${endpoint}`;

  try {
    const response = await fetch(url, { credentials: 'include' });

    if (!response.ok) {
      console.error(`Error: ${response.status}`);
      return;
    }

    const blob = await response.blob();
    const downloadUrl = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = downloadUrl;
    a.download = filename || endpoint.split('/').pop();
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(downloadUrl);
    document.body.removeChild(a);
  } catch (error) {
    console.error('Error downloading file:', error);
  }
}