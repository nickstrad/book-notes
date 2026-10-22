package jobs

import "time"

func (j *Service) Expire(now time.Time) error {
	_, e := j.S.DB.Exec(`UPDATE jobs SET state='queued',lease_token=NULL,leased_by=NULL,lease_expires_at=NULL WHERE state='leased' AND lease_expires_at <= $1`, now)
	return e
}
func (j *Service) Promote(now time.Time) error {
	_, e := j.S.DB.Exec(`UPDATE jobs SET state='queued' WHERE state IN ('scheduled','retry_wait') AND run_at <= $1`, now)
	return e
}
func (j *Service) Retain(now time.Time) error {
	_, e := j.S.DB.Exec(`DELETE FROM jobs j USING queues q WHERE j.queue_id=q.id AND j.state IN ('succeeded','cancelled','dead') AND j.finished_at < $1::timestamptz - q.retention_hours * interval '1 hour'`, now)
	return e
}
