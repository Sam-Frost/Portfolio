-- Notepad has its own label set, independent of internal/label and
-- internal/documentlabel.

CREATE TABLE IF NOT EXISTS notepad_labels (
    id    TEXT PRIMARY KEY,
    name  TEXT NOT NULL UNIQUE,
    color TEXT NOT NULL
);

ALTER TABLE notes ADD COLUMN IF NOT EXISTS label_id TEXT REFERENCES notepad_labels(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_notes_label_id ON notes (label_id);
