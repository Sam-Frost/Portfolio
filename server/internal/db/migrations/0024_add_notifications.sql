-- Notifications: outbound email + Web Push. The user is reachable when the
-- domain PWA isn't open (morning todo digest, per-todo reminders).

-- Settings gains a notifications section (singleton row, like the rest of
-- the settings columns). recipient email is nil until set in the UI;
-- morning_time is the local (IST) HH:MM the daily overdue-todo digest goes
-- out; the two enable flags gate each channel independently.
ALTER TABLE settings
    ADD COLUMN IF NOT EXISTS notification_recipient_email TEXT,
    ADD COLUMN IF NOT EXISTS notification_morning_time    TEXT    NOT NULL DEFAULT '07:00',
    ADD COLUMN IF NOT EXISTS notification_email_enabled   BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS notification_push_enabled    BOOLEAN NOT NULL DEFAULT TRUE;

-- One row per browser Web Push subscription (endpoint is the unique key the
-- push service hands us; p256dh/auth are the client's encryption keys).
CREATE TABLE IF NOT EXISTS push_subscriptions (
    id         TEXT PRIMARY KEY,
    endpoint   TEXT NOT NULL UNIQUE,
    p256dh     TEXT NOT NULL,
    auth       TEXT NOT NULL,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL
);

-- Dedup ledger for scheduled sends that must happen at most once per IST
-- day (the morning digest). UNIQUE(kind, ist_date) makes a double-send
-- impossible even if the scheduler ever runs on more than one instance.
CREATE TABLE IF NOT EXISTS notification_log (
    id       TEXT PRIMARY KEY,
    kind     TEXT NOT NULL,
    ist_date TEXT NOT NULL,
    sent_at  TIMESTAMPTZ NOT NULL,
    UNIQUE (kind, ist_date)
);
