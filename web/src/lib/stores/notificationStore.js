// src/stores/alertStore.js
import { writable } from 'svelte/store';

function createAlertStore() {
  const { subscribe, update } = writable([]);

  return {
    subscribe,
    addAlert: (alert) => {
      const id = Math.random().toString(36).substring(2, 9); // Unique ID for each alert

      // Define default values for the alert
      const defaultAlert = {
        type: 'success',
        title: 'We did it!',
      };

      // Merge the provided alert object with the default values, overriding defaults if specified
      const finalAlert = { ...defaultAlert, ...alert, id };

      update((alerts) => [...alerts, finalAlert]);

      setTimeout(() => {
        update((alerts) => alerts.filter((a) => a.id !== id));
      }, 3000); // Alerts disappear after 3 seconds
    },
  };
}

export const notification = createAlertStore();
