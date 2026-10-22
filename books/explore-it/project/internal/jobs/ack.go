package jobs

import (
	"encoding/json"
	"time"
)

func (j *Service) Ack(id int64, token string, result any, now time.Time) error {
	b, _ := json.Marshal(result)
	r, e := j.S.DB.Exec(`UPDATE jobs SET state='succeeded',event_at=$3,result=$2,finished_at=$3,attempts=attempts+1 WHERE id=$1 AND state='leased' AND lease_token=$4`, id, string(b), now, token)
	if e = Changed(r, e); e != nil {
		return e
	}
	_, e = j.S.DB.Exec(`INSERT INTO attempts SELECT id,attempts,leased_by,leased_at,$2,'succeeded','' FROM jobs WHERE id=$1`, id, now)
	if e != nil {
		return e
	}
	return j.Notify(id, now)
}
