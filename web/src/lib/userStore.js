// userStore.js
import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import { getApiEndpoint } from './config';

export const accessLevels = writable({});

/**
 * @typedef {Object} Permission
 * @property {string} Resource
 * @property {string[]} Action
 */

function createUserStore() {
	const { subscribe, set, update } = writable({ loading: true });

	async function fetchUser() {
		if (!browser) {
			set({ loading: false });
			return;
		}

		try {
			const baseUrl = await getApiEndpoint();
			const response = await fetch(baseUrl + '/api/profile', { credentials: 'include' });
			if (response.ok) {
				const data = await response.json();
				set({ ...data, loading: false });
				const accessListResponse = await fetch(baseUrl + '/api/profile/access-list', {
					credentials: 'include'
				});

				if (!accessListResponse.ok) {
					console.error('Failed to fetch access list', accessListResponse.status);
					accessLevels.set({});
					return;
				}

				const jsonResponse = await accessListResponse.json();
				/** @type {Permission[]} */
				const permissions = jsonResponse.Permissions ?? [];
				const accessList = permissions.reduce(
					(acc, permission) => {
						acc[permission.Resource] = acc[permission.Resource] || {};

						permission.Action.forEach((action) => {
							acc[permission.Resource][action] = true;
						});

						return acc;
					},
					/** @type {Record<string, Record<string, boolean>>} */ ({})
				);
				accessLevels.set(accessList);
				return;
			}

			accessLevels.set({});
			set({ loading: false });

			if (response.status === 401) {
				if (window.location.pathname !== '/login') {
					window.location.href = '/login';
				}
				return;
			}

			if (response.status === 403) {
				if (window.location.pathname !== '/auth') {
					window.location.href = '/auth';
				}
				return;
			}

			console.error('Failed to fetch user data', response.status);
		} catch (error) {
			accessLevels.set({});
			set({ loading: false });
			console.error('Failed to fetch user data', error);
		}
	}

	/** @param {Record<string, unknown>} userData */
	async function updateUser(userData) {
		// Optimistically update the store
		const baseUrl = await getApiEndpoint();

		update((current) => ({ ...current, ...userData }));

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
