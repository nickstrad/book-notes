package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"queuelet/internal/jobs"
	"queuelet/internal/schedule"
	"queuelet/internal/store"
	"queuelet/internal/webhook"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	S    *store.Store
	Jobs *jobs.Service
}
type identity struct {
	Namespace int64
	Admin     bool
}
type authKey struct{}

func New(s *store.Store) *Server     { return &Server{s, &jobs.Service{S: s}} }
func ident(r *http.Request) identity { return r.Context().Value(authKey{}).(identity) }
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		json.NewEncoder(w).Encode(v)
	}
}
func fail(w http.ResponseWriter, status int, code string, e any) {
	write(w, status, map[string]any{"error": map[string]any{"code": code, "message": fmt.Sprint(e)}})
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20))
	if e := d.Decode(v); e != nil {
		fail(w, 400, "invalid_json", e)
		return false
	}
	return true
}
func num(r *http.Request, name string) int64 {
	v, _ := strconv.ParseInt(r.PathValue(name), 10, 64)
	return v
}
func (s *Server) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { write(w, 200, map[string]any{"status": "ok"}) })
	m.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "queuelet_up 1") })
	m.HandleFunc("POST /v1/queues", s.createQueue)
	m.HandleFunc("GET /v1/queues", s.listQueues)
	m.HandleFunc("GET /v1/queues/{q}", s.getQueue)
	m.HandleFunc("PATCH /v1/queues/{q}", s.patchQueue)
	m.HandleFunc("DELETE /v1/queues/{q}", s.deleteQueue)
	m.HandleFunc("POST /v1/queues/{q}/{verb}", s.queueVerb)
	m.HandleFunc("GET /v1/queues/{q}/jobs", s.listJobs)
	m.HandleFunc("GET /v1/queues/{q}/stats", s.queueStats)
	m.HandleFunc("GET /v1/jobs/{id}", s.getJob)
	m.HandleFunc("GET /v1/jobs/{id}/attempts", s.attempts)
	m.HandleFunc("DELETE /v1/jobs/{id}", s.deleteJob)
	m.HandleFunc("POST /v1/jobs/{id}/{verb}", s.jobVerb)
	m.HandleFunc("POST /v1/workers", s.createWorker)
	m.HandleFunc("DELETE /v1/workers/{id}", s.deleteWorker)
	m.HandleFunc("POST /v1/schedules", s.createSchedule)
	m.HandleFunc("GET /v1/schedules", s.listSchedules)
	m.HandleFunc("GET /v1/schedules/{id}", s.getSchedule)
	m.HandleFunc("PATCH /v1/schedules/{id}", s.patchSchedule)
	m.HandleFunc("DELETE /v1/schedules/{id}", s.deleteSchedule)
	m.HandleFunc("POST /v1/schedules/{id}/run", s.runSchedule)
	m.HandleFunc("GET /v1/namespaces/{ns}/stats", s.namespaceStats)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || r.URL.Path == "/metrics" {
			m.ServeHTTP(w, r)
			return
		}
		h := sha256.Sum256([]byte(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")))
		var id identity
		if e := s.S.DB.QueryRow(`SELECT namespace_id,admin FROM api_keys WHERE key_hash=$1`, hex.EncodeToString(h[:])).Scan(&id.Namespace, &id.Admin); e != nil {
			fail(w, 401, "unauthorized", "invalid API key")
			return
		}
		m.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authKey{}, id)))
	})
}
func (s *Server) Loop(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			for _, step := range []func(time.Time) error{s.Jobs.Promote, s.Jobs.Expire, func(t time.Time) error { return schedule.Fire(s.S, t) }, s.Jobs.Retain, func(t time.Time) error { return webhook.Ship(s.S, t) }} {
				if e := step(now); e != nil {
					log.Printf("background task: %v", e)
					return
				}
			}
		}
	}
}
func (s *Server) queue(r *http.Request) (map[string]any, error) {
	name := r.PathValue("q")
	if strings.ContainsAny(name, "/ ") {
		return nil, fmt.Errorf("invalid queue path")
	}
	return s.S.One(`SELECT * FROM queues WHERE namespace_id=$1 AND name=$2`, ident(r).Namespace, name)
}
func (s *Server) queueOrError(w http.ResponseWriter, r *http.Request) (map[string]any, bool) {
	q, e := s.queue(r)
	if e != nil {
		fail(w, 404, "queue_not_found", e)
		return nil, false
	}
	return q, true
}
func output(w http.ResponseWriter, v any, e error) {
	if e != nil {
		fail(w, 500, "database_error", e)
		return
	}
	write(w, 200, v)
}
