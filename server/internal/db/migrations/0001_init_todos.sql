CREATE TABLE IF NOT EXISTS todos (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT,
    date_added  TIMESTAMPTZ NOT NULL,
    target_date TEXT,
    done        BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_todos_date_added ON todos (date_added);
CREATE INDEX IF NOT EXISTS idx_todos_target_date ON todos (target_date);
