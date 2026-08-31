-- Locking a note puts its editor into a view-only mode. Notes are unlocked
-- by default; a note stays locked only once explicitly locked, until it's
-- explicitly unlocked again.
ALTER TABLE notes ADD COLUMN IF NOT EXISTS locked BOOLEAN NOT NULL DEFAULT FALSE;
