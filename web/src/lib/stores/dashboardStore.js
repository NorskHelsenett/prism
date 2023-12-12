// src/stores/dashboardStore.js
import { writable } from 'svelte/store';

function createDashboardStore() {
  const { subscribe, set } = writable(0);

  let intervalId;

  async function fetchData() {
    try {
      const endpointResponse = await fetch('/.well-known/config.json');
      const config = await endpointResponse.json()
      const response = await fetch(`${config.apiEndpoint}/api/dashboard`, { credentials: 'include' });
      const data = await response.json();
      set(data);
    } catch (error) {
      console.error('Error fetching data:', error);
    }
  }

  function startPolling(interval = 60000) {
    stopPolling(); // Clear any existing interval
    fetchData(); // Fetch immediately
    intervalId = setInterval(fetchData, interval); // Then fetch periodically
  }

  function stopPolling() {
    clearInterval(intervalId);
  }

  return {
    subscribe,
    startPolling,
    stopPolling,
    refreshData: fetchData // Expose the fetchData method for manual refresh
  };
}

export const dashboardStore = createDashboardStore();
