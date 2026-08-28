-- The fitness food library became a single shared list instead of being
-- scoped to one cycle: it's now edited in Settings and reused by every
-- cycle's Protein tab. Drop fitness_foods.cycle_id (and its index).
--
-- Existing rows just lose their cycle association and become part of the
-- shared library. Protein logs still reference food_id and still carry
-- their own cycle_id, so intake history is unaffected.
DROP INDEX IF EXISTS idx_fitness_foods_cycle_id;
ALTER TABLE fitness_foods DROP COLUMN IF EXISTS cycle_id;
