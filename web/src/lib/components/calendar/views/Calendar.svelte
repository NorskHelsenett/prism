<script>
  import { run } from 'svelte/legacy';

	import { Fetch } from "$lib/fetchUtil";
	import { onMount } from "svelte";

  /**
   * @typedef {Object} Props
   * @property {boolean} [reload]
   */

  /** @type {Props} */
  let { reload = true } = $props();
  let calendarEvents = $state([])

onMount(async () => {fetchCalendarEvents})

async function fetchCalendarEvents() {
  calendarEvents = await Fetch("/api/planning?pageSize=${1000}")
}

run(() => {
    if (!reload) {
    fetchCalendarEvents();
  }
  });

// Example: February 2024
let year = 2024;
let month = new Date().getMonth(); // February (0-based index: January is 0, February is 1, etc.)
let months =["Jan", "Feb","Mar", "Apr","May","Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]

let startDay = getStartDay(year, month);
let daysInMonth = getDaysInMonth(year, month);
let days = Array.from({ length: daysInMonth }, (_, i) => i + 1); // [1, 2, 3, ..., 28/29]
let emptyStartDays = Array.from({ length: startDay }); // Empty slots before the first day of the month


function getStartDay(year, month) {
  let firstDay = new Date(year, month, 1);
  return (firstDay.getDay() + 6) % 7; // Adjusted to make Monday (0)
}

// Get the number of days in the month
function getDaysInMonth(year, month) {
  return new Date(year, month + 1, 0).getDate();
}

  function eventsForDay(day) {
    return calendarEvents.filter(event => {
      const eventStartDate = new Date(event.dateFrom).getDate();
      const eventEndDate = new Date(event.dateTo).getDate();
      return eventStartDate <= day && eventEndDate >= day;
    });
  }

const weekday = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]
</script>

<style>
  .grid {
    display: grid;
    grid-template-columns: repeat(7, 1fr); /* 7 columns for each day of the week */
    grid-gap: 0px;
  }
  .grid-item {
    padding: 0px;
    min-height: 5em; /* Adjust height as necessary */
    min-width: 5em;
    text-align: center;
  }
  .grid-item:hover{
    border: 1px solid rgb(89, 69, 221);
  }
  .event-bar {
    margin-top: 5px;
    padding: 2px;
    background-color: #774bb0; /* Example color, consider generating unique colors for each event */
    color: white;
    text-align: center;
    border-radius: 4px;
  }
</style>


<h2>{months[month]} {year}</h2>
<div class="grid">
  <!-- Weekday headers -->
  {#each weekday as day}
    <div class="grid-item text-secondary">{day}</div>
  {/each}

  <!-- Empty cells before the first day -->
  {#each emptyStartDays as _}
    <div class="grid-item"></div>
  {/each}

  <!-- Days and events -->
  {#each days as day}
    <div class="grid-item">
      {day}
      {#if calendarEvents?.length > 0}
        {#each eventsForDay(day) as event}
          <div class="event-bar">{event.title}</div>
        {/each}
      {/if}
    </div>
  {/each}
</div>