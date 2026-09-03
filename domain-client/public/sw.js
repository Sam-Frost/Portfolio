/* Service worker for the Domain Expansion PWA.
 *
 * Its only job is notifications: show every push (iOS revokes the
 * subscription otherwise) and route a click back into the app. It does NOT
 * cache anything — the app is online-only and CloudFront already serves it.
 */

self.addEventListener("install", () => {
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  event.waitUntil(self.clients.claim());
});

self.addEventListener("push", (event) => {
  let data = {};
  try {
    data = event.data ? event.data.json() : {};
  } catch {
    data = { title: "Domain Expansion", body: event.data ? event.data.text() : "" };
  }

  const title = data.title || "Domain Expansion";
  const options = {
    body: data.body || "",
    icon: "/icons/icon-192.png",
    badge: "/icons/icon-192.png",
    tag: data.tag || undefined,
    renotify: Boolean(data.tag),
    data: { url: data.url || "/" },
  };

  event.waitUntil(self.registration.showNotification(title, options));
});

// The browser can rotate a push subscription without the app being open.
// Re-subscribe with the same server key and tell the backend the new
// endpoint (unauthenticated /sync route — the SW has no bearer token). If
// this fails, the app-load re-sync in pwa.ts catches it next time it opens.
self.addEventListener("pushsubscriptionchange", (event) => {
  event.waitUntil(
    (async () => {
      try {
        const oldEndpoint = event.oldSubscription && event.oldSubscription.endpoint;
        const options =
          (event.oldSubscription && event.oldSubscription.options) || { userVisibleOnly: true };
        const sub = await self.registration.pushManager.subscribe(options);
        const keys = sub.toJSON().keys || {};
        await fetch("/api/notifications/subscriptions/sync", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            oldEndpoint: oldEndpoint || "",
            endpoint: sub.endpoint,
            p256dh: keys.p256dh || "",
            auth: keys.auth || "",
          }),
        });
      } catch {
        // Nothing more the SW can do here.
      }
    })(),
  );
});

self.addEventListener("notificationclick", (event) => {
  event.notification.close();
  const targetUrl = (event.notification.data && event.notification.data.url) || "/";

  event.waitUntil(
    self.clients.matchAll({ type: "window", includeUncontrolled: true }).then((clientList) => {
      for (const client of clientList) {
        if ("focus" in client) {
          client.focus();
          client.postMessage({ type: "notification-click", url: targetUrl });
          return;
        }
      }
      if (self.clients.openWindow) return self.clients.openWindow(targetUrl);
    }),
  );
});
