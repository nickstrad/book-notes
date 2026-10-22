package jobs

import "time"

func (j *Service) Cancel(id int64, now time.Time) error {
	r, e := j.S.DB.Exec(`UPDATE jobs SET state='cancelled',event_at=$2,finished_at=$2 WHERE id=$1 AND state NOT IN ('succeeded','dead','cancelled')`, id, now)
	if e = Changed(r, e); e != nil {
		return e
	}
	return j.Notify(id, now)
}
func (j *Service) Retry(id int64, now time.Time) error {
	r, e := j.S.DB.Exec(`UPDATE jobs SET state='queued',event_at=$2,attempts=0,run_at=$2,finished_at=NULL WHERE id=$1 AND state='dead'`, id, now)
	return Changed(r, e)
}
