# Settings API

Backing API for the `domain-client` Settings page (`src/features/settings/`). Layered under
`internal/handler` → `internal/service` → `internal/repository` per the standard project structure,
backed by Postgres.

Settings is a singleton resource — one row/document per (eventual) authenticated user, not a
collection. The Settings page renders one section per entry in the dashboard sidebar
(`src/components/DomainSidebar.tsx`); most sections have no configurable fields yet, so the object
only carries `dailyWorkTracker` today. New sections add a new top-level key here as they gain
settings, without changing the shape of existing keys.

## Auth

Same gate as the rest of the `/dashboard` domain area — no auth is enforced yet on either side.
When domain auth is implemented, these routes should require it, and settings should become scoped
per-user rather than a single global row; not addressed by this spec.

## Data model

```
Settings {
  dailyWorkTracker: {
    totalWorkHoursRequired: number | null   // hours/day, e.g. 8; null = not set
  }
  timeLeftClock: {
    goalDate: string | null                 // RFC3339; null = not set
    format: "weeks_days_time" | "days_time"
  }
  notifications: {
    recipientEmail: string | null           // where email notifications go; null = not set
    morningTime: string                     // local IST "HH:MM" for the daily digest (default "07:00")
    emailEnabled: boolean                    // default true
    pushEnabled: boolean                     // default true
  }
}
```

A `PATCH` may send any subset of the top-level sections; a section that's present replaces that
whole section (the fields within it are not individually optional). `notifications.recipientEmail`
sent as `null` or `""` clears it. See `docs/notifications-api.md` for how these drive delivery.

- `totalWorkHoursRequired` is optional and defaults to `null` (unset) until the user configures it
  on the Settings page.
- There's exactly one `Settings` object; it always exists (server seeds a default row with all
  fields `null` if none exists yet), so `GET` never 404s.

## Endpoints

### `GET /api/settings`

Returns the current settings object.

**200**
```json
{
  "dailyWorkTracker": {
    "totalWorkHoursRequired": null
  }
}
```

### `PATCH /api/settings`

Partial, deep-merged update — only the keys present in the body are changed. Used today by the
frontend to set `dailyWorkTracker.totalWorkHoursRequired`.

**Request**
```json
{
  "dailyWorkTracker": {
    "totalWorkHoursRequired": 8
  }
}
```

- `totalWorkHoursRequired`, if present, must be a number `>= 0` → `400` otherwise. `null` clears it.
- Section keys omitted from the body are left untouched (deep merge, not replace).

**200** → the full updated `Settings` object.

## Errors

Non-2xx responses body:
```json
{ "error": "human-readable message" }
```

| Status | When |
|---|---|
| 400 | Malformed body / invalid field value |
| 500 | Unhandled server error |

## Frontend usage today

`src/features/settings/api.ts` calls `GET /api/settings` on page load and `PATCH /api/settings`
whenever the "Total work hours required" field changes, sending just
`{ "dailyWorkTracker": { "totalWorkHoursRequired": <value> } }`. Every other section on the page is
rendered as an empty placeholder ("No settings yet.") until it grows real fields, at which point its
key should be added to the `Settings` object above the same way `dailyWorkTracker` was.
