package jobs

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

func (j *Service) Cancel(id int64, now time.Time) error {
	tx, e := j.S.DB.Begin()
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var state string
	e = tx.QueryRow(`SELECT state FROM jobs WHERE id=$1 AND state NOT IN ('succeeded','dead','cancelled') FOR UPDATE`, id).Scan(&state)
	if e == sql.ErrNoRows {
		return fmt.Errorf("state conflict")
	}
	if e != nil {
		return e
	}
	if state == "leased" {
		_, e = tx.Exec(`INSERT INTO attempts SELECT id,(SELECT count(*)+1 FROM attempts WHERE job_id=$1),leased_by,leased_at,$2,'cancelled','operator cancelled' FROM jobs WHERE id=$1`, id, now)
		if e != nil {
			return e
		}
	}
	_, e = tx.Exec(`UPDATE jobs SET state='cancelled',event_at=$2,finished_at=$2,attempts=attempts+CASE WHEN state='leased' THEN 1 ELSE 0 END,cancelled_lease_hash=encode(sha256(lease_token::bytea),'hex'),lease_token=NULL,leased_by=NULL,lease_expires_at=NULL WHERE id=$1`, id, now)
	if e != nil {
		return e
	}
	if e = tx.Commit(); e != nil {
		return e
	}
	return j.Notify(id, now)
}

// CancelRequested is an observation, never an extension of the revoked lease.
func (j *Service) CancelRequested(id int64, token string) (bool, error) {
	var requested bool
	digest := sha256.Sum256([]byte(token))
	e := j.S.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM jobs WHERE id=$1 AND state='cancelled' AND cancelled_lease_hash=$2)`, id, hex.EncodeToString(digest[:])).Scan(&requested)
	return requested, e
}

func (j *Service) Retry(id int64, now time.Time) error {
	r, e := j.S.DB.Exec(`UPDATE jobs SET state='queued',event_at=$2,attempts=0,run_at=$2,finished_at=NULL,purge_requested=false WHERE id=$1 AND state='dead'`, id, now)
	return Changed(r, e)
}
