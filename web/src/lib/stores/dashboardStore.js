// src/stores/dashboardStore.js
import { writable } from 'svelte/store';
import { Fetch } from '$lib/fetchUtil'; // Import customFetch

function createDashboardStore() {
  const { subscribe, set } = writable(0);

  let intervalId;

  async function fetchData(year = new Date().getFullYear()) {
    try {
      // Fetching local configuration file
      const endpointResponse = await fetch('/.well-known/config.json');
      const config = await endpointResponse.json();

      // Use customFetch for API call
      const data = await Fetch(`/api/dashboard?year=${year}`);

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
