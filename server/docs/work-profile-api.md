# Work Profile API

Backing API for the domain-client **Work Profile** area (`src/features/work-profile/`) — a
separate space from Todos for work items, organised into user-created **tabs** (workstreams /
Jira boards).

Feature package: `internal/workprofile/` (owns both `Tab` and `Task` — tightly coupled, one
feature). Same bearer-JWT gate as the rest of `/api/*`.

## Why separate from `internal/todo`

Tasks are grouped by tab, and completing one is gated on a "was this logged in Jira?"
acknowledgement (`jira_acknowledged`) — enough divergence that overloading `todos` would have
meant conditional columns and branchy handlers. Tables: `work_tabs`, `work_tasks` (migration
0026). `work_tasks` carries the same `CHECK (done = (completed_at IS NOT NULL))` invariant as
`todos` (0016).

## Data model

```
Tab  { id, name, position, createdAt }
Task { id, tabId, name, description|null, targetDate|null ("YYYY-MM-DD"),
       done, completedAt|null, jiraAcknowledged, createdAt }
```

## The Jira gate

`PATCH /api/work/tasks/{id}` with `{ "done": true }` is rejected `400` unless the same body
also has `{ "jiraAcknowledged": true }`. The UI shows a confirm dialog that supplies both; the
rule is enforced server-side so it can't be raced. Undoing (`{ "done": false }`) clears both
`completedAt` and `jiraAcknowledged`, so re-completing prompts again.

## Endpoints

| method + path | body | result |
| --- | --- | --- |
| `GET /api/work/tabs` | — | `200 → Tab[]` (by position) |
| `POST /api/work/tabs` | `{ name }` | `201 → Tab` (appended to the end) |
| `PATCH /api/work/tabs/{id}` | `{ name?, position? }` | `200 → Tab` |
| `DELETE /api/work/tabs/{id}` | — | `204` (cascades to its tasks) |
| `GET /api/work/tabs/{id}/tasks` | — | `200 → Task[]` |
| `POST /api/work/tabs/{id}/tasks` | `{ name, description?, targetDate? }` | `201 → Task` |
| `PATCH /api/work/tasks/{id}` | any of `{ name, description, targetDate, done, jiraAcknowledged }` | `200 → Task` |
| `DELETE /api/work/tasks/{id}` | — | `204` |
| `GET /api/work/overview` | — | `200 → { dueToday: Task[], overdue: Task[] }` |

`GET /api/work/overview` buckets every **open** task across all tabs by IST calendar date; each
row carries an extra `tabName`. Tasks with no `targetDate` are omitted.
