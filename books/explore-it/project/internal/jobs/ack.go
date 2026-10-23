package jobs

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

func (j *Service) Ack(id int64, token string, result any, now time.Time) error {
	b, _ := json.Marshal(result)
	// The joined row is the pre-update snapshot: it still names the holder.
	tx, e := j.S.DB.Begin()
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var worker int64
	e = tx.QueryRow(`UPDATE jobs j SET state='succeeded',event_at=$3,result=$2,finished_at=$3,attempts=j.attempts+1,lease_token=NULL,leased_by=NULL,lease_expires_at=NULL FROM jobs o WHERE o.id=j.id AND j.id=$1 AND `+Guard("j.state", EvAck)+` AND j.lease_token=$4 AND j.lease_expires_at > $3 RETURNING o.leased_by`, id, string(b), now, token).Scan(&worker)
	if e == sql.ErrNoRows {
		return fmt.Errorf("state conflict")
	}
	if e != nil {
		return e
	}
	_, e = tx.Exec(`INSERT INTO attempts SELECT id,(SELECT count(*)+1 FROM attempts WHERE job_id=$1),$3,leased_at,$2,'succeeded','' FROM jobs WHERE id=$1`, id, now, worker)
	if e != nil {
		return e
	}
	if e = notify(tx, id, now); e != nil {
		return e
	}
	return tx.Commit()
}
