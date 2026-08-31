-- Work sessions gained three things: a personal/professional category
-- picked at start, a checklist of goal bullets (set at start, ticked
-- done/not-done at end), and free-text remarks captured separately at
-- start and at end (the old single `note` column becomes the end remark).
--
-- Existing rows predate the split: they're all treated as 'professional'
-- with no goals and no start remark; their `note` stays as the end remark.
ALTER TABLE work_sessions
    ADD COLUMN IF NOT EXISTS category   TEXT  NOT NULL DEFAULT 'professional',
    ADD COLUMN IF NOT EXISTS goals      JSONB NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS start_note TEXT;
