//go:build checks

package checks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"queuelet/internal/client"
	"queuelet/internal/jobs"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Each rejected cell is an individual subtest. Expectations are a fixed public
// contract, not computed from the service's WHERE clauses.
var l07States = []string{"scheduled", "queued", "leased", "retry_wait", "succeeded", "dead", "cancelled"}
var l07Events = []string{"lease_B", "heartbeat_wrong", "ack_wrong", "nack_wrong", "cancel", "retry", "delete"}
var l07Codes = [][]int{
	{200, 409, 409, 409, 200, 409, 409},
	{200, 409, 409, 409, 200, 409, 409},
	{200, 409, 409, 409, 200, 409, 409},
	{200, 409, 409, 409, 200, 409, 409},
	{200, 409, 409, 409, 409, 409, 204},
	{200, 409, 409, 409, 409, 200, 204},
	{200, 409, 409, 409, 409, 409, 204},
}

type l07Fixture struct {
	c              *client.Client
	q              string
	qid, jid, w, b int64
	token          string
}

func l07Call(t *testing.T, c *client.Client, method, path string, body any, want int) map[string]any {
	t.Helper()
	raw, status, err := c.Do(method, path, body)
	if err != nil || status != want {
		t.Fatalf("%s %s status=%d want=%d err=%v body=%s", method, path, status, want, err, raw)
	}
	var value map[string]any
	if len(raw) > 0 {
		if err := client.DecodeJSON(raw, &value); err != nil {
			t.Fatal(err)
		}
	}
	return value
}
func l07Num(v any) int64 { n, _ := v.(json.Number).Int64(); return n }
func l07Make(t *testing.T, state string) l07Fixture {
	t.Helper()
	f := l07Fixture{c: client.Env(), q: fmt.Sprintf("l07-%d", time.Now().UnixNano())}
	v := l07Call(t, f.c, "POST", "/v1/queues", map[string]any{"name": f.q, "lease_seconds": 300, "max_attempts": 3, "backoff": map[string]any{"kind": "linear", "base_seconds": 3600}}, 201)
	f.qid = l07Num(v["id"])
	for _, dst := range []*int64{&f.w, &f.b} {
		v = l07Call(t, f.c, "POST", "/v1/workers", map[string]any{"name": f.q, "kinds": []string{"manual"}}, 201)
		*dst = l07Num(v["worker_id"])
	}
	body := map[string]any{"kind": "manual", "payload": map[string]any{"fixture": true}}
	if state == "scheduled" {
		body["delay_seconds"] = 3600
	}
	v = l07Call(t, f.c, "POST", "/v1/queues/"+f.q+"/jobs", body, 201)
	f.jid = l07Num(v["id"])
	if state != "scheduled" && state != "queued" {
		v = l07Call(t, f.c, "POST", "/v1/queues/"+f.q+"/lease", map[string]any{"worker_id": f.w}, 200)
		batch := v["jobs"].([]any)
		if len(batch) != 1 {
			t.Fatal(v)
		}
		item := batch[0].(map[string]any)
		if l07Num(item["id"]) != f.jid || l07Num(item["attempt"]) != 1 {
			t.Fatal(item)
		}
		f.token = item["lease_token"].(string)
		switch state {
		case "retry_wait":
			f.verb(t, "nack", map[string]any{"lease_token": f.token, "retry": true}, 200)
		case "dead":
			f.verb(t, "nack", map[string]any{"lease_token": f.token, "retry": false}, 200)
		case "succeeded":
			f.verb(t, "ack", map[string]any{"lease_token": f.token, "result": map[string]any{"done": true}}, 200)
		case "cancelled":
			f.verb(t, "cancel", map[string]any{}, 200)
		}
	}
	if f.job(t)["state"] != state {
		t.Fatal("fixture not in requested state", state, f.job(t))
	}
	t.Cleanup(func() {
		// Cleanup avoids leaving leases for later clock probes. Failures remain failures.
		v := f.job(t)
		if s := v["state"]; s != "cancelled" && s != "dead" && s != "succeeded" {
			f.verb(t, "cancel", map[string]any{}, 200)
		}
	})
	return f
}
func (f l07Fixture) path() string { return fmt.Sprintf("/v1/jobs/%d", f.jid) }
func (f l07Fixture) job(t *testing.T) map[string]any {
	t.Helper()
	raw, status, err := f.c.Do("GET", f.path(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if status == 404 {
		return map[string]any{"state": "cancelled"}
	}
	if status != 200 {
		t.Fatalf("job: %d %s", status, raw)
	}
	var v map[string]any
	if err := client.DecodeJSON(raw, &v); err != nil {
		t.Fatal(err)
	}
	return v
}
func (f l07Fixture) history(t *testing.T) any {
	return l07Call(t, f.c, "GET", f.path()+"/attempts", nil, 200)["attempts"]
}
func (f l07Fixture) verb(t *testing.T, verb string, body any, code int) map[string]any {
	return l07Call(t, f.c, "POST", f.path()+"/"+verb, body, code)
}
func l07NoLease(t *testing.T, v map[string]any) {
	t.Helper()
	for _, k := range []string{"lease_token", "leased_by", "lease_expires_at"} {
		if v[k] != nil {
			t.Fatalf("lease retained %s=%v", k, v[k])
		}
	}
}

func TestL07StateMatrix(t *testing.T) {
	rejected, noop, changed := 0, 0, 0
	for si, state := range l07States {
		for ei, event := range l07Events {
			t.Run(state+"/"+event, func(t *testing.T) {
				f := l07Make(t, state)
				before, history := f.job(t), f.history(t)
				want := l07Codes[si][ei]
				var v map[string]any
				switch event {
				case "lease_B":
					v = l07Call(t, f.c, "POST", "/v1/queues/"+f.q+"/lease", map[string]any{"worker_id": f.b}, want)
				case "delete":
					v = l07Call(t, f.c, "DELETE", f.path(), nil, want)
				case "heartbeat_wrong":
					v = f.verb(t, "heartbeat", map[string]any{"lease_token": "wrong", "extend_seconds": 600}, want)
				case "ack_wrong":
					v = f.verb(t, "ack", map[string]any{"lease_token": "wrong", "result": map[string]any{"wrong": true}}, want)
				case "nack_wrong":
					v = f.verb(t, "nack", map[string]any{"lease_token": "wrong", "retry": true, "error": "wrong"}, want)
				default:
					v = f.verb(t, event, map[string]any{}, want)
				}
				if want == 409 || event == "lease_B" && state != "queued" {
					if want == 409 {
						rejected++
						if v["error"].(map[string]any)["code"] != "state_conflict" || v["cancel_requested"] == true {
							t.Fatal(v)
						}
					} else {
						noop++
						if len(v["jobs"].([]any)) != 0 {
							t.Fatal(v)
						}
					}
					if !reflect.DeepEqual(before, f.job(t)) || !reflect.DeepEqual(history, f.history(t)) {
						t.Fatal("rejected/no-op cell changed job or attempts", before, f.job(t), history, f.history(t))
					}
					return
				}
				changed++
				if event == "delete" {
					l07Call(t, f.c, "GET", f.path(), nil, 404)
					return
				}
				after := f.job(t)
				switch event {
				case "lease_B":
					if after["state"] != "leased" || l07Num(after["leased_by"]) != f.b || after["lease_token"] == nil || after["lease_expires_at"] == nil || l07Num(after["attempts"]) != 0 || len(v["jobs"].([]any)) != 1 {
						t.Fatal(after, v)
					}
					if !reflect.DeepEqual(history, f.history(t)) {
						t.Fatal("leasing consumed attempt")
					}
				case "cancel":
					if after["state"] != "cancelled" || after["finished_at"] == nil {
						t.Fatal(after)
					}
					l07NoLease(t, after)
					if state == "leased" {
						if l07Num(after["attempts"]) != 1 || len(f.history(t).([]any)) != 1 {
							t.Fatal(after, f.history(t))
						}
					} else if !reflect.DeepEqual(history, f.history(t)) || before["attempts"] != after["attempts"] {
						t.Fatal("unleased cancel consumed attempt")
					}
				case "retry":
					if after["state"] != "queued" || l07Num(after["attempts"]) != 0 || after["finished_at"] != nil || after["purge_requested"] != false {
						t.Fatal(after)
					}
					l07NoLease(t, after)
					if !reflect.DeepEqual(history, f.history(t)) {
						t.Fatal("retry lost history")
					}
				}
			})
		}
	}
	if rejected != 34 || noop != 6 || changed != 9 {
		t.Fatalf("matrix counts %d/%d/%d", rejected, noop, changed)
	}
	t.Log("49 cells: 34 rejected, 6 empty leases, 9 accepted transitions")
}

func TestL07ExtraActors(t *testing.T) {
	for _, verb := range []string{"heartbeat", "ack", "nack"} {
		t.Run("holder/"+verb, func(t *testing.T) {
			f := l07Make(t, "leased")
			body := map[string]any{"lease_token": f.token, "extend_seconds": 600, "result": map[string]any{"ok": true}, "retry": false}
			before := f.job(t)
			v := f.verb(t, verb, body, 200)
			after := f.job(t)
			if verb == "heartbeat" {
				if v["cancel_requested"] != false || after["state"] != "leased" || after["lease_token"] != f.token || after["lease_expires_at"] == before["lease_expires_at"] || l07Num(after["attempts"]) != 0 || len(f.history(t).([]any)) != 0 {
					t.Fatal(v, after)
				}
			} else {
				want := "succeeded"
				outcome := "succeeded"
				if verb == "nack" {
					want = "dead"
					outcome = "failed"
				}
				a := f.history(t).([]any)
				if after["state"] != want || l07Num(after["attempts"]) != 1 || len(a) != 1 {
					t.Fatal(after, a)
				}
				entry := a[0].(map[string]any)
				if entry["outcome"] != outcome || l07Num(entry["worker_id"]) != f.w {
					t.Fatal(entry)
				}
				l07NoLease(t, after)
			}
		})
		t.Run("expired/"+verb, func(t *testing.T) {
			f := l07Make(t, "leased")
			before, h := f.job(t), f.history(t)
			s := l02Store(t)
			service := jobs.Service{S: s}
			now := time.Now().Add(time.Hour)
			var err error
			switch verb {
			case "heartbeat":
				err = service.Heartbeat(f.jid, f.token, 30, now)
			case "ack":
				err = service.Ack(f.jid, f.token, map[string]any{}, now)
			case "nack":
				err = service.Nack(f.jid, f.token, "late", true, nil, now)
			}
			if err == nil || !reflect.DeepEqual(before, f.job(t)) || !reflect.DeepEqual(h, f.history(t)) {
				t.Fatal("expired mutation accepted or changed effects", err)
			}
		})
	}
	for _, cell := range []string{"cancel-token", "cancel-wrong-token", "cancel-empty-token", "cancel-queued", "cancel-repeat", "cancel-late-ack", "cancel-late-nack"} {
		t.Run(cell, func(t *testing.T) {
			state := "leased"
			if cell == "cancel-queued" {
				state = "queued"
			}
			f := l07Make(t, state)
			f.verb(t, "cancel", map[string]any{}, 200)
			before, h := f.job(t), f.history(t)
			token := f.token
			if cell == "cancel-wrong-token" {
				token = "wrong"
			}
			if cell == "cancel-empty-token" {
				token = ""
			}
			verb := "heartbeat"
			if cell == "cancel-repeat" {
				verb = "cancel"
			}
			if cell == "cancel-late-ack" {
				verb = "ack"
			}
			if cell == "cancel-late-nack" {
				verb = "nack"
			}
			v := f.verb(t, verb, map[string]any{"lease_token": token, "extend_seconds": 3600, "retry": true}, 409)
			if (v["cancel_requested"] == true) != (cell == "cancel-token") {
				t.Fatal("observation authentication", cell, v)
			}
			if !reflect.DeepEqual(before, f.job(t)) || !reflect.DeepEqual(h, f.history(t)) {
				t.Fatal("cancellation observation/completion changed job/history")
			}
			if cell == "cancel-wrong-token" {
				for _, bad := range []string{"\\", "a\x00b", strings.Repeat("x", 32)} {
					response := f.verb(t, "heartbeat", map[string]any{"lease_token": bad, "extend_seconds": 30}, 409)
					if response["cancel_requested"] == true || !reflect.DeepEqual(before, f.job(t)) || !reflect.DeepEqual(h, f.history(t)) {
						t.Fatal("wrong-token observation changed contract", response)
					}
				}
			}
			l07NoLease(t, before)
		})
	}
	t.Log("13 additional cells: 3 holder, 3 expired, 7 cancel-request/late-completion")
}

func TestL07WorkerMissingInsideLease(t *testing.T) {
	f := l07Make(t, "queued")
	s := l02Store(t)
	service := jobs.Service{S: s}
	l07Call(t, f.c, "DELETE", fmt.Sprintf("/v1/workers/%d", f.w), nil, 200)
	before := f.job(t)
	_, err := service.Lease(f.qid, f.w, 1, time.Now())
	if err != jobs.ErrWorkerNotFound || !reflect.DeepEqual(before, f.job(t)) || len(f.history(t).([]any)) != 0 {
		t.Fatal("transactional missing-worker result/effects", err)
	}
}
func TestL07U34(t *testing.T) { l07Absent(t, "U34") }
func TestL07U35(t *testing.T) { l07Absent(t, "U35") }
func TestL07A3(t *testing.T)  { l07Absent(t, "A3") }

func l07Absent(t *testing.T, id string) {
	t.Helper()
	cmd := exec.Command("bash", "lab/repro/"+id+".sh")
	cmd.Dir = "../.."
	var diagnostics bytes.Buffer
	cmd.Stderr = &diagnostics
	out, err := cmd.Output()
	if err != nil || !strings.HasPrefix(string(out), "ABSENT: "+id+" — ") {
		t.Fatalf("%s: %v\n%s\n%s", id, err, out, diagnostics.String())
	}
	t.Log(strings.TrimSpace(string(out)))
	t.Log(diagnostics.String())
}

// A purge waiting on one queue's row must not block another queue's purge.
// Checks may synchronize via database locks; public repros never do.
func TestL07PurgeIndependentQueue(t *testing.T) {
	a, b := l07Make(t, "queued"), l07Make(t, "queued")
	observer := l02Store(t)
	tx, err := observer.DB.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`SELECT id FROM jobs WHERE id=$1 FOR UPDATE`, a.jid); err != nil {
		t.Fatal(err)
	}
	blocked := l02Store(t)
	blocked.DB.SetMaxOpenConns(1)
	app := fmt.Sprintf("l07-purge-%d", a.jid)
	if _, err = blocked.DB.Exec(`SELECT set_config('application_name',$1,false)`, app); err != nil {
		t.Fatal(err)
	}
	type result struct {
		n   int64
		err error
	}
	first := make(chan result, 1)
	go func() { n, err := (&jobs.Service{S: blocked}).Purge(a.qid, time.Now()); first <- result{n, err} }()
	deadline := time.Now().Add(3 * time.Second)
	for {
		var waiting bool
		if err = observer.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE application_name=$1 AND wait_event_type='Lock')`, app).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("purge did not wait on held row")
		}
		time.Sleep(10 * time.Millisecond)
	}
	second := make(chan result, 1)
	go func() { n, err := (&jobs.Service{S: observer}).Purge(b.qid, time.Now()); second <- result{n, err} }()
	select {
	case r := <-second:
		if r.err != nil || r.n != 1 {
			t.Fatal(r)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("unrelated queue serialized behind blocked purge")
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	select {
	case r := <-first:
		if r.err != nil || r.n != 1 {
			t.Fatal(r)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("purge failed to resume")
	}
	l07Call(t, a.c, "GET", a.path(), nil, 404)
	l07Call(t, b.c, "GET", b.path(), nil, 404)
}
