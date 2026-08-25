ALTER TABLE settings
    ADD COLUMN IF NOT EXISTS time_left_goal_date TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS time_left_format TEXT NOT NULL DEFAULT 'weeks_days_time';
