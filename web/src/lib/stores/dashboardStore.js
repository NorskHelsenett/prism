// src/stores/dashboardStore.js
import { writable } from 'svelte/store';
import { Fetch } from '$lib/fetchUtil'; // Import customFetch

function createDashboardStore() {
  const { subscribe, set } = writable(0);

  let intervalId;
  let currentYear = new Date().getFullYear();

  function setYear(year) {
    currentYear = year;
    fetchData();
  }

  async function fetchData(year) {
    if (year !== undefined) {
      currentYear = year;
    }
    try {
      const data = await Fetch(`/api/dashboard?year=${currentYear}`);
      set(data);
    } catch (error) {
      console.error('Error fetching data:', error);
    }
  }

  function startPolling(interval = 60000) {
    stopPolling(); // Clear any existing interval
    fetchData(); // Fetch immediately
    intervalId = setInterval(() => fetchData(), interval); // Then fetch periodically
  }

  function stopPolling() {
    clearInterval(intervalId);
  }

  return {
    subscribe,
    startPolling,
    stopPolling,
    setYear,
    refreshData: fetchData // Expose the fetchData method for manual refresh
  };
}

export const dashboardStore = createDashboardStore();
