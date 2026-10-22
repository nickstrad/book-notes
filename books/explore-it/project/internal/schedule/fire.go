package schedule

import (
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
	return s.Next(after.In(time.FixedZone(zone, offset))), nil
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
	rows, e := s.Rows(`SELECT * FROM schedules WHERE next_run_at <= $1 ORDER BY id`, now)
	if e != nil {
		return e
	}
	for _, r := range rows {
		q, e := s.One(`SELECT * FROM queues WHERE id=$1`, r["queue_id"])
		if e != nil {
			continue
		}
		if q["paused"].(bool) {
			continue
		}
		next := r["next_run_at"].(time.Time)
		for !next.After(now) {
			payload := r["payload"]
			b := JSON(payload)
			_, e = s.DB.Exec(`INSERT INTO jobs(queue_id,namespace_id,kind,payload,priority,state,run_at,max_attempts,created_at) VALUES($1,$2,$3,$4,5,'queued',$5,$6,$7)`, r["queue_id"], r["namespace_id"], r["kind"], b, next, q["max_attempts"], now)
			if e != nil {
				return e
			}
			following, e := Next(r["cron"].(string), r["timezone"].(string), int(r["zone_offset"].(int64)), next)
			if e != nil {
				return e
			}
			if !following.After(next) {
				return fmt.Errorf("schedule cannot advance")
			}
			_, e = s.DB.Exec(`UPDATE schedules SET last_run_at=$2,next_run_at=$3 WHERE id=$1`, r["id"], next, following)
			if e != nil {
				return e
			}
			next = following
		}
	}
	return nil
}
