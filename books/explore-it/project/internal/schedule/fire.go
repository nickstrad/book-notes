package schedule

import (
	"database/sql"
	"fmt"
	"github.com/robfig/cron/v3"
	"queuelet/internal/store"
	"time"
)

var parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

func Next(expression, zone string, offset int, after time.Time) (time.Time, error) {
	s, e := parser.Parse(expression)
	if e != nil {
		return time.Time{}, e
	}
	loc, e := time.LoadLocation(zone)
	if e != nil {
		return time.Time{}, e
	}
	return s.Next(after.In(loc)), nil
}
func Offset(zone string, now time.Time) (int, error) {
	loc, e := time.LoadLocation(zone)
	if e != nil {
		return 0, e
	}
	_, offset := now.In(loc).Zone()
	return offset, nil
}
func Fire(s *store.Store, now time.Time) error {
	rows, e := s.Rows(`SELECT id FROM schedules WHERE enabled AND next_run_at <= $1 ORDER BY id`, now)
	if e != nil {
		return e
	}
	for _, r := range rows {
		if e = fireOne(s, r["id"], now); e != nil {
			return e
		}
	}
	return nil
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
		_, e = tx.Exec(`INSERT INTO jobs(queue_id,namespace_id,kind,payload,priority,state,run_at,max_attempts,created_at) VALUES($1,$2,$3,$4,5,'queued',$5,$6,$7)`, q, ns, kind, string(payload), due, max, now)
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
