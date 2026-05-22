<!-- NotificationButton.svelte -->
<script>
	import { goto } from "$app/navigation";
	import { Fetch } from "$lib/fetchUtil";
	import { onMount } from "svelte";
  import { createEventDispatcher } from 'svelte';

  /**
   * @typedef {Object} Props
   * @property {string} [notificationPermission]
   */

  /** @type {Props} */
  let { notificationPermission = $bindable('default') } = $props();
  const dispatch = createEventDispatcher();
  let applicationServerKey = ""

  function updatePermission(newPermission) {
    notificationPermission = newPermission;
    dispatch('permissionChange', { notificationPermission });
  }

  function askNotificationPermission() {
    // Check if the browser supports notifications
    if (!("Notification" in window)) {
      console.log("This browser does not support notifications.");
      return;
    }

    if(notificationPermission === 'granted'){return}

    Notification.requestPermission().then((permission) => {
      if(permission === "granted") {
        subscribeUserToPush();
      }
      updatePermission(permission);
    });
  }

  onMount(async() => {
    // Check the current permission status when the component is mounted
    if ('Notification' in window) {
      notificationPermission = Notification.permission;
      updatePermission(notificationPermission);
    }
    applicationServerKey = await Fetch(`/api/notification/publicKey`);
  });

  async function subscribeUserToPush() {
  const registration = await navigator.serviceWorker.ready;
  const subscription = await registration.pushManager.subscribe({
    userVisibleOnly: true,
    // You may need to generate a suitable application server key
    // This could be a VAPID key for instance
    applicationServerKey: applicationServerKey
  });

  // Send the subscription to the Go server
  await Fetch("/api/notification/subscribe", {
    method: "POST",
    body: JSON.stringify(subscription),
  });
}
</script>

<button class="btn btn-primary btn-pill w-70" onclick={askNotificationPermission}>Enable</button>

<!-- <div>
  <label class="row">
    <span class="col">Push Notifications</span>
    <span class="col-auto">
      <label class="form-check form-check-single form-switch">
        <input class="form-check-input" type="checkbox" on:click={askNotificationPermission} checked="{notificationPermission === 'granted'}" disabled="{notificationPermission === 'granted'}">
      </label>
    </span>
  </label>
</div> -->