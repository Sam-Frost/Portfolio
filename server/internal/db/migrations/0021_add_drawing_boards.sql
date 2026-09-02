CREATE TABLE IF NOT EXISTS drawing_boards (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    scene_data JSONB NOT NULL DEFAULT '{"elements":[],"appState":{},"files":{}}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ
);

-- Board list is sorted most-recently-edited first.
CREATE INDEX IF NOT EXISTS idx_drawing_boards_updated_at
    ON drawing_boards (updated_at DESC) WHERE deleted_at IS NULL;
