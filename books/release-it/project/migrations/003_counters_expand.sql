-- Expand: 64-bit counters next to the 32-bit ones (the whale's byte total
-- passes 2^31). Nullable columns without a default are a catalog change only:
-- no rewrite, no long lock. Control writes both and reads new-or-old;
-- `control backfill` fills the old rows; contract/004 drops the old ones.
ALTER TABLE databases ADD COLUMN bytes_total_v2 bigint, ADD COLUMN rows_total_v2 bigint;
ALTER TABLE usage_ledger ADD COLUMN bytes_v2 bigint, ADD COLUMN rows_v2 bigint;
-- Which wire version each node's heartbeats speak (the contract waits for 2).
ALTER TABLE nodes ADD COLUMN wire_version int NOT NULL DEFAULT 1;
