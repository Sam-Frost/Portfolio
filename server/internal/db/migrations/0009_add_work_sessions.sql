CREATE TABLE IF NOT EXISTS work_sessions (
    id              TEXT PRIMARY KEY,
    planned_minutes INTEGER NOT NULL,
    started_at      TIMESTAMPTZ NOT NULL,
    ended_at        TIMESTAMPTZ,
    status          TEXT NOT NULL,
    note            TEXT,
    actual_minutes  INTEGER
);

CREATE INDEX IF NOT EXISTS idx_work_sessions_started_at ON work_sessions (started_at);
CREATE INDEX IF NOT EXISTS idx_work_sessions_status ON work_sessions (status);
