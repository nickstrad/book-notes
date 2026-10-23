package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	// Scope by retained event ownership, including after a job is deleted.
	v, e := s.S.Rows(`SELECT * FROM job_events WHERE job_id=$1 AND namespace_id=$2 ORDER BY id`, num(r, "id"), ident(r).Namespace)
	output(w, map[string]any{"events": v}, e)
}
func (s *Server) clockControl(w http.ResponseWriter, r *http.Request) {
	if !ident(r).Admin {
		fail(w, 403, "forbidden", "admin key required")
		return
	}
	var in struct {
		Action  string    `json:"action"`
		At      time.Time `json:"at"`
		Seconds float64   `json:"seconds"`
	}
	if !decode(w, r, &in) {
		return
	}
	_, e := s.Clock.Change(in.Action, in.At, in.Seconds)
	if e != nil {
		fail(w, 400, "invalid_clock", e)
		return
	}
	write(w, 200, map[string]any{"now": s.Clock.Now()})
}
func (s *Server) logEvents(ctx context.Context) {
	// Pending committed events survive restart. A crash between logging and marking
	// can replay an event; event_id allows consumers to deduplicate.
	t := time.NewTicker(200 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			rows, e := s.S.Rows(`SELECT * FROM job_events WHERE NOT logged ORDER BY id LIMIT 500`)
			if e != nil {
				slog.Error("event console", "error", e)
				continue
			}
			for _, r := range rows {
				slog.Info("job event", "event_id", r["id"], "job_id", r["job_id"], "queue", r["queue"], "event", r["event"], "worker", r["worker"], "attempt", r["attempt"], "outbox_id", r["outbox_id"], "status", r["status"], "error", r["error"], "at", r["at"])
				if _, e = s.S.DB.Exec(`UPDATE job_events SET logged=true WHERE id=$1`, r["id"]); e != nil {
					slog.Error("event console checkpoint", "error", e)
					break
				}
			}
		}
	}
}
