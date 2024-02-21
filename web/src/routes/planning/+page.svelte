<script>
	import Modal from '$lib/components/Modal.svelte';
  import { pageMeta } from '$lib/stores/pageMeta';
  import { onMount } from 'svelte';
	import { Fetch } from '$lib/fetchUtil';
	import Avatar from '$lib/components/Avatar.svelte';
	import NewAssessment from '$lib/components/calendar/newAssessment.svelte';
	import { goto } from '$app/navigation';
	import Assessments from '$lib/components/dashboard/Assessments.svelte';

  let showModal = false
  let selectedRow = -1

  onMount(async () => {
      pageMeta.set({ pretitle: 'Planning',title: 'Plan future world domination' });

      // calendarEvents = await Fetch("/api/planning")
  });

  let calendarEvents = []

  let months =["Jan", "Feb","Mar", "Apr","May","Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]

function eventIn(month, dateFrom, dateTo) {
  // Parse the month string to get the month index
  const monthIndex = months.indexOf(month);

  // Parse the dateFrom and dateTo strings to Date objects
  const fromDate = new Date(dateFrom);
  const toDate = new Date(dateTo);

  return fromDate.getMonth() <= monthIndex && toDate.getMonth() >= monthIndex;
}

// Define an async function to fetch the data
async function fetchCalendarEvents() {
  calendarEvents = await Fetch("/api/planning")
}

$: if(showModal == false) {
  fetchCalendarEvents()
}

function copyText(content){
    if (!navigator.clipboard) {
      // Clipboard API not available
      console.error('Clipboard API is not available.');
      return;
    }

    navigator.clipboard.writeText(content).then(() => {
      // notification.addAlert({message: 'Content copied to clipboard successfully'});
    }).catch(err => {
      console.error('Failed to copy content to clipboard:', err);
    });
}
  const formattedToday = months[new Date().getMonth()+1]

  function daysInMonth(month, year) {
    return new Date(year, month, 0).getDate();
  }

  let today = new Date();
  let currentDay = today.getDate(); // Get the current day (1-31)
  let monthIndex = today.getMonth(); // Month index (0-11)
  let currentYear = today.getFullYear();

  let totalDays = daysInMonth(monthIndex + 1, currentYear); // Month index is 0-based, add 1 to get the correct month
  let dayPercentage = (currentDay / totalDays) * 100; // Calculate the percentage of the month that has passed
</script>

<!-- svelte-ignore missing-declaration -->
<Modal bind:showModal on:close={() => showModal = false} large={false}>
    <div class="card-header" slot="title">
      <div class="card-title">New Assassment
      </div>
    </div>
  <NewAssessment bind:showModal on:close={() => showModal = false}/>
</Modal>

        <!-- Page header -->
        <div class="page-header d-print-none">
          <div class="container-xl">
            <div class="row g-2 align-items-center">
              <div class="col">
                <h2 class="page-title">
                  Planning
                </h2>
              </div>
              <!-- Page title actions -->
              <div class="col-auto ms-auto d-print-none">
                <a href="#" class="btn btn-primary" on:click={() => showModal = !showModal} >
                  <!-- Download SVG icon from http://tabler-icons.io/i/plus -->
                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M12 5l0 14"></path><path d="M5 12l14 0"></path></svg>
                  Add
                </a>
              </div>
            </div>
          </div>
        </div>
        <!-- Page body -->
        <div class="page-body" style="margin-top: 17px;">
          <div class="container-xl">
            <ul class="nav nav-bordered mb-4">
              <li class="nav-item">
                <a class="nav-link" aria-current="page" href="#">Calendar</a>
              </li>
              <li class="nav-item">
                <a class="nav-link active" href="#">List</a>
              </li>
              <li class="nav-item">
                <a class="nav-link" href="#">Board</a>
              </li>
            </ul>

            <div class="card">
              <div class="table-responsive small">
                <table class="table table-vcenter card-table">
                  <thead>
                    <tr>
                      <th class="sticky-col first-col">Title</th>
                      <th>Estimate</th>
                      <th>Project</th>
                      <th>Requested by</th>
                      <th>Responsible</th>
                      <th>Status</th>
                      <th>Hackers</th>
                      <th>Note</th>
                      {#each months as month}
                        <th>{month}</th>
                      {/each}
                    </tr>
                  </thead>
                  <tbody>
                    {#if calendarEvents?.length > 0}
                    {#each calendarEvents as event, index}
                    <tr on:dblclick={() => goto(`/planning/${event.id}/view`)} on:click={() => selectedRow === index ? selectedRow = -1 : selectedRow = index} class:selected="{selectedRow === index}" >
                      <td class="sticky-col first-col" style="min-width:20em">
                        <h4 class="text-capitalize">
                          <a href="#" on:click="{() => goto(`/planning/${event.id}/view`)}">{event?.title}</a>
                        </h4>
                        <div class="grid-container text-secondary">
                          <strong title="Work Order">AO:</strong>
                          <button class="btn-none" on:click|preventDefault="{() => copyText(event?.workorder)}">{event?.workorder || ""}</button>
                          <!-- <strong title="Estimate">Est.:</strong> <span>4000 h</span> -->
                        </div>
                      </td>
                      <td class="text-secondary">
                        {#if event?.estimate}
                          {event?.estimate} h
                        {/if}
                      </td>
                      <td style="min-width:10em">
                        {#each event.projects as project}
                          <div class="badge bg-cyan-lt mt-1"><a href="/project/{project.id}/view">{project.name}</a></div>
                        {/each}
                      </td>
                      <td>
                        <Avatar email="{event.requester}" option={{ showName: false, size: "sm", emptyFields: true, circle: true}}/>
                      </td>
                      <td>
                        <Avatar email="{event.responsible_hacker}" option={{ showName: false, size: "sm", emptyFields: false, circle: true}}/>
                      </td>
                      <td>
                        <h1 style="margin: 0">
                        {#if event.status == "Finished"}
                          <i class="ti ti-circle-check-filled text-green" title="Finished"></i>
                        {:else if event.status == "Approved"}
                          <i class="ti ti-clock-filled text-yellow" title="Approved"></i>
                        {:else}
                          <i class="ti ti-calendar-time text-orange" title="Planning"></i>
                        {/if}
                        </h1>
                      </td>
                      <td>
                        <div class="avatar-list avatar-list-stacked" style="min-width:7em">
                          {#each event.hackers as hacker}
                            <Avatar email="{hacker?.email}" option={{ showName: false, size: "sm", emptyFields: false, circle: true}}/>
                            {/each}
                        </div>
                      </td>
                      <td>
                        {#if event?.note}
                          <i class="ti ti-notes" title="{event?.note}"></i>
                        {/if}
                      </td>
                      {#each months as month}
                        {#if eventIn(month, event?.dateFrom, event?.dateTo)}
                          <td class="timeline-container" class:today={month === formattedToday}>
                            <span class="line bg-azure"></span>
                              <div class="tooltip-content">
                                <strong>start:</strong> {event.dateFrom}<br>
                                <strong>end:</strong> {event.dateTo}
                              </div>
                              <span class:today-line={month === formattedToday} style="--day-percentage: {dayPercentage}" title="today"></span>
                            </td>
                        {:else}
                          <td class:today={month === formattedToday}>
                            <span class:today-line={month === formattedToday} style="--day-percentage: {dayPercentage}" title="today"></span>
                          </td>
                        {/if}
                        {/each}
                    </tr>
                    {/each}
                    {/if}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>

<style>

.selected {
  background-color: rgba(184, 196, 228, 0.05);
  cursor: pointer;
}

.grid-container {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 2px;
}

  .today-line {
    position: absolute;
    left: calc(var(--day-percentage) * 1%); /* Use the calculated percentage */
    width: 2px;
    background-color: rgba(var(--tblr-azure-rgb), 0.3); /* Example style for highlighting today's date column */
    z-index: 1000;
    top: 0;
    height: 101%;
    bottom: 0;
  }

  .btn-none{
    text-align: inherit;
    border: none;
    padding: 0;
    margin: 0;
    background-color: transparent;
    color: inherit;
    box-shadow: none;
  }
  .btn-none:hover {
    color: rgba(var(--tblr-link-color-rgb),var(--tblr-link-opacity,1));
  }

  .sticky-col h4{
    margin-bottom: 0;
  }

  .sticky-col {
    position: -webkit-sticky; /* For Safari */
    position: sticky;
    background-color: var(--tblr-body-bg); /* Background color is necessary to avoid content overlap */
    left: 0;
    z-index: 100; /* Ensure the sticky column is above other elements */
}

/* Add this if you want a border separation */
.first-col {
    border-right: solid 1px var(--tblr-body-bg); /* Bootstrap's default border color */
}

.line {
  position: absolute;
  bottom: 40%; /* Adjust as necessary to align with the bottom of the month divs */
  width: 115%;
  left: 0;
  border-radius: 5px;
  height: 10px; /* Thickness of the line */
  /* background-color: #000; Color of the line */
  /* No need to set width and left here if you're doing it inline as in the HTML example */
}
.timeline-container {
  position: relative;
}

.tooltip-content {
  visibility: hidden;
  width: 120px;
  background-color: black;
  color: #fff;
  text-align: center;
  border-radius: 6px;
  padding: 5px 0;
  position: absolute;
  z-index: 1;
}

.timeline-container:hover .tooltip-content {
  visibility: visible;
}

  .today {
    position: relative;
    /* border-left: 2px solid rgba(var(--tblr-azure-rgb), 0.5); */
  }

</style>