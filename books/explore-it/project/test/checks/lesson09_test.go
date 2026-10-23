//go:build checks

package checks

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"queuelet/internal/client"
	"queuelet/internal/jobs"
	"queuelet/internal/schedule"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The E09 reproductions freeze the lab clock, so they must not run in parallel.
func l09Absent(t *testing.T, id string) {
	t.Helper()
	cmd := exec.Command("bash", "lab/repro/"+id+".sh")
	cmd.Dir = "../.."
	out, e := cmd.CombinedOutput()
	if e != nil || !strings.Contains(string(out), "ABSENT: "+id+" — ") {
		t.Fatalf("%s regression: %v\n%s", id, e, out)
	}
	t.Log(strings.TrimSpace(string(out)))
}
func TestL09U39(t *testing.T) { l09Absent(t, "U39") }
func TestL09U40(t *testing.T) { l09Absent(t, "U40") }
func TestL09U41(t *testing.T) { l09Absent(t, "U41") }
func TestL09U42(t *testing.T) { l09Absent(t, "U42") }
func TestL09U43(t *testing.T) { l09Absent(t, "U43") }
func TestL09U44(t *testing.T) { l09Absent(t, "U44") }
func TestL09U45(t *testing.T) { l09Absent(t, "U45") }

// A schedule stored before U40 (never matching, next_run_at in year 1) must not
// stop a later schedule, and an operator must still be able to disable it.
// The legacy row can only be made with SQL; Fire runs in-process.
func TestL09LegacyUnadvanceableSchedule(t *testing.T) {
	s := l02Store(t)
	c := client.Env()
	name := fmt.Sprintf("l09-legacy-%d", time.Now().UnixNano())
	q := l07Num(l07Call(t, c, "POST", "/v1/queues", map[string]any{"name": name}, 201)["id"])
	var ns, stuck, victim int64
	if e := s.DB.QueryRow(`SELECT namespace_id FROM queues WHERE id=$1`, q).Scan(&ns); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2033, 6, 1, 0, 0, 30, 0, time.UTC)
	insert := `INSERT INTO schedules(namespace_id,queue_id,name,kind,payload,cron,timezone,zone_offset,enabled,next_run_at) VALUES($1,$2,$3,'manual','{}',$4,'UTC',0,true,$5) RETURNING id`
	if e := s.DB.QueryRow(insert, ns, q, name+"-stuck", "0 0 30 2 *", time.Time{}).Scan(&stuck); e != nil {
		t.Fatal(e)
	}
	if e := s.DB.QueryRow(insert, ns, q, name+"-victim", "* * * * *", now.Add(-30*time.Second)).Scan(&victim); e != nil {
		t.Fatal(e)
	}
	defer s.DB.Exec(`DELETE FROM jobs WHERE queue_id=$1`, q)
	defer s.DB.Exec(`DELETE FROM schedules WHERE id IN ($1,$2)`, stuck, victim)
	e := schedule.Fire(s, now)
	if e == nil || !strings.Contains(e.Error(), fmt.Sprintf("schedule %d", stuck)) {
		t.Fatalf("Fire must report the stuck schedule by id, got %v", e)
	}
	var fired int
	if e := s.DB.QueryRow(`SELECT count(*) FROM jobs WHERE queue_id=$1`, q).Scan(&fired); e != nil || fired != 1 {
		t.Fatalf("later schedule fired %d jobs (%v)", fired, e)
	}
	l07Call(t, c, "PATCH", fmt.Sprintf("/v1/schedules/%d", stuck), map[string]any{"enabled": false}, 200)
	l07Call(t, c, "PATCH", fmt.Sprintf("/v1/schedules/%d", stuck), map[string]any{"enabled": true}, 400)
}

// TestL09ModelWalk drives random events against the transition table: before
// each event it reads the job's state, asks jobs.Targets what the event may do,
// acts through the public API and requires the outcome the table predicts. Each
// slot is a queue holding at most one job, so a lease names its job. Background
// events (promote, expire, retain) are driven by advancing the frozen clock.
func TestL09ModelWalk(t *testing.T) {
	steps, seed := 1000, int64(9)
	if v, e := strconv.Atoi(os.Getenv("QUEUELET_WALK_STEPS")); e == nil {
		steps = v
	}
	if v, e := strconv.ParseInt(os.Getenv("QUEUELET_WALK_SEED"), 10, 64); e == nil {
		seed = v
	}
	rng := rand.New(rand.NewSource(seed))
	c := client.Env()
	admin := &client.Client{URL: c.URL, Key: "admin-key"}
	l07Call(t, admin, "POST", "/admin/clock", map[string]any{"action": "set", "at": "2033-03-01T00:00:00Z"}, 200)
	defer admin.Do("POST", "/admin/clock", map[string]any{"action": "reset"})
	prefix := fmt.Sprintf("l09-%d", time.Now().UnixNano())
	v := l07Call(t, c, "POST", "/v1/workers", map[string]any{"name": prefix, "kinds": []string{"walk"}}, 201)
	worker := l07Num(v["worker_id"])
	defer c.Do("DELETE", fmt.Sprintf("/v1/workers/%d", worker), nil)

	type slot struct {
		q     string
		job   int64 // 0: no job (gone)
		token string
	}
	slots := make([]*slot, 6)
	for i := range slots {
		s := &slot{q: fmt.Sprintf("%s-%d", prefix, i)}
		// Lease 30 s; a nack waits 20 s; retention 0 h removes finished rows once time moves.
		l07Call(t, c, "POST", "/v1/queues", map[string]any{"name": s.q, "lease_seconds": 30, "max_attempts": 3, "retention_hours": 0, "backoff": map[string]any{"kind": "linear", "base_seconds": 20}}, 201)
		slots[i] = s
	}
	read := func(s *slot) (string, map[string]any) {
		if s.job == 0 {
			return jobs.Gone, nil
		}
		raw, status, err := c.Do("GET", fmt.Sprintf("/v1/jobs/%d", s.job), nil)
		if err != nil {
			t.Fatal(err)
		}
		if status == 404 {
			return jobs.Gone, nil
		}
		var row map[string]any
		if status != 200 || client.DecodeJSON(raw, &row) != nil {
			t.Fatalf("GET job %d: %d %s", s.job, status, raw)
		}
		return row["state"].(string), row
	}
	stamp := func(row map[string]any, k string) time.Time {
		at, _ := row[k].(string)
		tm, _ := time.Parse(time.RFC3339Nano, at)
		return tm
	}
	events := []string{"lease", "lease", "lease", "heartbeat", "ack", "ack", "nack", "nack", "nack_dead", "cancel", "retry", "retry", "delete", "purge", "advance", "advance"}
	counts := map[string]int{}
	now := time.Date(2033, 3, 1, 0, 0, 0, 0, time.UTC)
	for step := 1; step <= steps; step++ {
		s := slots[rng.Intn(len(slots))]
		before, _ := read(s)
		if before == jobs.Gone {
			body := map[string]any{"kind": "walk", "payload": map[string]any{"step": step}}
			if rng.Intn(2) == 0 {
				body["delay_seconds"] = 20
			}
			v := l07Call(t, c, "POST", "/v1/queues/"+s.q+"/jobs", body, 201)
			s.job, s.token = l07Num(v["id"]), ""
			counts["enqueue"]++
			continue
		}
		name := events[rng.Intn(len(events))]
		counts[name]++
		if name == "advance" {
			// Every slot's background event is predicted from its row before time moves.
			now = now.Add(40 * time.Second)
			type expect struct {
				from string
				ev   jobs.Event
				must bool
			}
			want := map[*slot]expect{}
			for _, o := range slots {
				st, row := read(o)
				switch {
				case st == jobs.Leased:
					want[o] = expect{st, jobs.EvExpire, !stamp(row, "lease_expires_at").After(now)}
				case st == jobs.Scheduled || st == jobs.RetryWait:
					want[o] = expect{st, jobs.EvPromote, !stamp(row, "run_at").After(now)}
				case jobs.Targets(jobs.EvRetain, st) != nil:
					want[o] = expect{st, jobs.EvRetain, true}
				default:
					want[o] = expect{st, "", false}
				}
			}
			l07Call(t, admin, "POST", "/admin/clock", map[string]any{"action": "advance", "seconds": 40}, 200)
			deadline := time.Now().Add(10 * time.Second)
			for {
				settled, detail := true, ""
				for _, o := range slots {
					w := want[o]
					st, _ := read(o)
					moved := st != w.from
					if moved && (w.ev == "" || !slices.Contains(jobs.Targets(w.ev, w.from), st)) {
						t.Fatalf("step %d advance: job %d moved %s→%s; table allows %s → %v", step, o.job, w.from, st, w.ev, jobs.Targets(w.ev, w.from))
					}
					if w.must && !moved {
						settled, detail = false, fmt.Sprintf("job %d still %s after %s was due", o.job, st, w.ev)
					}
				}
				if settled {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("step %d advance: %s", step, detail)
				}
				time.Sleep(200 * time.Millisecond)
			}
			continue
		}
		var ev jobs.Event
		var raw []byte
		var status int
		var err error
		job := fmt.Sprintf("/v1/jobs/%d", s.job)
		token := s.token
		if token == "" {
			token = "never-leased"
		}
		switch name {
		case "lease":
			ev = jobs.EvLease
			raw, status, err = c.Do("POST", "/v1/queues/"+s.q+"/lease", map[string]any{"worker_id": worker, "max_jobs": 1})
		case "heartbeat":
			ev = jobs.EvHeartbeat
			raw, status, err = c.Do("POST", job+"/heartbeat", map[string]any{"lease_token": token, "extend_seconds": 60})
		case "ack":
			ev = jobs.EvAck
			raw, status, err = c.Do("POST", job+"/ack", map[string]any{"lease_token": token, "result": map[string]any{"step": step}})
		case "nack", "nack_dead":
			ev = jobs.EvNack
			raw, status, err = c.Do("POST", job+"/nack", map[string]any{"lease_token": token, "error": "walk", "retry": name == "nack"})
		case "cancel":
			ev = jobs.EvCancel
			raw, status, err = c.Do("POST", job+"/cancel", nil)
		case "retry":
			ev = jobs.EvRetry
			raw, status, err = c.Do("POST", job+"/retry", nil)
		case "delete":
			ev = jobs.EvDelete
			raw, status, err = c.Do("DELETE", job, nil)
		case "purge":
			ev = jobs.EvPurge
			raw, status, err = c.Do("POST", "/v1/queues/"+s.q+"/purge", nil)
		}
		if err != nil {
			t.Fatal(err)
		}
		after, _ := read(s)
		targets := jobs.Targets(ev, before)
		accepted := targets != nil
		switch name {
		case "lease":
			var body struct{ Jobs []map[string]any }
			if status != 200 || client.DecodeJSON(raw, &body) != nil {
				t.Fatalf("step %d lease: %d %s", step, status, raw)
			}
			if accepted != (len(body.Jobs) == 1) {
				t.Fatalf("step %d lease of %s job %d returned %s; table accepts=%v", step, before, s.job, raw, accepted)
			}
			if accepted {
				s.token = body.Jobs[0]["lease_token"].(string)
			}
		case "purge":
			if status != 200 {
				t.Fatalf("step %d purge: %d %s", step, status, raw)
			}
		default:
			if ok := status >= 200 && status < 300; ok != accepted || !ok && status != 409 {
				t.Fatalf("step %d %s on %s job %d: %d %s; table accepts=%v", step, name, before, s.job, status, raw, accepted)
			}
		}
		if accepted && !slices.Contains(targets, after) || !accepted && after != before {
			t.Fatalf("step %d %s on job %d: %s→%s; table allows %v", step, name, s.job, before, after, targets)
		}
		if name == "nack_dead" && accepted && after != jobs.Dead {
			t.Fatalf("step %d nack without retry left job %d %s", step, s.job, after)
		}
		if after == jobs.Gone {
			s.job = 0
		}
	}
	for _, s := range slots {
		if s.job != 0 {
			c.Do("POST", fmt.Sprintf("/v1/jobs/%d/cancel", s.job), nil)
			c.Do("DELETE", fmt.Sprintf("/v1/jobs/%d", s.job), nil)
		}
		c.Do("DELETE", "/v1/queues/"+s.q, nil)
	}
	t.Logf("WALK HOLDS: %d steps, seed %d, events %v", steps, seed, counts)
}
