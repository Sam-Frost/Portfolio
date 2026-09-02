-- A diary day can now hold video log entries alongside (or instead of) the
-- written entry in diary_entries. Videos are their own rows — several per
-- day are allowed — and are uploaded straight from the browser to the blob
-- store via a multipart upload, so a row exists in 'pending' state (holding
-- the S3 upload_id) before its bytes have all landed and been confirmed.
CREATE TABLE IF NOT EXISTS diary_videos (
    id               TEXT PRIMARY KEY,
    entry_date       DATE NOT NULL,
    title            TEXT,
    s3_key           TEXT NOT NULL,
    upload_id        TEXT,
    content_type     TEXT NOT NULL,
    size_bytes       BIGINT NOT NULL DEFAULT 0,
    duration_seconds INTEGER,
    status           TEXT NOT NULL DEFAULT 'pending',
    created_at       TIMESTAMPTZ NOT NULL,
    uploaded_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_diary_videos_entry_date ON diary_videos (entry_date);
