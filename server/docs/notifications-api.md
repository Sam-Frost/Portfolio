# Notifications API

Backing API for outbound notifications — a morning digest of due/overdue todos and (Phase 2)
per-todo reminders — delivered by **email** (SMTP) and **Web Push** to the installed
`domain-client` PWA.

Feature package: `internal/notification/` (subscriptions + fan-out) + `internal/mailer/`
(shared SMTP boundary) + `internal/scheduler/` (the one background ticker). Config is env-only —
see `server-config.md`. Unconfigured channels no-op.

## Data model

```
PushSubscription {              // one row per browser, from PushManager.subscribe()
  endpoint  string              // push-service URL, unique key
  p256dh    string              // client public key (payload encryption)
  auth      string              // client auth secret
  userAgent string | null
}
```

`notification_log(kind, ist_date)` is an at-most-once-per-IST-day ledger; only the morning
digest uses it today (`kind = "morning_digest"`).

The recipient email + schedule + channel toggles live on the **settings** singleton
(`notifications` section) — see `docs/settings-api.md`.

## Auth

Same bearer-JWT gate as the rest of `/api/*`.

## Endpoints

### `GET /api/notifications/vapid-public-key`

`200 → { "key": "<base64url>" }`. Empty string when Web Push isn't configured — the client
treats that as "push unavailable".

### `POST /api/notifications/subscriptions`

Register (or refresh) this browser's push subscription. Upserts by `endpoint`.

```json
{ "endpoint": "https://...", "p256dh": "...", "auth": "...", "userAgent": "..." }
```

`204` on success · `400` if `endpoint`/`p256dh`/`auth` missing.

### `POST /api/notifications/subscriptions/sync`  — unauthenticated

Heals a **rotated** push subscription. Called by the service worker's
`pushsubscriptionchange` handler (no bearer token available there) and by the app on every
authenticated load (`syncPushSubscription` in `domain-client`).

```json
{ "oldEndpoint": "https://...", "endpoint": "https://...", "p256dh": "...", "auth": "..." }
```

If a row with `oldEndpoint` exists it's moved to the new `endpoint`/keys; otherwise the new
subscription is upserted. `204` · `400` if the new `endpoint`/`p256dh`/`auth` are missing.

Public because the SW is unauthenticated. It can only move/refresh a subscription — the
capability check is knowing the current push `endpoint` URL, which is a long high-entropy secret
issued by the push service. Fresh subscribes still require auth via the route above.

### `DELETE /api/notifications/subscriptions`

Body `{ "endpoint": "https://..." }`. Removes that subscription. `204` (missing endpoint is not
an error). The server also prunes any subscription a push attempt reports as `404`/`410`.

### `POST /api/notifications/test`

Sends a fixed "you're all set" notification through every enabled channel, so the user can
confirm a device + their email from the Settings screen. `202`.

## Subscription lifecycle

Push subscriptions are **independent of the login session**. Once registered they live in
`push_subscriptions` and the scheduler delivers to them with no user JWT involved, so an expired
token / "Leave Domain" does **not** stop notifications (only tapping one lands you on the
password gate). Delivery stops when: the user disables it in Settings (`DELETE` above), the PWA
is removed from the Home Screen, the push service reports the endpoint `404`/`410` (the server
prunes it on the next send), or the VAPID keys change. A rotated endpoint self-heals via the
`/sync` route (SW `pushsubscriptionchange` + the app's on-load `syncPushSubscription`).

## Delivery behaviour

`Notify` is best-effort per channel: a failing SMTP connection is logged and never blocks the
push, and vice-versa. Push payload delivered to the service worker's `push` handler:

```json
{ "title": "...", "body": "...", "url": "/todos", "tag": "morning-digest" }
```

## Morning digest

`internal/scheduler` ticks every minute. When the current IST wall-clock time is at/after
`settings.notifications.morningTime` and no `morning_digest` row exists for today's IST date, it
gathers not-done todos with `targetDate <= today` (due today or overdue), sends one notification,
and records the ledger row — recorded even when nothing is overdue, so the check doesn't re-run
all day.

## Reminders (`internal/reminder`)

Per-todo alarms. `reminders(id, todo_id, kind, fire_at, interval_seconds, created_at)`.

- **`kind = "once"`** — client sends an absolute RFC3339 `fireAt` (it computes "+15m" / "+2 days"
  itself). Fires once, then the row is deleted.
- **`kind = "repeat"`** — client sends only `intervalSeconds` (≥ 60); the server sets the first
  `fireAt` to now + interval and advances it after each fire (catching past any missed ticks so
  it doesn't fire in a burst).

The scheduler's due query is `reminders JOIN todos WHERE todos.done = FALSE AND fire_at <= now`,
so a repeating reminder **stops automatically the moment its todo is marked done**. Deleting a
todo cascades to its reminders.

### Endpoints

| method + path | body | result |
| --- | --- | --- |
| `GET /api/todos/{todoId}/reminders` | — | `200 → Reminder[]` (soonest first) |
| `POST /api/todos/{todoId}/reminders` | `{ kind, fireAt?, intervalSeconds? }` | `201 → Reminder` · `400` on bad input · `404` unknown todo |
| `DELETE /api/reminders/{id}` | — | `204` |

A fired reminder sends `{ title: "Reminder: <todo name>", url: "/todos", tag: "reminder-<todoId>" }`.
