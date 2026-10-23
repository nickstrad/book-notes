package schedule

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"queuelet/internal/store"
	"strings"
	"time"
)

var parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

// ErrNeverMatches is returned for an expression that names no real date.
var ErrNeverMatches = errors.New("cron expression never matches")

// location accepts IANA identifiers only. time.LoadLocation also accepts ""
// (UTC) and "Local" (the server process's zone), which are not IANA names.
func location(zone string) (*time.Location, error) {
	if zone == "" || zone == "Local" {
		return nil, fmt.Errorf("unknown time zone %q", zone)
	}
	return time.LoadLocation(zone)
}

func Next(expression, zone string, offset int, after time.Time) (time.Time, error) {
	// Exactly five fields: the parser would also take a CRON_TZ= or TZ= prefix
	// whose zone overrides the schedule's timezone.
	if n := len(strings.Fields(expression)); n != 5 {
		return time.Time{}, fmt.Errorf("cron uses five fields, got %d", n)
	}
	s, e := parser.Parse(expression)
	if e != nil {
		return time.Time{}, e
	}
	loc, e := location(zone)
	if e != nil {
		return time.Time{}, e
	}
	if strings.HasPrefix(strings.Fields(expression)[1], "*") {
		// Wildcard hours follow elapsed time across offset changes.
		if next := s.Next(after.In(loc)); !next.IsZero() {
			return next, nil
		}
		return time.Time{}, ErrNeverMatches
	}
	// Fixed hours match local civil times: a time skipped by a transition runs once
	// at the end of the gap, and a repeated time runs at its first occurrence only.
	l := after.In(loc)
	wall := time.Date(l.Year(), l.Month(), l.Day(), l.Hour(), l.Minute(), l.Second(), l.Nanosecond(), time.UTC)
	for range 1500 {
		if wall = s.Next(wall); wall.IsZero() {
			return wall, ErrNeverMatches
		}
		if t := civil(wall, loc); t.After(after) {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("schedule cannot advance")
}

// civil returns the first instant at which loc shows the wall-clock time carried in
// wall's UTC fields, or the end of the gap when a transition skips that time.
func civil(wall time.Time, loc *time.Location) time.Time {
	var first time.Time
	for _, probe := range []time.Time{wall.Add(-48 * time.Hour), wall.Add(48 * time.Hour)} {
		_, offset := probe.In(loc).Zone()
		t := wall.Add(-time.Duration(offset) * time.Second)
		if l := t.In(loc); l.Hour() == wall.Hour() && l.Minute() == wall.Minute() && l.Day() == wall.Day() && (first.IsZero() || t.Before(first)) {
			first = t
		}
	}
	if first.IsZero() {
		_, offset := wall.Add(-48 * time.Hour).In(loc).Zone()
		first, _ = wall.Add(-time.Duration(offset) * time.Second).In(loc).ZoneBounds()
	}
	return first
}
func Offset(zone string, now time.Time) (int, error) {
	loc, e := location(zone)
	if e != nil {
		return 0, e
	}
	_, offset := now.In(loc).Zone()
	return offset, nil
}

// Fire runs every due schedule. One schedule that fails does not stop the
// others; its error names it and the tick still reports failure.
func Fire(s *store.Store, now time.Time) error {
	rows, e := s.Rows(`SELECT id FROM schedules WHERE enabled AND next_run_at <= $1 ORDER BY id`, now)
	if e != nil {
		return e
	}
	var failed []error
	for _, r := range rows {
		if e = fireOne(s, r["id"], now); e != nil {
			failed = append(failed, fmt.Errorf("schedule %v: %w", r["id"], e))
		}
	}
	return errors.Join(failed...)
}
func fireOne(s *store.Store, id any, now time.Time) error {
	tx, e := s.DB.Begin()
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var q, ns int64
	var kind, expression, zone string
	var payload []byte
	var due time.Time
	e = tx.QueryRow(`SELECT queue_id,namespace_id,kind,payload,cron,timezone,next_run_at FROM schedules WHERE id=$1 AND enabled AND next_run_at <= $2 FOR UPDATE`, id, now).Scan(&q, &ns, &kind, &payload, &expression, &zone, &due)
	if e == sql.ErrNoRows {
		return nil
	}
	if e != nil {
		return e
	}
	var paused bool
	var max int
	e = tx.QueryRow(`SELECT paused,max_attempts FROM queues WHERE id=$1`, q).Scan(&paused, &max)
	if e != nil {
		return e
	}
	next, e := Next(expression, zone, 0, now)
	if e != nil {
		return e
	}
	if !next.After(now) {
		return fmt.Errorf("schedule cannot advance")
	}
	// Advance even while paused. At most one due execution can catch up after downtime.
	if !paused {
		_, e = tx.Exec(`INSERT INTO jobs(queue_id,namespace_id,kind,payload,priority,state,run_at,max_attempts,created_at,event_at) VALUES($1,$2,$3,$4,5,'queued',$5,$6,$7,$7)`, q, ns, kind, string(payload), due, max, now)
		if e != nil {
			return e
		}
	}
	_, e = tx.Exec(`UPDATE schedules SET next_run_at=$2,last_run_at=CASE WHEN $3 THEN last_run_at ELSE $4 END WHERE id=$1`, id, next, paused, due)
	if e != nil {
		return e
	}
	return tx.Commit()
}
