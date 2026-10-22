package jobs

import (
	"database/sql"
	"time"
)

func (j *Service) Expire(now time.Time) error {

	tx, e := j.S.DB.Begin()
	if e != nil {
		return e
	}
	defer tx.Rollback()
	// The event trigger labels this transaction's releases as expiries.
	if _, e = tx.Exec(`SET LOCAL queuelet.release='lease_expired'`); e != nil {
		return e
	}
	dead, e := release(tx, now, "expired", "lease expired", `state='leased' AND lease_expires_at <= $1 FOR UPDATE SKIP LOCKED`, now)
	if e != nil {
		return e
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

// ReleaseWorker ends a deregistered worker's leases inside the caller's
// transaction. The caller notifies the returned dead jobs after commit.
func (j *Service) ReleaseWorker(tx *sql.Tx, worker int64, now time.Time) ([]int64, error) {
	return release(tx, now, "released", "worker deleted", `state='leased' AND leased_by=$1 FOR UPDATE`, worker)
}

// release ends the selected leases. Each records an unsuccessful attempt; the
// job is queued again, or dead when its attempts are spent or its queue was
// purged while it ran.
func release(tx *sql.Tx, now time.Time, outcome, reason, where string, arg any) ([]int64, error) {
	rows, e := tx.Query(`SELECT id,attempts+1,max_attempts,leased_by,purge_requested FROM jobs WHERE `+where, arg)
	if e != nil {
		return nil, e
	}
	type expired struct {
		id, worker int64
		n, max     int
		purged     bool
	}
	batch := []expired{}
	for rows.Next() {
		var r expired
		if e = rows.Scan(&r.id, &r.n, &r.max, &r.worker, &r.purged); e != nil {
			rows.Close()
			return nil, e
		}
		batch = append(batch, r)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	dead := []int64{}
	for _, r := range batch {
		state := "queued"
		if r.n >= r.max || r.purged {
			state = "dead"
			dead = append(dead, r.id)
		}
		_, e = tx.Exec(`UPDATE jobs SET state=$2,event_at=$4,attempts=$3,last_error=$5,run_at=$4,lease_token=NULL,leased_by=NULL,lease_expires_at=NULL,finished_at=CASE WHEN $2='dead' THEN $4::timestamptz ELSE NULL END WHERE id=$1`, r.id, state, r.n, now, reason)
		if e != nil {
			return nil, e
		}
		_, e = tx.Exec(`INSERT INTO attempts SELECT $1,(SELECT count(*)+1 FROM attempts WHERE job_id=$1),$2,leased_at,$3,$4,$5 FROM jobs WHERE id=$1`, r.id, r.worker, now, outcome, reason)
		if e != nil {
			return nil, e
		}
	}
	return dead, nil
}

func (j *Service) Promote(now time.Time) error {
	_, e := j.S.DB.Exec(`UPDATE jobs SET state='queued',event_at=$1 WHERE state IN ('scheduled','retry_wait') AND run_at <= $1`, now)
	return e
}
func (j *Service) Retain(now time.Time) error {
	_, e := j.S.DB.Exec(`WITH removed AS (DELETE FROM jobs j USING queues q WHERE j.queue_id=q.id AND j.state IN ('succeeded','cancelled') AND extract(epoch FROM $1::timestamptz - j.finished_at) / 3600 > q.retention_hours RETURNING j.*) INSERT INTO job_events(job_id,namespace_id,at,event,queue,worker,attempt) SELECT r.id,r.namespace_id,$1,'retained',q.name,r.leased_by,r.attempts FROM removed r JOIN queues q ON q.id=r.queue_id`, now)
	return e
}
