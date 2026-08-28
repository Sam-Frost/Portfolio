-- Document Storage: a gated file manager. Folders form a tree
-- (parent_id self-reference); documents hold only metadata here — the bytes
-- live in a blob store (S3 in prod, local disk in dev, see internal/document).
-- Labels are a document-specific set, independent of internal/label.

CREATE TABLE IF NOT EXISTS document_labels (
    id    TEXT PRIMARY KEY,
    name  TEXT NOT NULL UNIQUE,
    color TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS document_folders (
    id         TEXT PRIMARY KEY,
    parent_id  TEXT REFERENCES document_folders(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_document_folders_parent_id ON document_folders (parent_id);

-- A folder name is unique among its siblings (case-insensitive). Root-level
-- folders share the sentinel '' parent for the purposes of this index.
CREATE UNIQUE INDEX IF NOT EXISTS idx_document_folders_parent_name
    ON document_folders (COALESCE(parent_id, ''), lower(name));

CREATE TABLE IF NOT EXISTS documents (
    id           TEXT PRIMARY KEY,
    folder_id    TEXT REFERENCES document_folders(id) ON DELETE CASCADE,
    label_id     TEXT REFERENCES document_labels(id) ON DELETE SET NULL,
    name         TEXT NOT NULL,
    s3_key       TEXT NOT NULL UNIQUE,
    content_type TEXT NOT NULL DEFAULT '',
    size_bytes   BIGINT NOT NULL DEFAULT 0,
    -- 'pending' between the create call and the browser's upload+complete;
    -- 'ready' once the blob is confirmed in the store. Listings show 'ready'.
    status       TEXT NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    uploaded_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_documents_folder_id  ON documents (folder_id);
CREATE INDEX IF NOT EXISTS idx_documents_label_id   ON documents (label_id);
CREATE INDEX IF NOT EXISTS idx_documents_name_lower ON documents (lower(name));
