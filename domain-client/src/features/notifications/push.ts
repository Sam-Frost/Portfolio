import { fetchVapidPublicKey, subscribePush, unsubscribePush } from "./api";

// Remembers the endpoint we last registered server-side, so syncPushSubscription
// can notice a rotation and drop the stale row.
const ENDPOINT_KEY = "domain_push_endpoint";

function readStoredEndpoint(): string | null {
  try {
    return localStorage.getItem(ENDPOINT_KEY);
  } catch {
    return null;
  }
}

function writeStoredEndpoint(endpoint: string | null): void {
  try {
    if (endpoint) localStorage.setItem(ENDPOINT_KEY, endpoint);
    else localStorage.removeItem(ENDPOINT_KEY);
  } catch {
    /* private mode / storage disabled — the server upsert still keeps things working */
  }
}

// Whether this browser can do Web Push at all. On iOS this is only true in
// the installed (Home-Screen) PWA, never a Safari tab — see isStandalone.
export function isPushSupported(): boolean {
  return "serviceWorker" in navigator && "PushManager" in window && "Notification" in window;
}

// iOS gates Web Push to the standalone home-screen app. Other platforms
// allow it in a tab too, so this is only used to tailor the hint text.
export function isStandalone(): boolean {
  const iosStandalone = (window.navigator as Navigator & { standalone?: boolean }).standalone === true;
  const displayMode = window.matchMedia?.("(display-mode: standalone)").matches ?? false;
  return iosStandalone || displayMode;
}

export function currentPermission(): NotificationPermission | "unsupported" {
  if (!isPushSupported()) return "unsupported";
  return Notification.permission;
}

async function readySW(): Promise<ServiceWorkerRegistration> {
  const reg = await navigator.serviceWorker.getRegistration();
  if (reg) return reg;
  return navigator.serviceWorker.register("/sw.js");
}

// Returns true if this device already has an active push subscription
// registered with our server.
export async function isSubscribedOnThisDevice(): Promise<boolean> {
  if (!isPushSupported()) return false;
  const reg = await navigator.serviceWorker.getRegistration();
  const sub = await reg?.pushManager.getSubscription();
  return Boolean(sub);
}

// Full opt-in flow: permission prompt -> browser push subscription ->
// register it server-side. Throws with a user-facing message on any step.
export async function enablePushOnThisDevice(): Promise<void> {
  if (!isPushSupported()) throw new Error("This browser doesn't support notifications.");

  const permission = await Notification.requestPermission();
  if (permission !== "granted") {
    throw new Error(
      permission === "denied"
        ? "Notifications are blocked. Enable them for this app in your device settings."
        : "Notification permission was dismissed.",
    );
  }

  const { key } = await fetchVapidPublicKey();
  if (!key) throw new Error("Push isn't configured on the server yet.");

  const reg = await readySW();
  await navigator.serviceWorker.ready;

  let sub = await reg.pushManager.getSubscription();
  if (!sub) {
    sub = await reg.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(key),
    });
  }

  const json = sub.toJSON();
  await subscribePush({
    endpoint: sub.endpoint,
    p256dh: json.keys?.p256dh ?? "",
    auth: json.keys?.auth ?? "",
    userAgent: navigator.userAgent,
  });
  writeStoredEndpoint(sub.endpoint);
}

export async function disablePushOnThisDevice(): Promise<void> {
  const reg = await navigator.serviceWorker.getRegistration();
  const sub = await reg?.pushManager.getSubscription();
  writeStoredEndpoint(null);
  if (!sub) return;
  await unsubscribePush(sub.endpoint).catch(() => {
    /* server may already have pruned it */
  });
  await sub.unsubscribe();
}

// Called on every authenticated app load. If the browser rotated the push
// subscription (or an earlier registration attempt half-failed), this
// re-registers the current one and drops the stale server row — so push
// keeps working even if the app hasn't been opened in a while and the SW's
// own pushsubscriptionchange handler never ran or failed. Cheap no-op when
// nothing changed. Never throws.
export async function syncPushSubscription(): Promise<void> {
  try {
    if (!isPushSupported() || Notification.permission !== "granted") return;

    const reg = await navigator.serviceWorker.getRegistration();
    if (!reg) return;
    const sub = await reg.pushManager.getSubscription();
    const stored = readStoredEndpoint();

    if (!sub) {
      // Permission is granted but the browser has no subscription — it was
      // dropped. Re-create it from the server's VAPID key.
      if (!stored) return;
      const { key } = await fetchVapidPublicKey();
      if (!key) return;
      const fresh = await reg.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(key),
      });
      const k = fresh.toJSON().keys ?? {};
      await subscribePush({ endpoint: fresh.endpoint, p256dh: k.p256dh ?? "", auth: k.auth ?? "", userAgent: navigator.userAgent });
      if (stored !== fresh.endpoint) await unsubscribePush(stored).catch(() => {});
      writeStoredEndpoint(fresh.endpoint);
      return;
    }

    if (stored === sub.endpoint) return; // already in sync

    const k = sub.toJSON().keys ?? {};
    await subscribePush({ endpoint: sub.endpoint, p256dh: k.p256dh ?? "", auth: k.auth ?? "", userAgent: navigator.userAgent });
    if (stored && stored !== sub.endpoint) await unsubscribePush(stored).catch(() => {});
    writeStoredEndpoint(sub.endpoint);
  } catch {
    /* best effort — the user can always re-enable from Settings */
  }
}

// VAPID keys travel as URL-safe base64; PushManager wants a Uint8Array.
function urlBase64ToUint8Array(base64String: string): Uint8Array<ArrayBuffer> {
  const padding = "=".repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, "+").replace(/_/g, "/");
  const raw = window.atob(base64);
  const buffer = new ArrayBuffer(raw.length);
  const output = new Uint8Array(buffer);
  for (let i = 0; i < raw.length; i++) output[i] = raw.charCodeAt(i);
  return output;
}
