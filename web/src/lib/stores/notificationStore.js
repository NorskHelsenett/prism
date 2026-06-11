import { writable, derived, get } from 'svelte/store';
import { apiEndpoint } from '$lib/stores/configStore';
import { Fetch } from '$lib/fetchUtil.js';
import { toast } from 'svelte-sonner';
import { goto } from '$app/navigation';

// In-app notifications for the signed-in user. Delivery is a per-user SSE
// stream (see api/routes/notification_hub.go); the full list is re-fetched on
// every (re)connect so missed events self-heal. State lives in this module so
// the layout owns the connection lifecycle and the dropdown just renders.

/**
 * @typedef {Object} AppNotification
 * @property {number} id
 * @property {string} who - Actor email.
 * @property {string} what
 * @property {boolean} read
 * @property {string} where - App-relative path.
 * @property {string} when - ISO timestamp.
 * @property {string} [kind]
 */

/** @type {import('svelte/store').Writable<AppNotification[]>} */
export const notifications = writable([]);
export const unreadCount = derived(notifications, (list) =>
  list.filter((n) => n.read === false).length
);

/** @type {EventSource | null} */
let sse = null;

/**
 * @param {AppNotification[]} list
 * @returns {AppNotification[]}
 */
function sortByWhenDesc(list) {
  return [...list].sort((a, b) => new Date(b.when).getTime() - new Date(a.when).getTime());
}

export async function loadNotifications() {
  const result = await Fetch('/api/notification');
  if (Array.isArray(result)) notifications.set(sortByWhenDesc(result));
}

/** @param {number} id */
export async function markRead(id) {
  notifications.update((list) =>
    list.map((n) => (n.id === id ? { ...n, read: true } : n))
  );
  await Fetch(`/api/notification/${id}/read`, { method: 'PUT' });
}

export async function markAllRead() {
  notifications.update((list) => list.map((n) => ({ ...n, read: true })));
  await Fetch('/api/notification/read-all', { method: 'PUT' });
}

/** @param {AppNotification} notification */
export async function openNotification(notification) {
  await markRead(notification.id);
  // The server only stores app-relative paths in `where`, so navigating to it
  // can't redirect off-site.
  await goto(notification.where);
}

/** @param {AppNotification} notification */
async function toastFor(notification) {
  let label = notification.who;
  try {
    const userData = await Fetch(`/api/profile/${notification.who}`);
    if (userData?.name) label = userData.name;
  } catch (_) {
    /* fall back to the raw identifier */
  }
  toast.info(`${label} — ${notification.what}`, {
    action: {
      label: 'Show',
      onClick: () => openNotification(notification)
    }
  });
}

// -- SSE ----------------------------------------------------------------

export function connect() {
  const endpoint = get(apiEndpoint);
  if (!endpoint || sse) return;
  try {
    sse = new EventSource(`${endpoint}/api/notification/events`, { withCredentials: true });
  } catch (e) {
    sse = null;
    return;
  }

  // Fires on every (re)connect — EventSource auto-reconnects after drops, so
  // refreshing the list here covers any events missed while disconnected.
  sse.onopen = () => {
    loadNotifications();
  };

  sse.addEventListener('notification.created', (e) => {
    try {
      const evt = JSON.parse(e.data);
      if (!evt?.notification) return;
      notifications.update((list) =>
        sortByWhenDesc([evt.notification, ...list.filter((n) => n.id !== evt.notification.id)])
      );
      toastFor(evt.notification);
    } catch (_) { /* malformed event — next reconnect refetches */ }
  });

  // Cross-tab sync: marking read in one tab clears the badge in the others.
  sse.addEventListener('notification.read', (e) => {
    try {
      const evt = JSON.parse(e.data);
      notifications.update((list) =>
        list.map((n) => (n.id === evt.id ? { ...n, read: true } : n))
      );
    } catch (_) { /* ignore */ }
  });

  sse.addEventListener('notification.readAll', () => {
    notifications.update((list) => list.map((n) => ({ ...n, read: true })));
  });

  sse.onerror = () => {
    // EventSource auto-reconnects; just let it.
  };
}

export function disconnect() {
  if (sse) {
    sse.close();
    sse = null;
  }
  // Clear state so nothing from the previous session survives a logout or
  // user switch on the same browser.
  notifications.set([]);
}

// syncPushSubscription re-binds the browser's existing push endpoint to the
// *current* user right after login. UpsertSubscriber on the server takes the
// endpoint over from any previous owner (gated on matching keys), which is
// what keeps a shared browser delivering pushes only to whoever signed in
// last. Never prompts and never creates a new subscription — that stays the
// explicit job of NotificationButton.
export async function syncPushSubscription() {
  if (!('Notification' in window) || Notification.permission !== 'granted') return;
  if (!('serviceWorker' in navigator)) return;
  try {
    const registration = await navigator.serviceWorker.ready;
    const subscription = await registration.pushManager.getSubscription();
    if (!subscription) return;
    await Fetch('/api/notification/subscribe', {
      method: 'POST',
      body: JSON.stringify(subscription)
    });
  } catch (_) {
    /* best effort — the next explicit enable or push round-trip heals it */
  }
}
