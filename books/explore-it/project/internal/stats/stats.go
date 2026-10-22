package stats

import (
	"queuelet/internal/store"
	"time"
)

func Read(s *store.Store, where string, arg any, now time.Time) (map[string]any, error) {
	rows, e := s.Rows(`SELECT state,count(*) AS count FROM jobs WHERE `+where+`=$1 GROUP BY state`, arg)
	if e != nil {
		return nil, e
	}
	m := map[string]any{"queued": int64(0), "leased": int64(0), "scheduled": int64(0), "retry_wait": int64(0), "succeeded": int64(0), "dead": int64(0), "cancelled": int64(0)}
	for _, r := range rows {
		m[r["state"].(string)] = r["count"]
	}
	var age float64
	e = s.DB.QueryRow(`SELECT COALESCE(GREATEST(0,EXTRACT(EPOCH FROM ($2::timestamptz-min(GREATEST(created_at,run_at))))),0) FROM jobs WHERE `+where+`=$1 AND state='queued'`, arg, now).Scan(&age)
	m["oldest_queued_age"] = age
	var n int
	e2 := s.DB.QueryRow(`SELECT count(*) FROM jobs WHERE `+where+`=$1 AND state='succeeded' AND finished_at >= $2`, arg, now.Add(-time.Hour)).Scan(&n)
	m["throughput_last_hour"] = n
	if e != nil {
		return nil, e
	}
	return m, e2
}
