package api

import (
	"database/sql"
	"github.com/lib/pq"
	"net/http"
	"net/url"
	"queuelet/internal/jobs"
	"queuelet/internal/stats"
	"regexp"
	"strings"
)

type QueueInput struct {
	Name           string  `json:"name"`
	LeaseSeconds   int     `json:"lease_seconds"`
	MaxAttempts    int     `json:"max_attempts"`
	RetentionHours float64 `json:"retention_hours"`
	Backoff        struct {
		Kind        string `json:"kind"`
		BaseSeconds int    `json:"base_seconds"`
		MaxSeconds  *int   `json:"max_seconds"`
	} `json:"backoff"`
	Webhook struct {
		URL    string `json:"url"`
		Secret string `json:"secret"`
	} `json:"webhook"`
}

func (s *Server) createQueue(w http.ResponseWriter, r *http.Request) {
	var in QueueInput
	in.LeaseSeconds = 30
	in.MaxAttempts = 3
	in.RetentionHours = 24
	in.Backoff.Kind = "exponential"
	in.Backoff.BaseSeconds = 5
	if !decode(w, r, &in) {
		return
	}
	if !regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`).MatchString(in.Name) || !inRange(&in.LeaseSeconds, &in.MaxAttempts, &in.RetentionHours) || in.Backoff.BaseSeconds < 0 || in.Backoff.BaseSeconds > jobs.MaxSeconds || (in.Backoff.Kind != "linear" && in.Backoff.Kind != "exponential") {
		fail(w, 400, "invalid_queue", "name of 1–128 characters and limits within range required")
		return
	}
	if m := in.Backoff.MaxSeconds; m != nil && (*m < 1 || *m < in.Backoff.BaseSeconds || *m > jobs.MaxSeconds) {
		fail(w, 400, "invalid_queue", "backoff.max_seconds must be between base_seconds and 2147483647")
		return
	}
	if !validWebhook(in.Webhook.URL) {
		fail(w, 400, "invalid_webhook", "absolute HTTP(S) URL required")
		return
	}
	v, e := s.S.One(`INSERT INTO queues(namespace_id,name,lease_seconds,max_attempts,retention_hours,backoff_kind,backoff_base,backoff_max,webhook_url,webhook_secret) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING *`, ident(r).Namespace, in.Name, in.LeaseSeconds, in.MaxAttempts, in.RetentionHours, in.Backoff.Kind, in.Backoff.BaseSeconds, in.Backoff.MaxSeconds, in.Webhook.URL, in.Webhook.Secret)
	if pe, ok := e.(*pq.Error); ok && pe.Code == "23505" {
		fail(w, 409, "queue_conflict", "a queue with this name exists")
		return
	}
	if e != nil {
		fail(w, 500, "database_error", e)
		return
	}
	write(w, 201, v)
}
func (s *Server) listQueues(w http.ResponseWriter, r *http.Request) {
	v, e := s.S.Rows(`SELECT * FROM queues WHERE namespace_id=$1 ORDER BY id`, ident(r).Namespace)
	output(w, map[string]any{"queues": v}, e)
}
func (s *Server) getQueue(w http.ResponseWriter, r *http.Request) {
	q, ok := s.queueOrError(w, r)
	if ok {
		write(w, 200, q)
	}
}
func (s *Server) patchQueue(w http.ResponseWriter, r *http.Request) {
	q, ok := s.queueOrError(w, r)
	if !ok {
		return
	}
	var in struct {
		LeaseSeconds   *int     `json:"lease_seconds"`
		MaxAttempts    *int     `json:"max_attempts"`
		RetentionHours *float64 `json:"retention_hours"`
		Webhook        *struct {
			URL    *string `json:"url"`
			Secret *string `json:"secret"`
		} `json:"webhook"`
	}
	if !decode(w, r, &in) {
		return
	}
	if !inRange(in.LeaseSeconds, in.MaxAttempts, in.RetentionHours) {
		fail(w, 400, "invalid_queue", "limits out of range")
		return
	}
	// Webhook members merge like the top level: an omitted member keeps its value.
	url, secret := q["webhook_url"], q["webhook_secret"]
	if in.Webhook != nil && in.Webhook.URL != nil {
		if !validWebhook(*in.Webhook.URL) {
			fail(w, 400, "invalid_webhook", "absolute HTTP(S) URL required")
			return
		}
		url = *in.Webhook.URL
	}
	if in.Webhook != nil && in.Webhook.Secret != nil {
		secret = *in.Webhook.Secret
	}
	v, e := s.S.One(`UPDATE queues SET lease_seconds=COALESCE($2,lease_seconds),max_attempts=COALESCE($3,max_attempts),retention_hours=COALESCE($4,retention_hours),webhook_url=$5,webhook_secret=$6 WHERE id=$1 RETURNING *`, q["id"], in.LeaseSeconds, in.MaxAttempts, in.RetentionHours, url, secret)
	output(w, v, e)
}
func (s *Server) deleteQueue(w http.ResponseWriter, r *http.Request) {

	tx, e := s.S.DB.Begin()
	if e != nil {
		fail(w, 500, "database_error", e)
		return
	}
	defer tx.Rollback()
	// Serialize owner deletion with inserts; refuse rather than orphan dependents.
	if _, e = tx.Exec(`LOCK TABLE queues,jobs,schedules IN SHARE ROW EXCLUSIVE MODE`); e != nil {
		fail(w, 500, "database_error", e)
		return
	}
	var id int64
	e = tx.QueryRow(`SELECT id FROM queues WHERE namespace_id=$1 AND name=$2`, ident(r).Namespace, r.PathValue("q")).Scan(&id)
	if e == sql.ErrNoRows {
		fail(w, 404, "queue_not_found", "queue not found")
		return
	}
	if e != nil {
		fail(w, 500, "database_error", e)
		return
	}
	var nonempty bool
	e = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM jobs WHERE queue_id=$1) OR EXISTS(SELECT 1 FROM schedules WHERE queue_id=$1)`, id).Scan(&nonempty)
	if e != nil {
		fail(w, 500, "database_error", e)
		return
	}
	if nonempty {
		fail(w, 409, "queue_not_empty", "queue owns jobs or schedules")
		return
	}
	_, e = tx.Exec(`DELETE FROM queues WHERE id=$1`, id)
	if e == nil {
		e = tx.Commit()
	}
	if e != nil {
		fail(w, 500, "database_error", e)
		return
	}
	write(w, 204, nil)
}

func (s *Server) queueVerb(w http.ResponseWriter, r *http.Request) {
	q, ok := s.queueOrError(w, r)
	if !ok {
		return
	}
	switch r.PathValue("verb") {
	case "jobs":
		s.enqueue(w, r, q)
	case "jobs:batch":
		s.batch(w, r, q)
	case "lease":
		s.lease(w, r, q)
	case "pause", "resume":
		_, e := s.S.DB.Exec(`UPDATE queues SET paused=$2 WHERE id=$1`, q["id"], r.PathValue("verb") == "pause")
		output(w, map[string]any{"ok": true}, e)
	case "purge":
		n, e := s.Jobs.Purge(q["id"].(int64), s.Clock.Now())
		if e != nil {
			fail(w, 500, "database_error", e)
			return
		}
		write(w, 200, map[string]any{"deleted": n})
	default:
		fail(w, 404, "not_found", "unknown operation")
	}
}
func (s *Server) queueStats(w http.ResponseWriter, r *http.Request) {
	q, ok := s.queueOrError(w, r)
	if !ok {
		return
	}
	v, e := stats.Read(s.S, "queue_id", q["id"], s.Clock.Now())
	output(w, v, e)
}
func (s *Server) namespaceStats(w http.ResponseWriter, r *http.Request) {
	var id int64
	e := s.S.DB.QueryRow(`SELECT id FROM namespaces WHERE name=$1`, r.PathValue("ns")).Scan(&id)
	if e != nil {
		fail(w, 404, "namespace_not_found", e)
		return
	}
	if id != ident(r).Namespace && !ident(r).Admin {
		fail(w, 403, "forbidden", "namespace mismatch")
		return
	}
	v, e := stats.Read(s.S, "namespace_id", id, s.Clock.Now())
	output(w, v, e)
}

// inRange checks the queue limits that are present; nil means unchanged.
func inRange(leaseSeconds, maxAttempts *int, retentionHours *float64) bool {
	return (leaseSeconds == nil || *leaseSeconds >= 1 && *leaseSeconds <= jobs.MaxSeconds) &&
		(maxAttempts == nil || *maxAttempts >= 1 && *maxAttempts <= jobs.MaxSeconds) &&
		(retentionHours == nil || *retentionHours >= 0 && *retentionHours <= maxRetentionHours)
}

// maxRetentionHours (about 114 years) keeps every retention cutoff a valid timestamp.
const maxRetentionHours = 1_000_000

func validWebhook(raw string) bool {
	if raw == "" {
		return true
	} // Empty disables notifications for future jobs.
	u, e := url.Parse(raw)
	if e != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") || strings.ContainsAny(raw, " \t\r\n") {
		return false
	}
	_, e = http.NewRequest("POST", raw, nil)
	return e == nil
}
