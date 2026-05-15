<script>
	import { Fetch } from "$lib/fetchUtil";
	import { onMount } from "svelte";
	import Avatar from "../Avatar.svelte";

  let { notification } = $props();
  let user = $state()

  onMount(async () => {
    user = await fetchUserData(notification.who);
  });

  // Function to fetch user data
  async function fetchUserData(email) {
    return await Fetch(`/api/profile/${email}`);
  }

  const getHour = () => notification.when.split("T")[1].slice(0, 5);
  // beyond cognitive load

  function getRelativeDateLabel(inputDate) {
    const date = new Date(inputDate);
    const today = new Date();
    today.setHours(0, 0, 0, 0); // Set the time to the start of today

    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);

    // Get start of this week (Monday)
    const dayOfWeek = today.getDay(); // 0 (Sunday) to 6 (Saturday)
    const daysSinceMonday = dayOfWeek === 0 ? 6 : dayOfWeek - 1; // Calculate days since last Monday
    const startOfThisWeek = new Date(today);
    startOfThisWeek.setDate(today.getDate() - daysSinceMonday);

    // Get start of last week (previous Monday)
    const startOfLastWeek = new Date(startOfThisWeek);
    startOfLastWeek.setDate(startOfLastWeek.getDate() - 7);

    // Get start of this month
    const startOfThisMonth = new Date(today.getFullYear(), today.getMonth(), 1);

    // Get start of last month
    const startOfLastMonth = new Date(startOfThisMonth);
    startOfLastMonth.setMonth(startOfLastMonth.getMonth() - 1);

    if (date.toDateString() === today.toDateString()) {
      return "today";
    } else if (date.toDateString() === yesterday.toDateString()) {
      return "yesterday";
    } else if (date >= startOfThisWeek && date < today) {
      return "this week";
    } else if (date >= startOfLastWeek && date < startOfThisWeek) {
      return "last week";
    } else if (date >= startOfThisMonth && date < startOfThisWeek) {
      return "this month";
    } else {
      return "beyond cognitive load";
    }
  }
</script>
<div class="list-group list-group-flush list-group-hoverable w-100">
  <div class="list-group-item">
    <div class="row align-items-center">
      <div class="col-auto">
        {#if notification?.read}
        <span class="badge"></span>
        {:else}
        <span class="badge bg-green"></span>
        {/if}
      </div>
      <div class="col-auto">
        <Avatar email="{notification.who}" option={{ showName: false, size: "lg", emptyFields: false, circle: false, tooltipEnabled: false}}/>
      </div>
      <div class="col text-truncate">
        <div class="row">
          <div class="text-reset col">
            {user?.name}
          </div>
          <div class="text-secondary col-auto">{getHour()}</div>
        </div>
        <div class="d-block text-secondary text-truncate mt-n1">{notification.what}</div>
        <div class="d-block text-cyan text-truncate mt-n1">{getRelativeDateLabel(notification.when)}</div>
      </div>
    </div>
  </div>
</div>

<style>
  .list-group-item{
    padding: 0 !important;
  }
  .cursor-pointer { cursor: pointer; }
</style>