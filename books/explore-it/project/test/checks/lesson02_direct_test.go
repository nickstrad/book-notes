//go:build checks

package checks

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"queuelet/internal/api"
	"queuelet/internal/store"
	"queuelet/internal/webhook"
	"testing"
	"time"
)

func l02Store(t *testing.T) *store.Store {
	t.Helper()
	url := os.Getenv("QUEUELET_DATABASE_URL")
	if url == "" {
		url = "postgres://queuelet:queuelet@localhost:15433/queuelet?sslmode=disable"
	}
	s, e := store.Open(url)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { s.DB.Close() })
	return s
}
func TestL02ProductionClockDisabled(t *testing.T) {
	s := l02Store(t)
	t.Setenv("QUEUELET_TEST_CLOCK", "0")
	r := httptest.NewRequest("POST", "/admin/clock", nil)
	r.Header.Set("Authorization", "Bearer admin-key")
	w := httptest.NewRecorder()
	api.New(s).Handler().ServeHTTP(w, r)
	if w.Code != 404 {
		t.Fatalf("production clock status=%d body=%s", w.Code, w.Body.String())
	}
}
func TestL02LegacyMalformedOutboxIsolation(t *testing.T) {
	// Direct persisted-legacy fixture, deliberately separate from public repros.
	// Private schema avoids racing the running API shipper or mutating shared rows.
	s := l02Store(t)
	s.DB.SetMaxOpenConns(1)
	schema := fmt.Sprintf("check_e02_%d", time.Now().UnixNano())
	if _, e := s.DB.Exec("CREATE SCHEMA " + schema); e != nil {
		t.Fatal(e)
	}
	defer s.DB.Exec("DROP SCHEMA " + schema + " CASCADE")
	if _, e := s.DB.Exec("SET search_path TO " + schema); e != nil {
		t.Fatal(e)
	}
	_, e := s.DB.Exec(`CREATE TABLE webhook_outbox(id bigint,job_id bigint,url text,secret text,body jsonb,attempts int DEFAULT 0,next_at timestamptz,delivered_at timestamptz,event_at timestamptz,status int,error text)`)
	if e != nil {
		t.Fatal(e)
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(204) }))
	defer server.Close()
	now := time.Now().UTC().Truncate(time.Microsecond)
	_, e = s.DB.Exec(`INSERT INTO webhook_outbox(id,job_id,url,secret,body,next_at) VALUES(1,1,'http://bad host/x','s','{}',$1),(2,2,$2,'s','{}',$1)`, now, server.URL)
	if e != nil {
		t.Fatal(e)
	}
	if e = webhook.Ship(s, now); e != nil {
		t.Fatal(e)
	}
	var attempts int
	var message string
	var next time.Time
	if e = s.DB.QueryRow(`SELECT attempts,error,next_at FROM webhook_outbox WHERE id=1`).Scan(&attempts, &message, &next); e != nil {
		t.Fatal(e)
	}
	if attempts != 1 || message == "" || !next.Equal(now.Add(5*time.Second)) {
		t.Fatalf("legacy failure attempts=%d error=%q next=%s", attempts, message, next)
	}
	var delivered bool
	if e = s.DB.QueryRow(`SELECT delivered_at IS NOT NULL FROM webhook_outbox WHERE id=2`).Scan(&delivered); e != nil {
		t.Fatal(e)
	}
	if !delivered || calls != 1 {
		t.Fatalf("healthy row blocked: delivered=%v calls=%d", delivered, calls)
	}
}
