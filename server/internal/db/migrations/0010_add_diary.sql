CREATE TABLE IF NOT EXISTS diary_entries (
    id         TEXT PRIMARY KEY,
    entry_date DATE NOT NULL UNIQUE,
    content    TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
