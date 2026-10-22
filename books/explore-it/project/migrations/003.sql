-- Observation only: these triggers never choose or reject a transition.
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS event_at timestamptz;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS leased_at timestamptz;
CREATE TABLE IF NOT EXISTS job_events(
 id bigserial PRIMARY KEY, job_id bigint NOT NULL, namespace_id bigint NOT NULL,
 at timestamptz, event text NOT NULL, queue text, worker bigint, attempt int,
 outbox_id bigint, status int, error text NOT NULL DEFAULT '', logged boolean NOT NULL DEFAULT false);
ALTER TABLE job_events ADD COLUMN IF NOT EXISTS logged boolean NOT NULL DEFAULT false;
CREATE INDEX IF NOT EXISTS job_events_job ON job_events(namespace_id,job_id,id);
ALTER TABLE webhook_outbox ADD COLUMN IF NOT EXISTS namespace_id bigint;
ALTER TABLE webhook_outbox ADD COLUMN IF NOT EXISTS queue text;
ALTER TABLE webhook_outbox ADD COLUMN IF NOT EXISTS event_at timestamptz;
ALTER TABLE webhook_outbox ADD COLUMN IF NOT EXISTS status int;
ALTER TABLE webhook_outbox ADD COLUMN IF NOT EXISTS error text NOT NULL DEFAULT '';
UPDATE webhook_outbox o SET namespace_id=j.namespace_id,queue=q.name
 FROM jobs j JOIN queues q ON q.id=j.queue_id WHERE o.job_id=j.id AND o.namespace_id IS NULL;
CREATE OR REPLACE FUNCTION observe_job() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE label text;
BEGIN
 IF TG_OP='INSERT' THEN label:='created';
 ELSIF NEW.state IS DISTINCT FROM OLD.state THEN label:=CASE WHEN OLD.state='leased' AND current_setting('queuelet.release',true)='lease_expired' THEN 'lease_expired' ELSE NEW.state END;
 ELSIF NEW.lease_expires_at IS DISTINCT FROM OLD.lease_expires_at THEN label:='heartbeat';
 ELSE RETURN NEW; END IF;
 INSERT INTO job_events(job_id,namespace_id,at,event,queue,worker,attempt,error)
 SELECT NEW.id,NEW.namespace_id,NEW.event_at,label,q.name,
 COALESCE(NEW.leased_by,CASE WHEN TG_OP='UPDATE' THEN OLD.leased_by END),
 NEW.attempts+CASE WHEN NEW.state='leased' THEN 1 ELSE 0 END,NEW.last_error
 FROM queues q WHERE q.id=NEW.queue_id;
 RETURN NEW;
END $$;
DROP TRIGGER IF EXISTS observe_job ON jobs;
CREATE TRIGGER observe_job AFTER INSERT OR UPDATE ON jobs FOR EACH ROW EXECUTE FUNCTION observe_job();
CREATE OR REPLACE FUNCTION observe_notification() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF NEW.namespace_id IS NOT NULL AND (TG_OP='INSERT' OR NEW.attempts IS DISTINCT FROM OLD.attempts) THEN
 INSERT INTO job_events(job_id,namespace_id,at,event,queue,attempt,outbox_id,status,error)
 VALUES(NEW.job_id,NEW.namespace_id,NEW.event_at,
 CASE WHEN TG_OP='INSERT' THEN 'notification_queued' WHEN NEW.delivered_at IS NULL THEN 'notification_failed' ELSE 'notification_delivered' END,
 NEW.queue,NEW.attempts,NEW.id,NEW.status,NEW.error);
 END IF;
 RETURN NEW;
END $$;
DROP TRIGGER IF EXISTS observe_notification ON webhook_outbox;
CREATE TRIGGER observe_notification AFTER INSERT OR UPDATE ON webhook_outbox FOR EACH ROW EXECUTE FUNCTION observe_notification();
GRANT SELECT ON job_events TO explorer;
