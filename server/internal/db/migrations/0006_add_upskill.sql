CREATE TABLE IF NOT EXISTS upskill_topics (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    target_date TEXT,
    date_added  TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS upskill_subtopics (
    id          TEXT PRIMARY KEY,
    topic_id    TEXT NOT NULL REFERENCES upskill_topics(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    target_date TEXT,
    done        BOOLEAN NOT NULL DEFAULT FALSE,
    date_added  TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS upskill_resources (
    id          TEXT PRIMARY KEY,
    subtopic_id TEXT NOT NULL REFERENCES upskill_subtopics(id) ON DELETE CASCADE,
    label       TEXT,
    url         TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_upskill_subtopics_topic_id ON upskill_subtopics (topic_id);
CREATE INDEX IF NOT EXISTS idx_upskill_resources_subtopic_id ON upskill_resources (subtopic_id);
