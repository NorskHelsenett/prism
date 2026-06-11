<!-- NotificationButton.svelte -->
<script>
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

  // Cached VAPID public key. Stored in localStorage keyed on origin so the
  // round trip to /api/notification/publicKey only happens once per browser
  // per server, and is mirrored into a SW-readable Cache entry so the
  // pushsubscriptionchange handler can re-subscribe without contacting the
  // page.
  const VAPID_CACHE_KEY = `prism:vapidPublicKey:${window.location.origin}`;
  let applicationServerKey = "";

  function updatePermission(newPermission) {
    notificationPermission = newPermission;
    dispatch('permissionChange', { notificationPermission });
  }

  async function ensureVapidKey() {
    if (applicationServerKey) return applicationServerKey;
    const cached = localStorage.getItem(VAPID_CACHE_KEY);
    if (cached) {
      applicationServerKey = cached;
    } else {
      applicationServerKey = await Fetch(`/api/notification/publicKey`);
      if (applicationServerKey) {
        localStorage.setItem(VAPID_CACHE_KEY, applicationServerKey);
      }
    }
    // Mirror into the cache so service-worker.js can read it without
    // touching the network when handling pushsubscriptionchange.
    if (applicationServerKey && 'caches' in window) {
      try {
        const cache = await caches.open('prism-notifications');
        await cache.put('/__vapid', new Response(applicationServerKey));
      } catch (_) { /* non-fatal */ }
    }
    return applicationServerKey;
  }

  function askNotificationPermission() {
    if (!("Notification" in window)) {
      console.log("This browser does not support notifications.");
      return;
    }
    if (notificationPermission === 'granted') {
      // Already granted — re-affirm with the server in case the row was
      // dropped (dead-endpoint cleanup, ResetNotifications, etc).
      reaffirmSubscription();
      return;
    }
    Notification.requestPermission().then((permission) => {
      if (permission === "granted") {
        subscribeUserToPush();
      }
      updatePermission(permission);
    });
  }

  onMount(async () => {
    if ('Notification' in window) {
      notificationPermission = Notification.permission;
      updatePermission(notificationPermission);
    }
    await ensureVapidKey();
    if (notificationPermission === 'granted') {
      // Reattach to whatever the browser already has so we don't churn
      // server rows on every page mount. UpsertSubscriber is idempotent
      // for the (email, endpoint) pair so this is cheap.
      await reaffirmSubscription();
    }
  });

  async function reaffirmSubscription() {
    const registration = await navigator.serviceWorker.ready;
    let subscription = await registration.pushManager.getSubscription();
    if (!subscription) {
      subscription = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: await ensureVapidKey()
      });
    }
    await Fetch("/api/notification/subscribe", {
      method: "POST",
      body: JSON.stringify(subscription),
    });
  }

  async function subscribeUserToPush() {
    const registration = await navigator.serviceWorker.ready;
    let subscription = await registration.pushManager.getSubscription();
    if (!subscription) {
      subscription = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: await ensureVapidKey()
      });
    }
    await Fetch("/api/notification/subscribe", {
      method: "POST",
      body: JSON.stringify(subscription),
    });
  }
</script>

<button class="btn btn-primary btn-pill w-70" onclick={askNotificationPermission}>Enable</button>
