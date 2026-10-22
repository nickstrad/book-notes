package api

import (
	"encoding/json"
	"net/http"
	"queuelet/internal/schedule"
	"time"
)

type ScheduleInput struct {
	Name     string         `json:"name"`
	Queue    string         `json:"queue"`
	Kind     string         `json:"kind"`
	Payload  map[string]any `json:"payload"`
	Cron     string         `json:"cron"`
	Timezone string         `json:"timezone"`
	Enabled  *bool          `json:"enabled"`
}

func (s *Server) createSchedule(w http.ResponseWriter, r *http.Request) {
	var in ScheduleInput
	if !decode(w, r, &in) {
		return
	}
	if in.Timezone == "" {
		in.Timezone = "UTC"
	}
	if in.Name == "" || in.Kind == "" {
		fail(w, 400, "invalid_schedule", "name and kind required")
		return
	}
	now := time.Now()
	off, e := schedule.Offset(in.Timezone, now)
	if e != nil {
		fail(w, 400, "invalid_timezone", e)
		return
	}
	next, e := schedule.Next(in.Cron, in.Timezone, off, now)
	if e != nil {
		fail(w, 400, "invalid_cron", e)
		return
	}
	q, e := s.S.One(`SELECT * FROM queues WHERE namespace_id=$1 AND name=$2`, ident(r).Namespace, in.Queue)
	if e != nil {
		fail(w, 404, "queue_not_found", e)
		return
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	var nextValue any = next
	if !enabled {
		nextValue = nil
	}
	v, e := s.S.One(`INSERT INTO schedules(namespace_id,queue_id,name,kind,payload,cron,timezone,zone_offset,enabled,next_run_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING *`, ident(r).Namespace, q["id"], in.Name, in.Kind, schedule.JSON(in.Payload), in.Cron, in.Timezone, off, enabled, nextValue)
	if e != nil {
		fail(w, 500, "database_error", e)
		return
	}
	write(w, 201, v)
}
func (s *Server) listSchedules(w http.ResponseWriter, r *http.Request) {
	v, e := s.S.Rows(`SELECT * FROM schedules WHERE namespace_id=$1 ORDER BY id`, ident(r).Namespace)
	output(w, map[string]any{"schedules": v}, e)
}
func (s *Server) getSchedule(w http.ResponseWriter, r *http.Request) {
	v, e := s.S.One(`SELECT * FROM schedules WHERE id=$1 AND namespace_id=$2`, num(r, "id"), ident(r).Namespace)
	if e != nil {
		fail(w, 404, "schedule_not_found", e)
		return
	}
	write(w, 200, v)
}
func (s *Server) patchSchedule(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Cron     *string `json:"cron"`
		Timezone *string `json:"timezone"`
		Enabled  *bool   `json:"enabled"`
	}
	if !decode(w, r, &in) {
		return
	}
	v, e := s.S.One(`UPDATE schedules SET cron=COALESCE($3,cron),timezone=COALESCE($4,timezone),enabled=COALESCE($5,enabled) WHERE id=$1 AND namespace_id=$2 RETURNING *`, num(r, "id"), ident(r).Namespace, in.Cron, in.Timezone, in.Enabled)
	if e != nil {
		fail(w, 404, "schedule_not_found", e)
		return
	}
	write(w, 200, v)
}
func (s *Server) deleteSchedule(w http.ResponseWriter, r *http.Request) {
	_, e := s.S.DB.Exec(`DELETE FROM schedules WHERE id=$1 AND namespace_id=$2`, num(r, "id"), ident(r).Namespace)
	if e != nil {
		fail(w, 500, "database_error", e)
		return
	}
	write(w, 204, nil)
}
func (s *Server) runSchedule(w http.ResponseWriter, r *http.Request) {
	v, e := s.S.One(`SELECT * FROM schedules WHERE id=$1 AND namespace_id=$2`, num(r, "id"), ident(r).Namespace)
	if e != nil {
		fail(w, 404, "schedule_not_found", e)
		return
	}
	q, e := s.S.One(`SELECT * FROM queues WHERE id=$1`, v["queue_id"])
	if e != nil {
		fail(w, 404, "queue_not_found", e)
		return
	}
	var in JobInput
	b, _ := json.Marshal(v["payload"])
	json.Unmarshal(b, &in.Payload)
	in.Kind = v["kind"].(string)
	job, e := s.insertJob(q, in, time.Now())
	if e != nil {
		fail(w, 400, "invalid_job", e)
		return
	}
	write(w, 201, job)
}
