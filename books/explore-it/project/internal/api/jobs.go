package api

import (
	"database/sql"
	"github.com/lib/pq"
	"net/http"
	"queuelet/internal/jobs"
)

func (s *Server) getJob(w http.ResponseWriter, r *http.Request) {
	if !s.ownJob(w, r) {
		return
	}
	v, e := s.S.One(`SELECT * FROM jobs WHERE id=$1`, num(r, "id"))
	if e != nil {
		fail(w, 404, "job_not_found", e)
		return
	}
	write(w, 200, v)
}
func (s *Server) attempts(w http.ResponseWriter, r *http.Request) {
	if !s.ownJob(w, r) {
		return
	}
	v, e := s.S.Rows(`SELECT * FROM attempts WHERE job_id=$1 ORDER BY n`, num(r, "id"))
	output(w, map[string]any{"attempts": v}, e)
}
func (s *Server) deleteJob(w http.ResponseWriter, r *http.Request) {
	if !s.ownJob(w, r) {
		return
	}
	res, e := s.S.DB.Exec(`WITH removed AS (DELETE FROM jobs WHERE id=$1 AND state IN ('succeeded','dead','cancelled') RETURNING *) INSERT INTO job_events(job_id,namespace_id,at,event,queue,worker,attempt) SELECT r.id,r.namespace_id,$2,'deleted',q.name,r.leased_by,r.attempts FROM removed r JOIN queues q ON q.id=r.queue_id`, num(r, "id"), s.Clock.Now())
	if e = jobs.Changed(res, e); e != nil {
		fail(w, 409, "state_conflict", e)
		return
	}
	write(w, 204, nil)
}
func (s *Server) jobVerb(w http.ResponseWriter, r *http.Request) {
	if !s.ownJob(w, r) {
		return
	}
	id := num(r, "id")
	now := s.Clock.Now()
	var e error
	switch r.PathValue("verb") {
	case "cancel":
		e = s.Jobs.Cancel(id, now)
	case "retry":
		e = s.Jobs.Retry(id, now)
	case "ack", "nack", "heartbeat":
		var in struct {
			LeaseToken    string `json:"lease_token"`
			Result        any    `json:"result"`
			Error         string `json:"error"`
			Retry         *bool  `json:"retry"`
			DelaySeconds  *int   `json:"delay_seconds"`
			ExtendSeconds int    `json:"extend_seconds"`
		}
		if !decode(w, r, &in) {
			return
		}
		switch r.PathValue("verb") {
		case "ack":
			e = s.Jobs.Ack(id, in.LeaseToken, in.Result, now)
		case "nack":
			if in.DelaySeconds != nil && *in.DelaySeconds > jobs.MaxSeconds {
				fail(w, 400, "invalid_delay", "delay_seconds must be at most 2147483647")
				return
			}
			retry := true
			if in.Retry != nil {
				retry = *in.Retry
			}
			e = s.Jobs.Nack(id, in.LeaseToken, in.Error, retry, in.DelaySeconds, now)
		case "heartbeat":
			if in.ExtendSeconds < 1 || in.ExtendSeconds > jobs.MaxSeconds {
				fail(w, 400, "invalid_extension", "extend_seconds must be 1–2147483647")
				return
			}
			e = s.Jobs.Heartbeat(id, in.LeaseToken, in.ExtendSeconds, now)
		}
	default:
		fail(w, 404, "not_found", "unknown operation")
		return
	}
	if e != nil {
		fail(w, 409, "state_conflict", e)
		return
	}
	write(w, 200, map[string]any{"ok": true})
}
func (s *Server) createWorker(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name  string   `json:"name"`
		Kinds []string `json:"kinds"`
	}
	if !decode(w, r, &in) {
		return
	}
	if in.Name == "" || len(in.Kinds) == 0 {
		fail(w, 400, "invalid_worker", "name and kinds required")
		return
	}
	token := jobs.Token()
	var id int64
	e := s.S.DB.QueryRow(`INSERT INTO workers(namespace_id,name,kinds,token,last_seen_at) VALUES($1,$2,$3,$4,$5) RETURNING id`, ident(r).Namespace, in.Name, pq.Array(in.Kinds), token, s.Clock.Now()).Scan(&id)
	if e != nil {
		fail(w, 500, "database_error", e)
		return
	}
	write(w, 201, map[string]any{"worker_id": id, "token": token})
}
func (s *Server) deleteWorker(w http.ResponseWriter, r *http.Request) {
	tx, e := s.S.DB.Begin()
	if e != nil {
		fail(w, 500, "database_error", e)
		return
	}
	defer tx.Rollback()
	res, e := tx.Exec(`DELETE FROM workers WHERE id=$1 AND namespace_id=$2`, num(r, "id"), ident(r).Namespace)
	if e = jobs.Changed(res, e); e != nil {
		fail(w, 404, "worker_not_found", e)
		return
	}
	_, e = tx.Exec(`UPDATE jobs SET state='queued',event_at=$2,lease_token=NULL,leased_by=NULL,lease_expires_at=NULL WHERE leased_by=$1 AND state='leased'`, num(r, "id"), s.Clock.Now())
	if e == nil {
		e = tx.Commit()
	}
	output(w, map[string]any{"ok": true}, e)
}
func (s *Server) lease(w http.ResponseWriter, r *http.Request, q map[string]any) {
	var in struct {
		WorkerID int64 `json:"worker_id"`
		MaxJobs  int   `json:"max_jobs"`
	}
	in.MaxJobs = 1
	if !decode(w, r, &in) {
		return
	}
	if in.MaxJobs < 1 || in.MaxJobs > 100 {
		fail(w, 400, "invalid_batch", "max_jobs must be 1–100")
		return
	}
	var exists bool
	s.S.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM workers WHERE id=$1 AND namespace_id=$2)`, in.WorkerID, ident(r).Namespace).Scan(&exists)
	if !exists {
		fail(w, 404, "worker_not_found", "unknown worker")
		return
	}
	v, e := s.Jobs.Lease(q["id"].(int64), in.WorkerID, in.MaxJobs, s.Clock.Now())
	output(w, map[string]any{"jobs": v}, e)
}

func (s *Server) ownJob(w http.ResponseWriter, r *http.Request) bool {
	var id int64
	e := s.S.DB.QueryRow(`SELECT id FROM jobs WHERE id=$1 AND namespace_id=$2`, num(r, "id"), ident(r).Namespace).Scan(&id)
	if e == sql.ErrNoRows {
		fail(w, 404, "job_not_found", "job not found")
		return false
	}
	if e != nil {
		fail(w, 500, "database_error", e)
		return false
	}
	return true
}
