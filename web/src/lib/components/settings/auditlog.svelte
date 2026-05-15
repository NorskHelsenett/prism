<script>
  import { Fetch } from "$lib/fetchUtil";
  import { onMount, onDestroy } from 'svelte';
	import { fade } from 'svelte/transition';
	import { quintIn } from 'svelte/easing';
  import formatNumber from '$lib/formatNumber';


  let auditLogs = $state([]);
  let intervalId; // Declare intervalId in the component scope
  let newLogIds = $state(new Set()); // Set to track new log IDs
  let totalAudits = $state(0)
  let filterLogsBasedOnPath = $state(false)

  async function fetchEvents() {
    try {
      const result = await Fetch("/api/settings/audit");
      const newLogs = result.audits;
      totalAudits = result.total;
      const existingLogTimestamps = new Set(auditLogs.map(log => log.Timestamp));

      const currentPath = "/api/settings/audit"; // Get the current URL path

      const logsToAdd = newLogs.filter(log => {
        // Check if we should filter based on the path and if the log's timestamp is new
        const isNewLog = !existingLogTimestamps.has(log.Timestamp);
        const isPathDifferent = filterLogsBasedOnPath ? log.Action != currentPath : true;
        return isNewLog && isPathDifferent;
      });

      const fetchedLogIds = new Set(newLogs.map(log => log.Timestamp));
      newLogIds = new Set([...fetchedLogIds].filter(id => !existingLogTimestamps.has(id)));

      if (logsToAdd.length > 0) {
        auditLogs = [...logsToAdd, ...auditLogs];
      }

      // Remove the new log IDs after the transition duration
      setTimeout(() => {
        newLogIds = new Set();
      }, 2000); // Set this to the duration of your slide transition
    } catch (error) {
      console.error("Error fetching events:", error);
      // Optionally, handle the error in the UI
    }
  }

  onMount(() => {
    fetchEvents(); // Initial fetch when component mounts
    intervalId = setInterval(fetchEvents, 5000); // Set up the interval
    return () => {
      clearInterval(intervalId); // Clear the interval when the component is destroyed
    };
  });

  onDestroy(() => {
    clearInterval(intervalId);
  });

  function formatDate(dateString) {
    const options = { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hour12: false };
    return new Date(dateString).toLocaleDateString('en-US', options).replace(/\//g, '.').replace(',', '');
  }
</script>
<div class="card-body">

<div class="row mb-0">
  <div class="col-9">
    <div class="card-body">
    <div class="row align-items-center">
      <div class="col">
        <label class="form-check">
                <input class="form-check-input" type="checkbox" onclick={() => filterLogsBasedOnPath = !filterLogsBasedOnPath} value="{filterLogsBasedOnPath}">
                <span class="form-check-label">
                  Filter current path
                </span>
                <span class="form-check-description">
                  Do not show `/api/settints/audit` entries
                </span>
              </label>
    </div>
  </div>
  </div>

  </div>
  <div class="card-sm col-3">
    <div class="card-body">
      <div class="row align-items-center">
        <div class="col-auto">
          <span class="bg-info text-white avatar"><!-- Download SVG icon from http://tabler-icons.io/i/currency-dollar -->
            <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-file-text" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M14 3v4a1 1 0 0 0 1 1h4" /><path d="M17 21h-10a2 2 0 0 1 -2 -2v-14a2 2 0 0 1 2 -2h7l5 5v11a2 2 0 0 1 -2 2z" /><path d="M9 9l1 0" /><path d="M9 13l6 0" /><path d="M9 17l6 0" /></svg>
          </span>
        </div>
        <div class="col">
          <div class="font-weight-medium">
            Total {formatNumber(totalAudits)} audits
          </div>
          <div class="text-secondary">
            50 shown here
          </div>
        </div>
      </div>
    </div>
  </div>
</div>

<div class="table-responsive">
  <table class="table table-vcenter card-table table-striped">
    <thead>
      <tr>
        <th>Timestamp</th>
        <th>User</th>
        <th>Method</th>
        <th>PATH</th>
        <th>Status</th>
        <th></th>
      </tr>
    </thead>
    <tbody>

      {#each auditLogs as event (event.Timestamp)}
      <tr transition:fade={{ delay: 0, duration: 300, easing: quintIn}} class:bg-azure-lt={newLogIds.has(event.Timestamp)}>
        <td>{formatDate(event.Timestamp)}</td>
        <td>{event.UserEmail}</td>
        <td>{event.Method}</td>
        <td>{event.Action}</td>
        <td>{event.Status}</td>
        <td>{event.Description}</td>
      </tr>
      {/each}

    </tbody>
  </table>
</div>
</div>

<style>
  .new-log {
    background-color: green;
  }
</style>