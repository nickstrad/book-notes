package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"queuelet/internal/jobs"
	"strconv"
	"time"
)

type JobInput struct {
	Kind           string          `json:"kind"`
	Payload        json.RawMessage `json:"payload"`
	Priority       *int            `json:"priority"`
	RunAt          *time.Time      `json:"run_at"`
	DelaySeconds   *int            `json:"delay_seconds"`
	MaxAttempts    *int            `json:"max_attempts"`
	IdempotencyKey *string         `json:"idempotency_key"`
}

type jobWriter interface{ QueryRow(string, ...any) *sql.Row }

func (s *Server) insertJob(q map[string]any, in JobInput, now time.Time) (map[string]any, error) {
	return insertJob(s.S.DB, q, in, now)
}
func insertJob(db jobWriter, q map[string]any, in JobInput, now time.Time) (map[string]any, error) {
	if in.Kind == "" {
		return nil, fmt.Errorf("kind required")
	}
	if in.RunAt != nil && in.DelaySeconds != nil {
		return nil, fmt.Errorf("run_at and delay_seconds are exclusive")
	}
	priority := 5
	if in.Priority != nil {
		priority = *in.Priority
	}
	if priority < 0 || priority > 9 {
		return nil, fmt.Errorf("priority must be 0–9")
	}
	max := int(q["max_attempts"].(int64))
	if in.MaxAttempts != nil {
		max = *in.MaxAttempts
	}
	if max < 1 {
		return nil, fmt.Errorf("max_attempts must be positive")
	}
	at := now
	if in.RunAt != nil {
		at = *in.RunAt
	}
	if in.DelaySeconds != nil {
		if *in.DelaySeconds > jobs.MaxSeconds {
			return nil, fmt.Errorf("delay_seconds must be at most %d", jobs.MaxSeconds)
		}
		at = now.Add(time.Duration(*in.DelaySeconds) * time.Second)
	}
	state := "queued"
	if at.After(now) {
		state = "scheduled"
	}
	if in.Payload == nil {
		in.Payload = json.RawMessage(`{}`)
	}
	in.Payload = bytes.TrimSpace(in.Payload)
	if len(in.Payload) == 0 || in.Payload[0] != '{' || !json.Valid(in.Payload) {
		return nil, fmt.Errorf("payload must be a JSON object")
	}
	b := in.Payload
	var id int64
	var status string
	e := db.QueryRow(`INSERT INTO jobs(queue_id,namespace_id,kind,payload,priority,state,run_at,max_attempts,idempotency_key,created_at,event_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$10) ON CONFLICT(queue_id,idempotency_key) WHERE idempotency_key IS NOT NULL DO NOTHING RETURNING id,state`, q["id"], q["namespace_id"], in.Kind, string(b), priority, state, at, max, in.IdempotencyKey, now).Scan(&id, &status)
	if e == sql.ErrNoRows && in.IdempotencyKey != nil {
		e = db.QueryRow(`SELECT id,state FROM jobs WHERE queue_id=$1 AND idempotency_key=$2`, q["id"], *in.IdempotencyKey).Scan(&id, &status)
	}
	return map[string]any{"id": id, "state": status}, e
}
func (s *Server) enqueue(w http.ResponseWriter, r *http.Request, q map[string]any) {
	var in JobInput
	if !decode(w, r, &in) {
		return
	}
	v, e := s.insertJob(q, in, s.Clock.Now())
	if e != nil {
		fail(w, 400, "invalid_job", e)
		return
	}
	write(w, 201, v)
}
func (s *Server) batch(w http.ResponseWriter, r *http.Request, q map[string]any) {
	var in struct {
		Jobs []JobInput `json:"jobs"`
	}
	if !decode(w, r, &in) {
		return
	}
	if len(in.Jobs) < 1 || len(in.Jobs) > 500 {
		fail(w, 400, "invalid_batch", "jobs must contain 1–500 elements")
		return
	}
	tx, e := s.S.DB.Begin()
	if e != nil {
		fail(w, 500, "database_error", e)
		return
	}
	defer tx.Rollback()
	ids := []any{}
	now := s.Clock.Now()
	for _, j := range in.Jobs {
		v, e := insertJob(tx, q, j, now)
		if e != nil {
			fail(w, 400, "invalid_job", e)
			return
		}
		ids = append(ids, v["id"])
	}
	if e = tx.Commit(); e != nil {
		fail(w, 500, "database_error", e)
		return
	}
	write(w, 201, map[string]any{"ids": ids})
}
func (s *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	q, ok := s.queueOrError(w, r)
	if !ok {
		return
	}
	state := r.URL.Query().Get("state")
	switch state {
	case "", "queued", "scheduled", "leased", "retry_wait", "succeeded", "dead", "cancelled":
	default:
		fail(w, 400, "invalid_state", "unknown state filter")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 500 {
		limit = 100
	}
	cursor, _ := strconv.ParseInt(r.URL.Query().Get("cursor"), 10, 64)
	v, e := s.S.Rows(`SELECT * FROM jobs WHERE queue_id=$1 AND ($2='' OR state=$2) AND ($3='' OR kind=$3) AND id>$4 ORDER BY id LIMIT $5`, q["id"], state, r.URL.Query().Get("kind"), cursor, limit)
	var next any
	if len(v) == limit {
		next = v[len(v)-1]["id"]
	}
	output(w, map[string]any{"jobs": v, "next_cursor": next}, e)
}
