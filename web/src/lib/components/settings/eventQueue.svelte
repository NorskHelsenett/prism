<script>
  import { Fetch } from "$lib/fetchUtil";
  import { onMount, onDestroy } from 'svelte';

  let events = [];
  let intervalId; // Declare intervalId in the component scope

  async function fetchEvents() {
    try {
      events = await Fetch("/api/settings/events");
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

  async function retryEvent(id) {
    try {
      await Fetch(`/api/settings/events/${id}/update/false`, {
        method: "PUT"
      });
      await fetchEvents(); // Re-fetch events after retrying
    } catch (error) {
      console.error("Error retrying event:", error);
    }
  }
</script>

<div class="card-body">
<div class="table-responsive">
  <table class="table table-vcenter card-table table-striped">
    <thead>
      <tr>
        <th>Created</th>
        <th>Table ID</th>
        <th>Table Name</th>
        <th>Processed</th>
        <th>Processed at</th>
        <th></th>
      </tr>
    </thead>
    <tbody>

      {#each events as event}
      <tr>
        <td>{formatDate(event.CreatedAt)}</td>
        <td>{event.TableID}</td>
        <td>{event.TableName}</td>
        <td>
          {#if event.Processed}
            <span class="badge bg-success me-1"></span> Success
          {:else}
            <span class="badge bg-warning me-1 badge-blink"></span> Pending
          {/if}
        </td>
        <td>
          {#if event.UpdatedAt == "0001-01-01T00:00:00Z"}
          {:else}
            {formatDate(event.UpdatedAt)}
          {/if}
        </td>
        <td>
          <a href="#" class="btn btn-sm btn-pill" on:click={retryEvent(event.ID)}>retry</a>
        </td>
      </tr>
      {/each}

    </tbody>
  </table>
</div>
</div>