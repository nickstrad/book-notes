-- Attempts share the lifetime of their job. Legacy orphans remain inspectable;
-- NOT VALID still enforces new writes and cascades all subsequent job deletions.
DO $$ BEGIN
 IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname='attempts_job_owner' AND conrelid='attempts'::regclass) THEN
  ALTER TABLE attempts ADD CONSTRAINT attempts_job_owner
   FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE NOT VALID;
 END IF;
END $$;
CREATE INDEX IF NOT EXISTS attempts_job ON attempts(job_id);
