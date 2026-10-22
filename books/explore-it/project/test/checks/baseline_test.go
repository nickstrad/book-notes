//go:build checks

package checks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"queuelet/internal/client"
	"testing"
	"time"
)

func TestL00QueueCRUD(t *testing.T) {
	c := client.Env()
	name := fmt.Sprintf("check-crud-%d", time.Now().UnixNano())
	var q map[string]any
	if e := c.JSON("POST", "/v1/queues", map[string]any{"name": name}, &q); e != nil {
		t.Fatal(e)
	}
	if e := c.JSON("GET", "/v1/queues/"+name, nil, &q); e != nil {
		t.Fatal(e)
	}
	_, status, e := c.Do("DELETE", "/v1/queues/"+name, nil)
	if e != nil || status != 204 {
		t.Fatalf("delete: %d %v", status, e)
	}
}
func TestL00HappyPath(t *testing.T) {
	c := client.Env()
	name := fmt.Sprintf("check-flow-%d", time.Now().UnixNano())
	var q, j, w map[string]any
	must := func(e error) {
		t.Helper()
		if e != nil {
			t.Fatal(e)
		}
	}
	must(c.JSON("POST", "/v1/queues", map[string]any{"name": name, "webhook": map[string]any{"url": "http://hooksink:8201/hook", "secret": "report-secret"}}, &q))
	must(c.JSON("POST", "/v1/queues/"+name+"/jobs", map[string]any{"kind": "echo", "payload": map[string]any{"hello": "world"}}, &j))
	must(c.JSON("POST", "/v1/workers", map[string]any{"name": "check-worker", "kinds": []string{"echo"}}, &w))
	var lease struct {
		Jobs []struct {
			ID    int64  `json:"id"`
			Token string `json:"lease_token"`
		}
	}
	must(c.JSON("POST", "/v1/queues/"+name+"/lease", map[string]any{"worker_id": w["worker_id"], "max_jobs": 1}, &lease))
	if len(lease.Jobs) != 1 {
		t.Fatal("no job leased")
	}
	job := lease.Jobs[0]
	var out any
	must(c.JSON("POST", fmt.Sprintf("/v1/jobs/%d/ack", job.ID), map[string]any{"lease_token": job.Token, "result": map[string]any{"ok": true}}, &out))
	must(c.JSON("GET", fmt.Sprintf("/v1/jobs/%d", job.ID), nil, &j))
	if j["state"] != "succeeded" {
		t.Fatal(j)
	}
	hook := os.Getenv("QUEUELET_HOOK_URL")
	if hook == "" {
		hook = "http://localhost:18201"
	}
	for i := 0; i < 50; i++ {
		resp, e := http.Get(hook + "/admin/records")
		must(e)
		var records struct {
			Records []struct {
				Body struct {
					ID int64 `json:"job_id"`
				} `json:"body"`
			} `json:"records"`
		}
		json.NewDecoder(resp.Body).Decode(&records)
		resp.Body.Close()
		for _, record := range records.Records {
			if record.Body.ID == job.ID {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("webhook not delivered")
}
func TestL00Auth(t *testing.T) {
	c := client.Env()
	c.Key = "invalid"
	_, status, e := c.Do("GET", "/v1/queues", nil)
	if e != nil || status != 401 {
		t.Fatalf("auth status %d error %v", status, e)
	}
}

func TestL00ConcurrentLeases(t *testing.T) {
	c := client.Env()
	name := fmt.Sprintf("check-leases-%d", time.Now().UnixNano())
	var out map[string]any
	if e := c.JSON("POST", "/v1/queues", map[string]any{"name": name}, &out); e != nil {
		t.Fatal(e)
	}
	var highID float64
	for _, priority := range []int{1, 9, 3} {
		if e := c.JSON("POST", "/v1/queues/"+name+"/jobs", map[string]any{"kind": "echo", "payload": map[string]any{}, "priority": priority}, &out); e != nil {
			t.Fatal(e)
		}
		if priority == 9 {
			highID = out["id"].(float64)
		}
	}
	workers := []any{}
	for i := 0; i < 2; i++ {
		if e := c.JSON("POST", "/v1/workers", map[string]any{"name": name, "kinds": []string{"echo"}}, &out); e != nil {
			t.Fatal(e)
		}
		workers = append(workers, out["worker_id"])
	}
	type leased struct {
		Jobs []struct {
			ID int64 `json:"id"`
		} `json:"jobs"`
	}
	var first leased
	if e := c.JSON("POST", "/v1/queues/"+name+"/lease", map[string]any{"worker_id": workers[0], "max_jobs": 1}, &first); e != nil {
		t.Fatal(e)
	}
	if len(first.Jobs) != 1 || first.Jobs[0].ID != int64(highID) {
		t.Fatalf("highest priority was not leased first: %+v", first)
	}
	type answer struct {
		batch leased
		err   error
	}
	ch := make(chan answer, 2)
	for _, id := range workers {
		go func(id any) {
			var b leased
			e := c.JSON("POST", "/v1/queues/"+name+"/lease", map[string]any{"worker_id": id, "max_jobs": 1}, &b)
			ch <- answer{b, e}
		}(id)
	}
	a, b := <-ch, <-ch
	if a.err != nil || b.err != nil {
		t.Fatalf("lease errors: %v %v", a.err, b.err)
	}
	if len(a.batch.Jobs) != 1 || len(b.batch.Jobs) != 1 {
		t.Fatalf("missing leases: %+v %+v", a.batch, b.batch)
	}
	if a.batch.Jobs[0].ID == b.batch.Jobs[0].ID {
		t.Fatal("concurrent workers acquired the same job")
	}
}
