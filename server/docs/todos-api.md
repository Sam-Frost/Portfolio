# Todos API

Backing API for the `domain-client` Todos page (`src/features/todos/`). Implemented as a
self-contained, feature-sliced package at `internal/todo/` (`model.go`, `repository.go` +
`repository_memory.go`, `service.go`, `handler.go`) rather than split across technical-layer
packages — see "Backend architecture" in the root `CLAUDE.md` for why.

Currently backed (`internal/todo/repository_memory.go`) by an in-process, in-memory store guarded
by a mutex — state resets on server restart. `todo.Repository` (`internal/todo/repository.go`) is
the persistence boundary; a future Postgres-backed implementation of that interface can be swapped
in without touching `service.go` or `handler.go`. `Update` takes a full `UpdateInput` rather than a
mutation closure specifically so a SQL implementation can build a real `SET` clause.

Error handling and JSON responses go through the shared `internal/httpx` and `internal/apperr`
packages (see CLAUDE.md) rather than being hand-rolled per handler.

## Auth

Same gate as the rest of the `/dashboard` domain area — no auth is enforced yet on either side.
When domain auth is implemented, these routes should require it; not addressed by this spec.

## Data model

```
Todo {
  id:          string   // hex-encoded random ID, server-generated
  name:        string   // required, non-empty
  description: string | null   // optional
  dateAdded:   string   // RFC3339 timestamp, server-set on creation, immutable
  targetDate:  string | null   // "YYYY-MM-DD", optional
  done:        boolean  // defaults to false on creation
  completedAt: string | null   // RFC3339 timestamp, server-set; non-null exactly when done is true
}
```

`completedAt` is owned entirely by the server: it's stamped with the current time
whenever `done` flips to `true` and cleared whenever `done` flips to `false`, so an
undo followed by a redo records a fresh timestamp rather than keeping the old one.
The Postgres schema enforces the `done ⇔ completedAt is not null` invariant with a
`CHECK` constraint (migration `0016`); pre-existing completed todos were backfilled
with their `dateAdded` as the completion date.

`name` sorting is done client-side (over whatever page is loaded); `dateAdded`, `targetDate`,
and `completedAt` sorting is done server-side via `GET /api/todos` query params, see below.

## Endpoints

### `GET /api/todos`

Returns all todos, sorted.

**Query params**
- `sortBy` — `dateAdded` (default) | `targetDate` | `completedAt`
- `order` — `desc` (default) | `asc`

Todos with `targetDate: null` always sort to the end, regardless of `order`, when
`sortBy=targetDate`; likewise todos with `completedAt: null` (i.e. not yet done)
when `sortBy=completedAt`. The domain-client only offers the `completedAt` sort on
the Completed tab.

**200**
```json
[
  {
    "id": "b1f2...",
    "name": "Write API spec for todos",
    "description": null,
    "dateAdded": "2026-08-20T09:00:00Z",
    "targetDate": null,
    "done": false,
    "completedAt": null
  }
]
```

### `POST /api/todos`

Creates a todo.

**Request**
```json
{
  "name": "Buy groceries",
  "description": "Milk, eggs, bread.",   // optional
  "targetDate": "2026-08-25"             // optional, "YYYY-MM-DD"
}
```

- `name` required, non-empty → `400` otherwise.
- `targetDate`, if present, must parse as `YYYY-MM-DD` → `400` otherwise.
- Server sets `id`, `dateAdded` (now), and `done: false`.

**201** → the created `Todo`.

### `GET /api/todos/count`

Returns the count of todos where `done` is `false`.

**200**
```json
{ "active": 3 }
```

### `PATCH /api/todos/{id}`

Partial update. Used today by the frontend only to toggle `done`, but accepts any editable field.

**Request** (any subset of)
```json
{ "name": "...", "description": "...", "targetDate": "...", "done": true }
```

- `dateAdded` is immutable — not accepted in the body.
- `name`, if present, must be non-empty (after trimming) → `400` otherwise.
- `targetDate`, if present, must parse as `YYYY-MM-DD` → `400` otherwise.
- `completedAt` is not accepted in the body — toggling `done` sets/clears it server-side.
- Unknown `id` → `404`.

**200** → the updated `Todo`.

### `DELETE /api/todos/{id}`

Deletes a todo.

- Unknown `id` → `404`.

**204**, empty body.

## Errors

Non-2xx responses body:
```json
{ "error": "human-readable message" }
```

| Status | When |
|---|---|
| 400 | Malformed body / missing or invalid field |
| 404 | `id` doesn't exist |
| 500 | Unhandled server error |

## Frontend usage today

`src/features/todos/api.ts` calls `GET /api/todos` (with `sortBy`/`order` when sorting by date
added or target date), `POST /api/todos` (the "add todo" box on the Todos page), `GET
/api/todos/count` (the "N active" line), `PATCH /api/todos/{id}` (toggling `done`), and `DELETE
/api/todos/{id}`.
