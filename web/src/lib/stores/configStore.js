import { goto } from '$app/navigation';
import { writable } from 'svelte/store';

export const apiEndpoint = writable('');
export const providers = writable('')
export const isLoading = writable(true);
export const isAuthenticated = writable(false);

function isValidAppRedirect(path) {
  if (!path || typeof path !== 'string') return false;
  if (!path.startsWith('/')) return false;
  if (path.startsWith('/+')) return false;
  if (path.includes('.svelte')) return false;
  if (path.startsWith('/_app/')) return false;
  return true;
}

export async function initializeApiEndpoint(anon = false) {
  isLoading.set(true);
  try {
    const response = await fetch('/.well-known/config.json');
    if (response.ok) {
      const config = await response.json();
      apiEndpoint.set(config.apiEndpoint);
      providers.set(config.providers)

      if(!anon){
        const url = `${config.apiEndpoint}/api/profile`;
        const responseAuth = await fetch(url, { credentials: 'include' });

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
          isLoading.set(false);
          const returnPath = window.location.pathname + window.location.search;
          if (returnPath != "/login" && isValidAppRedirect(returnPath)){
            localStorage.setItem('redirectToAfterLogin', returnPath);
          }
          goto("/login")
        }

        if (responseAuth.status === 403) {
          const errormessage = await responseAuth.json()

          if (errormessage?.initiateOTP === true) {
            isLoading.set(false);
            goto("/auth")
          }
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
