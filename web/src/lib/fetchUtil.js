import { get } from 'svelte/store';
import { apiEndpoint } from '$lib/stores/configStore';
import { goto } from '$app/navigation'; // for declarative redirects

export async function Fetch(endpoint, options = {}) {
  const apiUrl = get(apiEndpoint);
  const url = `${apiUrl}${endpoint}`;

  try {
    const response = await fetch(url, { credentials: 'include', ...options });

    if (!response.ok) {
      console.error(`Error: ${response.status}`);

      if (response.status === 401) {
        const returnPath = window.location.pathname + window.location.search;
        if (returnPath != "/login"){
          localStorage.setItem('redirectToAfterLogin', returnPath);
        }
        window.location.href = '/login';
        return Promise.resolve(null);
      }

      if (response.status === 403) {
        const errormessage = await response.json();

        if (errormessage?.initiateOTP === true) {
          window.location.href = "/auth"
          return Promise.resolve(null);
        }

        goto('/403');
        return Promise.resolve(null);
      }

      if (response.status === 404) {
        window.location.href = '/404';
        return Promise.resolve(null);
      }

      const errorResponse = await response.json();
      const errorMessage = errorResponse.error || 'Unknown error occurred';
      // Return the error message instead of null
      return { error: errorMessage };
    }

    return await response.json();
  } catch (error) {
    console.error('Error fetching data:', error);
    return null;
  }
}

export async function FetchBlob(endpoint, filename = "") {
  const apiUrl = get(apiEndpoint);
  const url = `${apiUrl}${endpoint}`;

  try {
    const response = await fetch(url, { credentials: 'include' });

    if (!response.ok) {
      console.error(`Error: ${response.status}`);
      return;
    }

    return await response.blob();
  } catch (error) {
    console.error('Error downloading file:', error);
  }
}

export async function FetchFile(endpoint, filename = "") {
  const apiUrl = get(apiEndpoint);
  const url = `${apiUrl}${endpoint}`;

  try {
    const response = await fetch(url, { credentials: 'include' });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const blob = await response.blob();
    const downloadUrl = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.style.display = 'none'; // Hide the anchor element
    a.href = downloadUrl;
    a.download = filename || endpoint.split('/').pop();
    document.body.appendChild(a);

    return new Promise((resolve, reject) => {
      const cleanup = () => {
        window.URL.revokeObjectURL(downloadUrl);
        if (document.body.contains(a)) {
          document.body.removeChild(a);
        }
      };

      a.onclick = () => {
        setTimeout(() => {
          cleanup();
          resolve();
        }, 100);
      };

      a.click();

      // Fallback in case the click event doesn't fire
      setTimeout(() => {
        cleanup();
        resolve();
      }, 1000);
    });
  } catch (error) {
    console.error('Error downloading file:', error);
    throw error;
  }
}
