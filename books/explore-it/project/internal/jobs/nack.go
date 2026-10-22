package jobs

import "time"

func (j *Service) Nack(id int64, token, message string, retry bool, delay *int, now time.Time) error {
	tx, e := j.S.DB.Begin()
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var attempts, max, base int
	var kind string
	var worker int64
	e = tx.QueryRow(`SELECT j.attempts+1,j.max_attempts,q.backoff_kind,q.backoff_base,j.leased_by FROM jobs j JOIN queues q ON q.id=j.queue_id WHERE j.id=$1 AND j.state='leased' AND j.lease_token=$2 FOR UPDATE OF j`, id, token).Scan(&attempts, &max, &kind, &base, &worker)
	if e != nil {
		return e
	}
	state := "retry_wait"
	if !retry || attempts >= max {
		state = "dead"
	}
	wait := Backoff(kind, base, attempts)
	if delay != nil {
		wait = time.Duration(*delay) * time.Second
	}
	_, e = tx.Exec(`UPDATE jobs SET state=$2,event_at=$6,attempts=$3,last_error=$4,run_at=$5,finished_at=CASE WHEN $2='dead' THEN $6::timestamptz ELSE NULL END,lease_token=NULL,leased_by=NULL,lease_expires_at=NULL WHERE id=$1`, id, state, attempts, message, now.Add(wait), now)
	if e != nil {
		return e
	}
	_, e = tx.Exec(`INSERT INTO attempts SELECT $1,$2,$3,leased_at,$4,'failed',$5 FROM jobs WHERE id=$1`, id, attempts, worker, now, message)
	if e != nil {
		return e
	}
	if e = tx.Commit(); e != nil {
		return e
	}
	if state == "dead" {
		return j.Notify(id, now)
	}
	return nil
}
