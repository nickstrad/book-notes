//go:build chaos

// Package chaos is the test harness for litefleet's integration points: it
// drives the real clients against the running stack while Toxiproxy (network
// faults) or the blob-store fake (application faults) misbehave, and checks
// that every call fails the way it should, inside its budget.
//
// Run with `make test-chaos` (inside the Compose network). The tests use only
// the clients' public calls, so the same tests run against an older tree.
package chaos

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"litefleet/internal/blob"
	"litefleet/internal/breaker"
	"litefleet/internal/httpx"
	"litefleet/internal/nodeclient"
	"litefleet/internal/routes"
	"litefleet/internal/wire"
)

const (
	toxiproxy  = "http://toxiproxy:8474"
	nodeAProxy = "http://toxiproxy:9071"
	blobProxy  = "http://toxiproxy:9081"
	blobAdmin  = "http://blobstore:8081"
	ctrlProxy  = "http://toxiproxy:9090"
)

// kindOf is the failure class the caller can act on: the client's own
// classification when it has one, else what the standard error types say.
func kindOf(err error) string {
	if err == nil {
		return "ok"
	}
	if errors.Is(err, breaker.ErrOpen) {
		return "breaker_open"
	}
	var k interface{ Kind() string }
	if errors.As(err, &k) {
		return k.Kind()
	}
	var re *httpx.RemoteError
	var ne net.Error
	switch {
	case errors.Is(err, context.DeadlineExceeded), errors.As(err, &ne) && ne.Timeout():
		return "timeout"
	case errors.As(err, &re) && re.Code == "stale_epoch":
		return "stale_epoch"
	case errors.As(err, &re) && re.Status >= 500:
		return "remote_5xx"
	case errors.As(err, &re):
		return "remote_4xx"
	}
	return "unclassified"
}

func api(t *testing.T, method, url string, body any) {
	t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, url, rd)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		t.Fatalf("%s %s: %s", method, url, resp.Status)
	}
}

// toxic adds a toxic for the rest of the test. It is removed by cycling the
// proxy (Toxiproxy can wedge removing a toxic from a proxy with open links).
func toxic(t *testing.T, proxy, typ, stream string, attrs map[string]any) {
	t.Helper()
	api(t, "POST", toxiproxy+"/proxies/"+proxy+"/toxics", map[string]any{"name": "chaos", "type": typ, "stream": stream, "attributes": attrs})
	t.Cleanup(func() {
		api(t, "POST", toxiproxy+"/proxies/"+proxy, map[string]any{"enabled": false})
		api(t, "DELETE", toxiproxy+"/proxies/"+proxy+"/toxics/chaos", nil)
		api(t, "POST", toxiproxy+"/proxies/"+proxy, map[string]any{"enabled": true})
	})
}

func proxyDown(t *testing.T, proxy string) {
	t.Helper()
	api(t, "POST", toxiproxy+"/proxies/"+proxy, map[string]any{"enabled": false})
	t.Cleanup(func() { api(t, "POST", toxiproxy+"/proxies/"+proxy, map[string]any{"enabled": true}) })
}

type manifest struct {
	Tenants []struct {
		Key   string `json:"key"`
		Class string `json:"class"`
		DBs   []struct {
			Name string `json:"name"`
			ID   string `json:"id"`
			Node string `json:"node"`
		} `json:"dbs"`
	} `json:"tenants"`
}

// nodeA returns a normal tenant's database on node-a, its route through the
// node-a proxy, and the tenant's key.
func nodeA(t *testing.T) (wire.Route, string, string) {
	t.Helper()
	data, err := os.ReadFile("/results/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var m manifest
	json.Unmarshal(data, &m)
	for _, tn := range m.Tenants {
		for _, d := range tn.DBs {
			if tn.Class == "normal" && d.Node == "node-a" && d.Name == "default" {
				return wire.Route{DatabaseID: d.ID, NodeURL: nodeAProxy, Node: "node-a", Epoch: 1}, tn.Key, d.Name
			}
		}
	}
	t.Fatal("no node-a database in the manifest")
	return wire.Route{}, "", ""
}

// matrix logs one row of the harness matrix: client, misbehaviour, outcome, elapsed.
func matrix(t *testing.T, client, misbehaviour, outcome string, start time.Time) time.Duration {
	d := time.Since(start)
	t.Logf("MATRIX %-10s %-16s %-14s %6.0f ms", client, misbehaviour, outcome, float64(d.Microseconds())/1000)
	return d
}

// query calls the node through the gateway's client with the gateway's
// request deadline and reads the whole answer.
func query(rt wire.Route, sql string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2900*time.Millisecond)
	defer cancel()
	rows, err := nodeclient.New().Query(ctx, rt, wire.QueryRequest{SQL: sql})
	if err != nil {
		return err
	}
	_, err = rows.WriteTo(io.Discard)
	return err
}

func nodeCase(t *testing.T, name, sql, wantKind string, budget time.Duration, inject func()) {
	rt, _, _ := nodeA(t)
	inject()
	start := time.Now()
	err := query(rt, sql)
	got := kindOf(err)
	d := matrix(t, "node", name, got, start)
	if got != wantKind {
		t.Errorf("%s: outcome %q (%v); want %q", name, got, err, wantKind)
	}
	if d > budget {
		t.Errorf("%s: took %s; budget %s", name, d, budget)
	}
}

const small = "SELECT id, seq FROM items WHERE id <= 5"
const big = "SELECT * FROM items" // about 80 KB

func TestNodeLatency(t *testing.T) {
	nodeCase(t, "latency 3s", small, "timeout", 3*time.Second, func() {
		toxic(t, "node-a", "latency", "downstream", map[string]any{"latency": 3000})
	})
}

func TestNodeTimeout(t *testing.T) {
	nodeCase(t, "accept, no answer", small, "timeout", 3*time.Second, func() {
		toxic(t, "node-a", "timeout", "upstream", map[string]any{"timeout": 0})
	})
}

func TestNodeSlicer(t *testing.T) {
	nodeCase(t, "slicer", big, "ok", 3*time.Second, func() {
		toxic(t, "node-a", "slicer", "downstream", map[string]any{"average_size": 64, "size_variation": 32, "delay": 100})
	})
}

func TestNodeBandwidth(t *testing.T) {
	nodeCase(t, "bandwidth 1KB/s", big, "timeout", 3*time.Second, func() {
		toxic(t, "node-a", "bandwidth", "downstream", map[string]any{"rate": 1})
	})
}

func TestNodeResetPeer(t *testing.T) {
	nodeCase(t, "reset_peer", small, "unreachable", time.Second, func() {
		toxic(t, "node-a", "reset_peer", "upstream", map[string]any{"timeout": 0})
	})
}

func TestNodeDown(t *testing.T) {
	nodeCase(t, "down", small, "unreachable", time.Second, func() { proxyDown(t, "node-a") })
}

func TestNodeStaleEpoch(t *testing.T) {
	rt, _, _ := nodeA(t)
	rt.Epoch = 99
	start := time.Now()
	err := query(rt, small)
	got := kindOf(err)
	matrix(t, "node", "stale epoch", got, start)
	if got != "stale_epoch" {
		t.Errorf("outcome %q (%v); want stale_epoch", got, err)
	}
}

func blobCase(t *testing.T, mode, wantKind string, budget time.Duration) {
	api(t, "POST", blobAdmin+"/admin/mode", map[string]string{"mode": mode})
	t.Cleanup(func() { api(t, "POST", blobAdmin+"/admin/mode", map[string]string{"mode": "normal"}) })
	path := filepath.Join(t.TempDir(), "snap.db")
	os.WriteFile(path, bytes.Repeat([]byte("x"), 64<<10), 0o644)
	start := time.Now()
	_, err := blob.New(blobProxy).Put("chaos-test", 1, path)
	got := kindOf(err)
	d := matrix(t, "blobstore", mode, got, start)
	if got != wantKind {
		t.Errorf("%s: outcome %q (%v); want %q", mode, got, err, wantKind)
	}
	if d > budget {
		t.Errorf("%s: took %s; budget %s", mode, d, budget)
	}
}

func TestBlob500WithHTML(t *testing.T)   { blobCase(t, "500", "remote_5xx", time.Second) }
func TestBlobMalformedJSON(t *testing.T) { blobCase(t, "malformed-json", "bad_payload", time.Second) }
func TestBlobHugeErrorBody(t *testing.T) { blobCase(t, "huge-body", "remote_5xx", time.Second) }
func TestBlobSlowHeaders(t *testing.T)   { blobCase(t, "slow-headers", "timeout", 1500*time.Millisecond) }

func TestControlBreakerOpensAndRecovers(t *testing.T) {
	_, key, db := nodeA(t)
	c := routes.New(ctrlProxy)
	proxyDown(t, "control")
	var err error
	for i := 0; i < 10 && !errors.Is(err, breaker.ErrOpen); i++ {
		_, err = c.Resolve(context.Background(), key, fmt.Sprintf("%s-%d", db, i))
	}
	start := time.Now()
	_, err = c.Resolve(context.Background(), key, db)
	matrix(t, "control", "down", kindOf(err), start)
	if !errors.Is(err, breaker.ErrOpen) || time.Since(start) > 10*time.Millisecond {
		t.Fatalf("breaker did not fail fast while control was down: %v after %s", err, time.Since(start))
	}
	api(t, "POST", toxiproxy+"/proxies/control", map[string]any{"enabled": true})
	time.Sleep(5500 * time.Millisecond)
	start = time.Now()
	_, err = c.Resolve(context.Background(), key, db)
	matrix(t, "control", "back up", kindOf(err), start)
	if err != nil || c.BreakerState() != breaker.Closed {
		t.Fatalf("after recovery: %v, breaker %v; want a resolve and a closed breaker", err, c.BreakerState())
	}
}
