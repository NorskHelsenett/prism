// userStore.js
import { writable } from 'svelte/store';
import { getApiEndpoint } from './config';

export const accessLevels = writable({});

function createUserStore() {
  const { subscribe, set, update } = writable({ loading: true });

  async function fetchUser() {
    const baseUrl = await getApiEndpoint();
    const response = await fetch(baseUrl + '/api/profile', { credentials: 'include' });
    if (response.ok) {
      const data = await response.json();
      set({ ...data, loading: false });
      const accessListResponse = await fetch(baseUrl + ("/api/profile/access-list"), { credentials: 'include' })
      const jsonResponse = await accessListResponse.json()
      // You transform the JSON into the desired dictionary format.
      const accessList = jsonResponse.Permissions.reduce((acc, permission) => {
        // Initialize an empty object for each resource if not already initialized
        acc[permission.Resource] = acc[permission.Resource] || {};

        // Map each action as a key with the value true
        permission.Action.forEach(action => {
          acc[permission.Resource][action] = true;
        });

        return acc;
      }, {});
      accessLevels.set(accessList);
    } else if (response.status == 401 && window.location.pathname !== "/login") {
      set({ loading: false });
      window.location.href = "/login"
    } else if (response.status == 403 && window.location.pathname !== "/auth"){
      set({ loading: false });
      window.location.href = "/auth"}
    else {
      console.error('Failed to fetch user data', response.status);
    }
  }

  async function updateUser(userData) {
    // Optimistically update the store
    const baseUrl = await getApiEndpoint();

    update(current => ({ ...current, ...userData }));

    // Then push to the API
    const response = await fetch(baseUrl + '/api/profile', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(userData),
      credentials: 'include'
    });

    if (!response.ok) {
      // If the API call fails, revert the changes
      console.error('Failed to update user data');
      fetchUser(); // Refresh user data from the server
    }
  }

  // Initially fetch user data
  fetchUser();

  return {
    subscribe,
    set,
    updateUser
  };
}

export const userStore = createUserStore();
