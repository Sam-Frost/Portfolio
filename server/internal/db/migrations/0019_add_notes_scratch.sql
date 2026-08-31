ALTER TABLE notes ADD COLUMN IF NOT EXISTS scratch BOOLEAN NOT NULL DEFAULT FALSE;

-- At most one live scratch note exists at a time: it's the singleton
-- "Random Notepad" jot buffer, get-or-created on first open and never shown
-- in the notes list. The partial unique index enforces that invariant even
-- under a concurrent get-or-create race.
CREATE UNIQUE INDEX IF NOT EXISTS idx_notes_single_scratch
    ON notes (scratch) WHERE scratch AND deleted_at IS NULL;
