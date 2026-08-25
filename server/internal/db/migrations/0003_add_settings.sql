CREATE TABLE IF NOT EXISTS settings (
    id                         TEXT PRIMARY KEY,
    total_work_hours_required DOUBLE PRECISION
);

INSERT INTO settings (id) VALUES ('singleton') ON CONFLICT (id) DO NOTHING;
