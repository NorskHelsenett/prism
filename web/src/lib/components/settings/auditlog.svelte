<script>
  import { Fetch } from "$lib/fetchUtil";
  import { onMount, onDestroy } from 'svelte';
	import { fade } from 'svelte/transition';
	import { quintIn } from 'svelte/easing';

  let auditLogs = [];
  let intervalId; // Declare intervalId in the component scope
  let newLogIds = new Set(); // Set to track new log IDs

  async function fetchEvents() {
    try {
      const newLogs = await Fetch("/api/settings/audit");
      const existingLogTimestamps = new Set(auditLogs.map(log => log.Timestamp));

      const logsToAdd = newLogs.filter(log => !existingLogTimestamps.has(log.Timestamp));

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