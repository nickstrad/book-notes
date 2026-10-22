package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type JobInput struct {
	Kind           string         `json:"kind"`
	Payload        map[string]any `json:"payload"`
	Priority       *int           `json:"priority"`
	RunAt          *time.Time     `json:"run_at"`
	DelaySeconds   *int           `json:"delay_seconds"`
	MaxAttempts    *int           `json:"max_attempts"`
	IdempotencyKey *string        `json:"idempotency_key"`
}

func (s *Server) insertJob(q map[string]any, in JobInput, now time.Time) (map[string]any, error) {
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
		at = now.Add(time.Duration(*in.DelaySeconds) * time.Second)
	}
	state := "queued"
	if at.After(now) {
		state = "scheduled"
	}
	if in.Payload == nil {
		in.Payload = map[string]any{}
	}
	b, _ := json.Marshal(in.Payload)
	v, e := s.S.One(`INSERT INTO jobs(queue_id,namespace_id,kind,payload,priority,state,run_at,max_attempts,idempotency_key,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT(queue_id,idempotency_key) WHERE idempotency_key IS NOT NULL DO NOTHING RETURNING id,state`, q["id"], q["namespace_id"], in.Kind, string(b), priority, state, at, max, in.IdempotencyKey, now)
	if e == sql.ErrNoRows && in.IdempotencyKey != nil {
		return s.S.One(`SELECT id,state FROM jobs WHERE queue_id=$1 AND idempotency_key=$2`, q["id"], *in.IdempotencyKey)
	}
	return v, e
}
func (s *Server) enqueue(w http.ResponseWriter, r *http.Request, q map[string]any) {
	var in JobInput
	if !decode(w, r, &in) {
		return
	}
	v, e := s.insertJob(q, in, time.Now())
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
	if len(in.Jobs) > 500 {
		in.Jobs = in.Jobs[:500]
	}
	ids := []any{}
	now := time.Now()
	for _, j := range in.Jobs {
		v, e := s.insertJob(q, j, now)
		if e != nil {
			fail(w, 400, "invalid_job", e)
			return
		}
		ids = append(ids, v["id"])
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
	case "queued", "scheduled", "leased", "retry_wait", "succeeded", "dead", "cancelled":
	default:
		state = ""
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
