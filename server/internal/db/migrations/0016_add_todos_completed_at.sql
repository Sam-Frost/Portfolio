ALTER TABLE todos ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ;

-- Backfill existing completed todos: no completion timestamp was ever recorded,
-- so fall back to date_added as the "done on" date for old entries.
UPDATE todos SET completed_at = date_added WHERE done = TRUE AND completed_at IS NULL;

-- Enforce it at the database level: a done todo must always carry a completion
-- timestamp, and a not-done todo must not. Undo/redo of a todo clears and re-sets
-- completed_at, so this invariant holds through every transition.
ALTER TABLE todos DROP CONSTRAINT IF EXISTS todos_done_completed_at_consistent;
ALTER TABLE todos ADD CONSTRAINT todos_done_completed_at_consistent
    CHECK (done = (completed_at IS NOT NULL));

CREATE INDEX IF NOT EXISTS idx_todos_completed_at ON todos (completed_at);
