package jobs

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"queuelet/internal/store"
	"time"
)

type Service struct{ S *store.Store }

func Token() string { b := make([]byte, 16); rand.Read(b); return hex.EncodeToString(b) }
func Backoff(kind string, base, attempt int) time.Duration {
	if kind == "linear" {
		return time.Duration(base*attempt) * time.Second
	}
	return time.Duration(base) * time.Second * time.Duration(uint64(1)<<uint(attempt-1))
}

const LeaseOrder = "priority DESC, created_at, id"

func (j *Service) Lease(q, w int64, max int, now time.Time) ([]map[string]any, error) {
	tx, e := j.S.DB.Begin()
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	var sec int
	var paused bool
	if e = tx.QueryRow(`SELECT lease_seconds,paused FROM queues WHERE id=$1`, q).Scan(&sec, &paused); e != nil {
		return nil, e
	}
	if paused {
		return []map[string]any{}, nil
	}
	var kinds pq.StringArray
	if e = tx.QueryRow(`SELECT kinds FROM workers WHERE id=$1`, w).Scan(&kinds); e != nil {
		return nil, e
	}
	rows, e := tx.Query(`SELECT id FROM jobs WHERE queue_id=$1 AND state='queued' AND kind=ANY($2) ORDER BY `+LeaseOrder+` LIMIT $3 FOR UPDATE SKIP LOCKED`, q, kinds, max)
	if e != nil {
		return nil, e
	}
	ids := []int64{}
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		ids = append(ids, id)
	}
	rows.Close()
	out := []map[string]any{}
	for _, id := range ids {
		token := Token()
		var payload []byte
		var kind string
		var attempt int
		e = tx.QueryRow(`UPDATE jobs SET state='leased',leased_at=$5,event_at=$5,leased_by=$2,lease_token=$3,lease_expires_at=$4 WHERE id=$1 RETURNING kind,payload,attempts+1`, id, w, token, now.Add(time.Duration(sec)*time.Second), now).Scan(&kind, &payload, &attempt)
		if e != nil {
			return nil, e
		}
		out = append(out, map[string]any{"id": id, "kind": kind, "payload": json.RawMessage(payload), "attempt": attempt, "lease_token": token, "lease_expires_at": now.Add(time.Duration(sec) * time.Second)})
	}
	return out, tx.Commit()
}
func (j *Service) Notify(id int64, now time.Time) error {
	var url, secret, state, errtext string
	var result []byte
	e := j.S.DB.QueryRow(`SELECT q.webhook_url,q.webhook_secret,j.state,COALESCE(j.result,'null'),j.last_error FROM jobs j JOIN queues q ON q.id=j.queue_id WHERE j.id=$1`, id).Scan(&url, &secret, &state, &result, &errtext)
	if e != nil {
		return e
	}
	if url == "" {
		return nil
	}
	body, _ := json.Marshal(map[string]any{"job_id": id, "state": state, "result": json.RawMessage(result), "last_error": errtext})
	_, e = j.S.DB.Exec(`INSERT INTO webhook_outbox(job_id,url,secret,body,next_at,event_at,namespace_id,queue) SELECT $1,$2,$3,$4,$5,$5,j.namespace_id,q.name FROM jobs j JOIN queues q ON q.id=j.queue_id WHERE j.id=$1`, id, url, secret, string(body), now)
	return e
}
func (j *Service) Heartbeat(id int64, token string, seconds int, now time.Time) error {
	r, e := j.S.DB.Exec(`UPDATE jobs SET lease_expires_at=$3,event_at=$4 WHERE id=$1 AND lease_token=$2 AND state='leased'`, id, token, now.Add(time.Duration(seconds)*time.Second), now)
	return Changed(r, e)
}
func Changed(r sql.Result, e error) error {
	if e != nil {
		return e
	}
	n, e := r.RowsAffected()
	if e != nil {
		return e
	}
	if n == 0 {
		return fmt.Errorf("state conflict")
	}
	return nil
}
