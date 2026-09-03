import { router } from "../router";
import { getToken } from "../features/auth/token";
import { syncPushSubscription } from "../features/notifications/push";

// Registers the service worker (needed for Web Push) and bridges a
// notification click from the SW back into the in-app router. Safe to call
// unconditionally — it no-ops where service workers aren't supported.
export function registerServiceWorker(): void {
  if (!("serviceWorker" in navigator)) return;

  window.addEventListener("load", () => {
    navigator.serviceWorker
      .register("/sw.js")
      .then(() => {
        // Heal a rotated/dropped push subscription while we're authenticated —
        // the SW's own pushsubscriptionchange handler can't reach an
        // authenticated endpoint and may never have run.
        if (getToken()) void syncPushSubscription();
      })
      .catch((err) => {
        console.warn("[pwa] service worker registration failed", err);
      });
  });

  // The SW posts this when the user taps a notification; the standalone app
  // uses a memory router, so a plain URL open wouldn't navigate an
  // already-running instance.
  navigator.serviceWorker.addEventListener("message", (event) => {
    const data = event.data as { type?: string; url?: string } | null;
    if (data?.type === "notification-click" && data.url) {
      router.navigate(data.url).catch(() => {
        /* target route may not exist in this build — ignore */
      });
    }
  });
}
