-- Contract: drop the 32-bit counters, keep the 64-bit ones under the old
-- names. It ships with the release that reads and writes only the final
-- names, and refuses to run while anything still needs the old shape.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM databases WHERE bytes_total_v2 IS NULL OR rows_total_v2 IS NULL)
     OR EXISTS (SELECT 1 FROM usage_ledger WHERE bytes_v2 IS NULL OR rows_v2 IS NULL) THEN
    RAISE EXCEPTION 'contract refused: rows still to backfill (run control backfill)';
  END IF;
  IF EXISTS (SELECT 1 FROM nodes WHERE wire_version < 2 AND last_heartbeat_at > now() - interval '1 day') THEN
    RAISE EXCEPTION 'contract refused: a node has reported wire version 1 in the last day';
  END IF;
END $$;
ALTER TABLE databases DROP COLUMN bytes_total, DROP COLUMN rows_total;
ALTER TABLE databases RENAME COLUMN bytes_total_v2 TO bytes_total;
ALTER TABLE databases RENAME COLUMN rows_total_v2 TO rows_total;
ALTER TABLE databases ALTER COLUMN bytes_total SET NOT NULL, ALTER COLUMN rows_total SET NOT NULL;
ALTER TABLE usage_ledger DROP COLUMN bytes, DROP COLUMN rows;
ALTER TABLE usage_ledger RENAME COLUMN bytes_v2 TO bytes;
ALTER TABLE usage_ledger RENAME COLUMN rows_v2 TO rows;
ALTER TABLE usage_ledger ALTER COLUMN bytes SET NOT NULL, ALTER COLUMN rows SET NOT NULL;
