import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import { Fetch } from '$lib/fetchUtil';

function createUserStore() {
    // Initialize the store with data from local storage, if available
    const initialData = browser ? JSON.parse(localStorage.getItem('userStore') || '{}') : {};
    const { subscribe, set, update } = writable(initialData);

    function persistToLocalStorage(store) {
        if (browser) {
            const persistedData = Object.fromEntries(
                Object.entries(store).map(([email, userData]) => [
                    email,
                    {
                        email: userData.email,
                        name: userData.name,
                        picture: userData.picture
                    }
                ])
            );
            localStorage.setItem('userStore', JSON.stringify(persistedData));
        }
    }

    return {
        subscribe,
        getUser: async (email) => {
            return update(store => {
                if (store[email]) {
                    return store;
                } else {
                    Fetch(`/api/profile/${email}`).then(user => {
                        update(s => {
                            const newStore = { ...s, [email]: user };
                            persistToLocalStorage(newStore);
                            return newStore;
                        });
                    });
                    return store;
                }
            });
        },
        updateUser: (email, userData) => {
            update(store => {
                const updatedStore = {
                    ...store,
                    [email]: { ...store[email], ...userData }
                };
                persistToLocalStorage(updatedStore);
                return updatedStore;
            });
        },
        clear: () => {
            set({});
            if (browser) {
                localStorage.removeItem('userStore');
            }
        }
    };
}

export const userStore = createUserStore();