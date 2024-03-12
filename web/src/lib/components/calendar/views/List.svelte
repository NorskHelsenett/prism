<script>
import Avatar from '$lib/components/Avatar.svelte';
import { goto } from '$app/navigation';
import Dropdown from '$lib/components/Dropdown.svelte';
import Markdown from '$lib/components/Markdown.svelte';
import { Fetch } from '$lib/fetchUtil';

let selectedRow = -1
let calendarEvents = []
let showDropdown = []
export let reload = true

let months =["Jan", "Feb","Mar", "Apr","May","Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]

function eventIn(month, dateFrom, dateTo) {
  // Parse the month string to get the month index
  const monthIndex = months.indexOf(month);

  // Parse the dateFrom and dateTo strings to Date objects
  const fromDate = new Date(dateFrom);
  const toDate = new Date(dateTo);

  return fromDate.getMonth() <= monthIndex && toDate.getMonth() >= monthIndex;
}

async function fetchCalendarEvents() {
  const today = new Date().getFullYear()
  const startDate = `${today}-01-01`
  const endDate = `${today}-12-31`
  calendarEvents = await Fetch(`/api/planning?startDate=${startDate}&endDate=${endDate}`)
}

$: if (!reload) {
  fetchCalendarEvents();
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
  const formattedToday = months[new Date().getMonth()]

  function daysInMonth(month, year) {
    return new Date(year, month, 0).getDate();
  }

  let today = new Date();
  let currentDay = today.getDate(); // Get the current day (1-31)
  let monthIndex = today.getMonth(); // Month index (0-11)
  let currentYear = today.getFullYear();

  let totalDays = daysInMonth(monthIndex +1 , currentYear); // Month index is 0-based, add 1 to get the correct month
  let dayPercentage = (currentDay / totalDays) * 100;

function calculateStartPercentage(dateFrom, monthStr) {
    const date = new Date(dateFrom);
    const dateMonth = date.getMonth();
    const givenMonth = new Date(Date.parse(monthStr +" 1, " + date.getFullYear())).getMonth();

    if (dateMonth !== givenMonth) {
        return 0;
    } else {
        const totalDays = new Date(date.getFullYear(), dateMonth + 1, 0).getDate();
        return (date.getDate() / totalDays) * 100;
    }
}

function calculateLineLength(dateTo, monthStr) {
    const date = new Date(dateTo);
    const dateMonth = date.getMonth();
    const givenMonth = new Date(Date.parse(monthStr +" 1, " + date.getFullYear())).getMonth();


    if (dateMonth !== givenMonth) {
        return 100;
    } else {
        const totalDays = new Date(date.getFullYear(), dateMonth + 1, 0).getDate();
        return ((date.getDate() / totalDays) * 100)*0.8; //0.8 offset to address width overlap to not have gaps between the bars
    }
}

let formattedTotalHours; // Declare the variable outside the reactive block

const numberFormatter = new Intl.NumberFormat('sv-SE', {
    style: 'decimal',
    useGrouping: true,
    minimumFractionDigits: 0
});

$: if (calendarEvents) {
  const totalHours = calendarEvents.reduce((accumulator, event) => {
    return accumulator + (event?.estimate || 0);
  }, 0);

  formattedTotalHours = numberFormatter.format(totalHours);
  console.log(formattedTotalHours);
}

</script>
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
                          <button class="link" on:click="{() => goto(`/planning/${event.id}/view`)}">{event?.title}</button>
                        </h4>
                        <div class="grid-container text-secondary">
                          <strong title="Work Order">AO:</strong>
                          <button class="btn-none" on:click|preventDefault="{() => copyText(event?.workorder)}">{event?.workorder || ""}</button>
                          <!-- <strong title="Estimate">Est.:</strong> <span>4000 h</span> -->
                        </div>
                      </td>
                      <td class="text-secondary">
                        {#if event?.estimate}
                          {numberFormatter.format(event?.estimate)} h
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
                            <i class="ti ti-circle-check-filled text-azure" title="Finished"></i>
                          {:else if event.status == "Approved"}
                            <i class="ti ti-player-record-filled text-green" title="Approved"></i>
                          {:else}
                            <i class="ti ti-circle text-yellow"  title="Planning"></i>
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
                          <i class="ti ti-notes cursor-pointer" on:click="{() => showDropdown[index] = true}"></i>
                          <Dropdown bind:show={showDropdown[index]}>
                              <div class="card-body">
                                <Markdown markdown={event.note}/>
                              </div>
                          </Dropdown>
                        {/if}
                      </td>
                      {#each months as month}
                        {#if eventIn(month, event?.dateFrom, event?.dateTo)}

                          <td class="timeline-container" class:today={month === formattedToday}>
                            <span
        class="line"
        style="--start-percentage: {calculateStartPercentage(event?.dateFrom, month)}%; --line-length: {calculateLineLength(event?.dateTo, month)}%"
      ></span>
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
              <div class="card-footer d-flex align-items-center">
                  <p class="m-0 text-secondary">Showing <span>{calendarEvents.length}</span> assessments with a total of {formattedTotalHours} hours</p>
                </div>
            </div>

<style>
.table{
  overflow-x: hidden;
}

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
  bottom: 5px;
  left: var(--start-percentage);
  width: calc(var(--line-length)*1.15);
  height: 10px;
  background-color: var(--tblr-azure);
  border-radius: 5px;
  bottom: 40%;
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