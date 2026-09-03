-- Work Profile: user-created tabs (workstreams / Jira boards), each holding
-- todo-like tasks. Separate from `todos` because tasks are grouped by tab
-- and completing one is gated on a "logged in Jira?" acknowledgement.

CREATE TABLE IF NOT EXISTS work_tabs (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    position   INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS work_tasks (
    id                TEXT PRIMARY KEY,
    tab_id            TEXT NOT NULL REFERENCES work_tabs(id) ON DELETE CASCADE,
    name              TEXT NOT NULL,
    description       TEXT,
    target_date       TEXT,
    done              BOOLEAN NOT NULL DEFAULT FALSE,
    completed_at      TIMESTAMPTZ,
    jira_acknowledged BOOLEAN NOT NULL DEFAULT FALSE,
    created_at        TIMESTAMPTZ NOT NULL,

    -- Same invariant todos enforce (0016): a done task always carries a
    -- completion timestamp, a not-done one never does.
    CONSTRAINT work_tasks_done_completed_at_consistent CHECK (done = (completed_at IS NOT NULL))
);

CREATE INDEX IF NOT EXISTS idx_work_tasks_tab_id ON work_tasks (tab_id);
CREATE INDEX IF NOT EXISTS idx_work_tasks_target_date ON work_tasks (target_date);
