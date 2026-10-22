package jobs

import (
	"time"
)

func (j *Service) Nack(id int64, token, message string, retry bool, delay *int, now time.Time) error {
	var attempts, max, base int
	var kind string
	e := j.S.DB.QueryRow(`SELECT j.attempts+1,j.max_attempts,q.backoff_kind,q.backoff_base FROM jobs j JOIN queues q ON q.id=j.queue_id WHERE j.id=$1`, id).Scan(&attempts, &max, &kind, &base)
	if e != nil {
		return e
	}
	state := "retry_wait"
	if !retry || attempts > max {
		state = "dead"
	}
	wait := Backoff(kind, base, attempts)
	if delay != nil {
		wait = time.Duration(*delay) * time.Second
	}
	_, e = j.S.DB.Exec(`UPDATE jobs SET state=$2,attempts=$3,last_error=$4,run_at=$5,finished_at=CASE WHEN $2='dead' THEN $6::timestamptz ELSE NULL END WHERE id=$1`, id, state, attempts, message, now.Add(wait), now)
	if e != nil {
		return e
	}
	_, e = j.S.DB.Exec(`INSERT INTO attempts SELECT id,attempts,leased_by,$2,$2,'failed',$3 FROM jobs WHERE id=$1`, id, now, message)
	if e != nil {
		return e
	}
	if state == "dead" {
		return j.Notify(id, now)
	}
	return nil
}
