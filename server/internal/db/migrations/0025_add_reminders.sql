-- Per-todo reminders. A one-time reminder fires once at fire_at and is
-- deleted; a repeating one carries interval_seconds and fire_at is advanced
-- after each fire. The scheduler's due query joins todos and skips done
-- ones, so a repeating reminder stops the moment its todo is completed.
CREATE TABLE IF NOT EXISTS reminders (
    id               TEXT PRIMARY KEY,
    todo_id          TEXT NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    kind             TEXT NOT NULL,
    fire_at          TIMESTAMPTZ NOT NULL,
    interval_seconds INTEGER,
    created_at       TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_reminders_fire_at ON reminders (fire_at);
CREATE INDEX IF NOT EXISTS idx_reminders_todo_id ON reminders (todo_id);
