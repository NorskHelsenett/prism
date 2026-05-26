// Service worker for PRISM web-push.
//
// Two listeners:
//  - push: renders the notification on screen.
//  - notificationclick: focuses an existing window if there's one open at the
//    target URL, otherwise opens a new one. We validate that the URL is a
//    same-origin path before navigating so a malicious push payload can't
//    redirect the user off-site.
//  - pushsubscriptionchange: when the browser silently rotates or expires
//    the subscription endpoint, re-subscribe with the cached VAPID key and
//    POST the new subscription to the server so push keeps working without
//    requiring the user to re-enable notifications.

self.addEventListener('push', event => {
  if (!event.data) return;
  let data;
  try {
    data = event.data.json();
  } catch (e) {
    return;
  }
  event.waitUntil(
    self.registration.showNotification(data.title || 'PRISM', {
      body: data.body || '',
      image: data.icon,
      data: { url: data.url || '/' }
    })
  );
});

self.addEventListener('notificationclick', function(event) {
  event.notification.close();

  const rawUrl = event.notification.data && event.notification.data.url;
  // Only follow same-origin paths. Anything else is dropped — push payloads
  // are server-trusted but defense-in-depth costs nothing here.
  const safePath = (typeof rawUrl === 'string' && rawUrl.startsWith('/')) ? rawUrl : '/';
  const urlToOpen = new URL(safePath, self.location.origin).href;

  const promiseChain = clients
    .matchAll({ type: 'window', includeUncontrolled: true })
    .then((windowClients) => {
      for (const windowClient of windowClients) {
        if (windowClient.url === urlToOpen) {
          return windowClient.focus();
        }
      }
      return clients.openWindow(urlToOpen);
    });

  event.waitUntil(promiseChain);
});

// pushsubscriptionchange fires when the user-agent invalidates the existing
// subscription (key rotation, quota cleanup, profile reset, etc.). Without
// handling it, the browser silently stops receiving pushes until the user
// next opens the app and the page-side button re-subscribes. We rebuild here
// so notifications survive these rotations even when the tab is closed.
self.addEventListener('pushsubscriptionchange', event => {
  event.waitUntil(handlePushSubscriptionChange());
});

async function handlePushSubscriptionChange() {
  const applicationServerKey = await getCachedVapidKey();
  if (!applicationServerKey) return;
  try {
    const newSub = await self.registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey
    });
    await fetch('/api/notification/subscribe', {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(newSub)
    });
  } catch (e) {
    // Either no permission, no network, or the server rejected — there's
    // nothing the SW can do beyond logging. The page-side mount will retry
    // on next visit.
    console.warn('pushsubscriptionchange: re-subscribe failed', e);
  }
}

// getCachedVapidKey reads from the keyval cache that the page-side script
// keeps in IndexedDB-equivalent localStorage (mirrored into a SW-readable
// cache by NotificationButton on subscribe). The SW has no localStorage, so
// we fall back to fetching the public key from the API when no cache exists.
async function getCachedVapidKey() {
  try {
    const cache = await caches.open('prism-notifications');
    const hit = await cache.match('/__vapid');
    if (hit) {
      return await hit.text();
    }
  } catch (e) { /* fall through to network */ }
  try {
    const resp = await fetch('/api/notification/publicKey', { credentials: 'include' });
    if (!resp.ok) return null;
    return await resp.json();
  } catch (e) {
    return null;
  }
}
