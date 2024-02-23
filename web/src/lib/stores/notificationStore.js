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
        message: `It was successfull`
      };

      // Check if alert is a string and apply that to the message as default
      let finalAlert;
      if (typeof alert === 'string') {
        finalAlert = { ...defaultAlert, message: alert, id };
      } else {
        // Merge the provided alert object with the default values, overriding defaults if specified
        finalAlert = { ...defaultAlert, ...alert, id };
      }

      // Add the finalAlert to the list of alerts
      update((alerts) => [...alerts, finalAlert]);

      setTimeout(() => {
        update((alerts) => alerts.filter((a) => a.id !== id));
      }, 3000); // Alerts disappear after 3 seconds
    },
  };
}

export const notification = createAlertStore();
