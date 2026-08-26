CREATE TABLE IF NOT EXISTS spotify_tokens (
    id                      TEXT PRIMARY KEY,
    refresh_token_cipher    TEXT NOT NULL,
    access_token            TEXT NOT NULL DEFAULT '',
    access_token_expires_at TIMESTAMPTZ NOT NULL DEFAULT 'epoch',
    connected_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);
