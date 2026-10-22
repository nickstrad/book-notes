package shed

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestShedsOverTheLimit(t *testing.T) {
	release := make(chan struct{})
	var started sync.WaitGroup
	l := New(2, func(*http.Request) bool { return true })
	h := l.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started.Done()
		<-release
	}))
	started.Add(2)
	for i := 0; i < 2; i++ {
		go h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("POST", "/q", nil))
	}
	started.Wait()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/q", nil))
	close(release)
	if rec.Code != http.StatusServiceUnavailable || rec.Header().Get("Retry-After") == "" {
		t.Fatalf("got %d (Retry-After %q); want 503 with Retry-After", rec.Code, rec.Header().Get("Retry-After"))
	}
	if state, _ := l.Check(); state != "degraded" {
		t.Fatalf("health = %s; want degraded while shedding", state)
	}
}

func TestHealthIsNeverShed(t *testing.T) {
	l := New(0, func(*http.Request) bool { return true })
	rec := httptest.NewRecorder()
	l.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(rec, httptest.NewRequest("GET", "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("health got %d", rec.Code)
	}
}
