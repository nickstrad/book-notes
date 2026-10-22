package jobs

import "time"

func (j *Service) Expire(now time.Time) error {

	tx, e := j.S.DB.Begin()
	if e != nil {
		return e
	}
	defer tx.Rollback()
	rows, e := tx.Query(`SELECT id,attempts+1,max_attempts,leased_by FROM jobs WHERE state='leased' AND lease_expires_at <= $1 FOR UPDATE SKIP LOCKED`, now)
	if e != nil {
		return e
	}
	type expired struct {
		id, worker int64
		n, max     int
	}
	batch := []expired{}
	for rows.Next() {
		var r expired
		if e = rows.Scan(&r.id, &r.n, &r.max, &r.worker); e != nil {
			rows.Close()
			return e
		}
		batch = append(batch, r)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return e
	}
	dead := []int64{}
	for _, r := range batch {
		state := "queued"
		if r.n >= r.max {
			state = "dead"
			dead = append(dead, r.id)
		}
		_, e = tx.Exec(`UPDATE jobs SET state=$2,event_at=$4,attempts=$3,last_error='lease expired',run_at=$4,lease_token=NULL,leased_by=NULL,lease_expires_at=NULL,finished_at=CASE WHEN $2='dead' THEN $4::timestamptz ELSE NULL END WHERE id=$1`, r.id, state, r.n, now)
		if e != nil {
			return e
		}
		_, e = tx.Exec(`INSERT INTO attempts SELECT $1,(SELECT count(*)+1 FROM attempts WHERE job_id=$1),$2,leased_at,$3,'expired','lease expired' FROM jobs WHERE id=$1`, r.id, r.worker, now)
		if e != nil {
			return e
		}
	}
	if e = tx.Commit(); e != nil {
		return e
	}
	for _, id := range dead {
		if e = j.Notify(id, now); e != nil {
			return e
		}
	}
	return nil
}

func (j *Service) Promote(now time.Time) error {
	_, e := j.S.DB.Exec(`UPDATE jobs SET state='queued',event_at=$1 WHERE state IN ('scheduled','retry_wait') AND run_at <= $1`, now)
	return e
}
func (j *Service) Retain(now time.Time) error {
	_, e := j.S.DB.Exec(`WITH removed AS (DELETE FROM jobs j USING queues q WHERE j.queue_id=q.id AND j.state IN ('succeeded','cancelled','dead') AND extract(epoch FROM $1::timestamptz - j.finished_at) / 3600 > q.retention_hours RETURNING j.*) INSERT INTO job_events(job_id,namespace_id,at,event,queue,worker,attempt) SELECT r.id,r.namespace_id,$1,'retained',q.name,r.leased_by,r.attempts FROM removed r JOIN queues q ON q.id=r.queue_id`, now)
	return e
}
