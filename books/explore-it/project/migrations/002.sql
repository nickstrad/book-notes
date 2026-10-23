-- Preserve existing UTC timestamps while retaining fractional creation time.
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns
             WHERE table_name='jobs' AND column_name='created_at'
               AND data_type='timestamp without time zone') THEN
    ALTER TABLE jobs ALTER COLUMN created_at TYPE timestamptz(6)
      USING created_at AT TIME ZONE 'UTC';
  END IF;
  -- Existing orphan rows are retained for inspection; new writes/deletes must
  -- preserve ownership even when enqueue and owner deletion race.
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname='jobs_queue_owner') THEN
    ALTER TABLE jobs ADD CONSTRAINT jobs_queue_owner
      FOREIGN KEY (queue_id) REFERENCES queues(id) NOT VALID;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname='schedules_queue_owner') THEN
    ALTER TABLE schedules ADD CONSTRAINT schedules_queue_owner
      FOREIGN KEY (queue_id) REFERENCES queues(id) NOT VALID;
  END IF;
END $$;
