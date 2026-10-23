-- A job leased when its queue is purged finishes its attempt but is never pending again.
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS purge_requested boolean NOT NULL DEFAULT false;
