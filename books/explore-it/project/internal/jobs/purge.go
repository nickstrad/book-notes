package jobs

import (
	"github.com/lib/pq"
	"time"
)

func (j *Service) Purge(queue int64, now time.Time) (int64, error) {
	tx, err := j.S.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	// Lock every pending or leased row first. A row that moves between pending
	// and leased while we wait still matches, and the statements below see its
	// committed state. No queue/global lock is needed. ReleaseWorker also locks
	// in id order; Lease and Expire skip locked jobs.
	rows, err := tx.Query(`SELECT id FROM jobs WHERE queue_id=$1 AND `+Guard("state", EvPurge)+` ORDER BY id FOR UPDATE`, queue)
	if err != nil {
		return 0, err
	}
	ids := []int64{}
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return 0, err
	}
	// Restrict both writes to the locked snapshot: concurrently inserted jobs
	// may legitimately belong after this purge. Capture the boundary via ids.
	if _, err = tx.Exec(`UPDATE jobs SET purge_requested=true WHERE id=ANY($1) AND `+Guard("state", EvPurge, Leased), pq.Array(ids)); err != nil {
		return 0, err
	}
	res, err := tx.Exec(`WITH removed AS (DELETE FROM jobs WHERE id=ANY($1) AND `+Guard("state", EvPurge, Gone)+` RETURNING *) INSERT INTO job_events(job_id,namespace_id,at,event,queue,worker,attempt) SELECT r.id,r.namespace_id,$2,'purged',q.name,r.leased_by,r.attempts FROM removed r JOIN queues q ON q.id=r.queue_id`, pq.Array(ids), now)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return n, tx.Commit()
}
