import { goto } from '$app/navigation';
import { writable } from 'svelte/store';
import { normalizeApiEndpoint } from '$lib/config';

export const apiEndpoint = writable('');
export const providers = writable([]);
export const isLoading = writable(true);
export const isAuthenticated = writable(false);

const REQUEST_TIMEOUT_MS = 8000;

function isValidAppRedirect(path) {
  if (!path || typeof path !== 'string') return false;
  if (!path.startsWith('/')) return false;
  if (path.startsWith('/+')) return false;
  if (path.includes('.svelte')) return false;
  if (path.startsWith('/_app/')) return false;
  return true;
}

async function fetchWithTimeout(url, options = {}) {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);

  try {
    return await fetch(url, { ...options, signal: controller.signal });
  } finally {
    clearTimeout(timeout);
  }
}

function rememberRedirect(path = window.location.pathname + window.location.search) {
  if (path != "/login" && isValidAppRedirect(path)) {
    localStorage.setItem('redirectToAfterLogin', path);
  }
}

function redirectToLogin() {
  if (window.location.pathname !== "/login") {
    goto("/login");
  }
}

export async function initializeApiEndpoint(anon = false) {
  isLoading.set(true);
  try {
    const response = await fetchWithTimeout('/.well-known/config.json');
    if (response.ok) {
      const config = await response.json();
      const endpoint = normalizeApiEndpoint(config.apiEndpoint);
      apiEndpoint.set(endpoint);
      providers.set(config.providers ?? []);

      if(!anon){
        const url = `${endpoint}/api/profile`;
        const responseAuth = await fetchWithTimeout(url, { credentials: 'include' });

        if (responseAuth.ok) {
          isAuthenticated.set(true)
          isLoading.set(false);
          const redirectPath = localStorage.getItem('redirectToAfterLogin');
          if (redirectPath && isValidAppRedirect(redirectPath)) {
            localStorage.removeItem('redirectToAfterLogin');
            window.location.href = redirectPath != "/login" ? redirectPath : "/";
          } else if (redirectPath) {
            localStorage.removeItem('redirectToAfterLogin');
            window.location.href = "/";
          }
        }

        if (responseAuth.status === 401) {
          isAuthenticated.set(false);
          rememberRedirect();
          redirectToLogin();
        }

        if (responseAuth.status === 403) {
          isAuthenticated.set(false);
          const errormessage = await responseAuth.json()

          if (errormessage?.initiateOTP === true) {
            goto("/auth")
          } else {
            redirectToLogin();
          }
        }
      }

    } else {
      console.error('Failed to load API endpoint config:', response.status);
      isAuthenticated.set(false);
      if (!anon) {
        rememberRedirect();
        redirectToLogin();
      }
    }
  } catch (error) {
    console.error('Error fetching API endpoint config:', error);
    isAuthenticated.set(false);
    if (!anon) {
      rememberRedirect();
      redirectToLogin();
    }
  } finally {
    isLoading.set(false);
  }
}
