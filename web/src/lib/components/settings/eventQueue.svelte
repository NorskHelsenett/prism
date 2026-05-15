<script>
  import { Fetch } from "$lib/fetchUtil";
  import { onMount, onDestroy } from 'svelte';

  let events = $state([]);
  let intervalId; // Declare intervalId in the component scope

  async function fetchEvents() {
    try {
      events = await Fetch("/api/settings/events");
    } catch (error) {
      console.error("Error fetching events:", error);
      // Optionally, handle the error in the UI
    }
  }

  async function deleteEvent(id) {
    try {
      await Fetch(`/api/settings/events/${id}`, {
        method: "DELETE"
      });
      await fetchEvents(); // Re-fetch events after retrying
    } catch (error) {
      console.error("Error retrying event:", error);
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

  async function updateEvent(id, status=false) {
    try {
      await Fetch(`/api/settings/events/${id}/update/${status}`, {
        method: "PUT"
      });
      await fetchEvents(); // Re-fetch events after retrying
    } catch (error) {
      console.error("Error retrying event:", error);
    }
  }

  let showErrorTooltip = false

  function getKind(kind) {
    if (kind === 1) {
      return "New Vulnerability"
    } else if (kind === 2) {
      return "Comment"
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
        <th>Kind</th>
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
        <td class="text-azure" title="{getKind(event.Kind)}">
        {#if event.Kind === 1}
        <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-file-plus" width="44" height="44" viewBox="0 0 24 24" stroke-width="1.5" stroke="#00abfb" fill="none" stroke-linecap="round" stroke-linejoin="round">
          <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
          <path d="M14 3v4a1 1 0 0 0 1 1h4" />
          <path d="M17 21h-10a2 2 0 0 1 -2 -2v-14a2 2 0 0 1 2 -2h7l5 5v11a2 2 0 0 1 -2 2z" />
          <path d="M12 11l0 6" />
          <path d="M9 14l6 0" />
        </svg>
        {:else if event.Kind === 2}
        <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-message-plus" width="44" height="44" viewBox="0 0 24 24" stroke-width="1.5" stroke="#00abfb" fill="none" stroke-linecap="round" stroke-linejoin="round">
          <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
          <path d="M8 9h8" />
          <path d="M8 13h6" />
          <path d="M12.01 18.594l-4.01 2.406v-3h-2a3 3 0 0 1 -3 -3v-8a3 3 0 0 1 3 -3h12a3 3 0 0 1 3 3v5.5" />
          <path d="M16 19h6" />
          <path d="M19 16v6" />
        </svg>
        {/if}
        </td>
        <td>
          {#if event.Error && event.Processed}
            <span class="badge bg-danger me-1"></span> <span class="badge bg-red-lt">Error</span>
          {:else if event.Processed}
            <span class="badge bg-success me-1"></span> <span class="badge bg-green-lt">Success</span>
          {:else}
            <span class="badge bg-warning me-1 badge-blink"></span> <span class="badge bg-yellow-lt">Pending</span>
          {/if}
        </td>
        <td class="text-center">
          {#if event.UpdatedAt == "0001-01-01T00:00:00Z" || (event.Error && event.Processed)}
            <svg onmouseover={() => event.showTooltip = true} onmouseout={() => event.showTooltip = false} xmlns="http://www.w3.org/2000/svg" class="cursor-pointer text-danger icon icon-tabler icon-tabler-alert-triangle" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M12 9v4" /><path d="M10.363 3.591l-8.106 13.534a1.914 1.914 0 0 0 1.636 2.871h16.214a1.914 1.914 0 0 0 1.636 -2.87l-8.106 -13.536a1.914 1.914 0 0 0 -3.274 0z" /><path d="M12 16h.01" /></svg>
              {#if event.showTooltip}
                    <div class="rounded-2 alert alert-danger bg-body hover-alert" role="alert">
                      <div class="d-flex">
                        <div>
                          <!-- Download SVG icon from http://tabler-icons.io/i/alert-circle -->
                          <svg xmlns="http://www.w3.org/2000/svg" class="icon alert-icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M3 12a9 9 0 1 0 18 0a9 9 0 0 0 -18 0"></path><path d="M12 8v4"></path><path d="M12 16h.01"></path></svg>
                        </div>
                        <div>
                          <h4 class="alert-title">Error message!</h4>
                          <div class="text-secondary">{event.Error}</div>
                        </div>
                      </div>
                    </div>
              {/if}
          {:else if !event.Error}
            {formatDate(event.UpdatedAt)}
          {/if}
        </td>
        <td class="text-center">
          {#if event.Processed}
            <a href="#" class="btn btn-sm btn-pill text-cyan" onclick={updateEvent(event.ID)}>retry</a>
            <a href="#" class="btn btn-sm btn-pill text-red" onclick={deleteEvent(event.ID)}>delete</a>
          {:else}
            <a href="#" class="btn btn-sm btn-pill text-green" onclick={updateEvent(event.ID, true)}>finish</a>
            <!-- <a href="#" class="btn btn-sm btn-pill text-red" on:click={deleteEvent(event.ID)}>delete</a> -->
          {/if}
        </td>
      </tr>
      {/each}

    </tbody>
  </table>
</div>
</div>

<style>
.hover-alert {
  transform: translateX(-10%);
  position: absolute;
}
</style>