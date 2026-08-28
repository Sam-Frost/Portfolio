-- Fitness tracking: a "cycle" (like a project) owns exercises, weight logs,
-- a food library, and protein logs. Exactly one cycle is active at a time
-- (partial unique index below); starting a new one archives the previous.
--
-- NOTE: 0015 later drops fitness_foods.cycle_id — the food library became
-- a single shared list (edited in Settings, reused by every cycle).
CREATE TABLE IF NOT EXISTS fitness_cycles (
    id             TEXT PRIMARY KEY,
    name           TEXT NOT NULL,
    start_date     TEXT NOT NULL,
    weight_start   DOUBLE PRECISION,
    weight_target  DOUBLE PRECISION,
    protein_target DOUBLE PRECISION,
    status         TEXT NOT NULL DEFAULT 'active',
    created_at     TIMESTAMPTZ NOT NULL,
    archived_at    TIMESTAMPTZ
);

-- At most one active cycle. A partial unique index leaves archived rows
-- (status = 'archived') unconstrained.
CREATE UNIQUE INDEX IF NOT EXISTS idx_fitness_cycles_active
    ON fitness_cycles (status) WHERE status = 'active';

CREATE TABLE IF NOT EXISTS fitness_exercises (
    id            TEXT PRIMARY KEY,
    cycle_id      TEXT NOT NULL REFERENCES fitness_cycles(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    goal_date     TEXT,
    goal_quantity DOUBLE PRECISION,
    unit          TEXT,
    created_at    TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS fitness_exercise_logs (
    id          TEXT PRIMARY KEY,
    exercise_id TEXT NOT NULL REFERENCES fitness_exercises(id) ON DELETE CASCADE,
    log_date    TEXT NOT NULL,
    quantity    DOUBLE PRECISION NOT NULL,
    UNIQUE (exercise_id, log_date)
);

CREATE TABLE IF NOT EXISTS fitness_weight_logs (
    id       TEXT PRIMARY KEY,
    cycle_id TEXT NOT NULL REFERENCES fitness_cycles(id) ON DELETE CASCADE,
    log_date TEXT NOT NULL,
    weight   DOUBLE PRECISION NOT NULL,
    UNIQUE (cycle_id, log_date)
);

CREATE TABLE IF NOT EXISTS fitness_foods (
    id               TEXT PRIMARY KEY,
    cycle_id         TEXT NOT NULL REFERENCES fitness_cycles(id) ON DELETE CASCADE,
    name             TEXT NOT NULL,
    unit             TEXT NOT NULL,
    protein_per_unit DOUBLE PRECISION NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS fitness_protein_logs (
    id         TEXT PRIMARY KEY,
    cycle_id   TEXT NOT NULL REFERENCES fitness_cycles(id) ON DELETE CASCADE,
    food_id    TEXT NOT NULL REFERENCES fitness_foods(id) ON DELETE CASCADE,
    log_date   TEXT NOT NULL,
    quantity   DOUBLE PRECISION NOT NULL,
    protein    DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_fitness_exercises_cycle_id ON fitness_exercises (cycle_id);
CREATE INDEX IF NOT EXISTS idx_fitness_exercise_logs_exercise_id ON fitness_exercise_logs (exercise_id);
CREATE INDEX IF NOT EXISTS idx_fitness_weight_logs_cycle_id ON fitness_weight_logs (cycle_id);
CREATE INDEX IF NOT EXISTS idx_fitness_foods_cycle_id ON fitness_foods (cycle_id);
CREATE INDEX IF NOT EXISTS idx_fitness_protein_logs_cycle_date ON fitness_protein_logs (cycle_id, log_date);
