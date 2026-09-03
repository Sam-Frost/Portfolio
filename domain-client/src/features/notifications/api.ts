import { apiRequest } from "../../lib/apiClient";

export interface PushSubscriptionInput {
  endpoint: string;
  p256dh: string;
  auth: string;
  userAgent?: string | null;
}

export function fetchVapidPublicKey(): Promise<{ key: string }> {
  return apiRequest<{ key: string }>("/api/notifications/vapid-public-key");
}

export function subscribePush(input: PushSubscriptionInput): Promise<void> {
  return apiRequest<void>("/api/notifications/subscriptions", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function unsubscribePush(endpoint: string): Promise<void> {
  return apiRequest<void>("/api/notifications/subscriptions", {
    method: "DELETE",
    body: JSON.stringify({ endpoint }),
  });
}

export function sendTestNotification(): Promise<void> {
  return apiRequest<void>("/api/notifications/test", { method: "POST" });
}
