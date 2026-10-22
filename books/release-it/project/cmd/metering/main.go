// Command metering is an honest fake of the usage-metering partner. It
// accepts batches of usage events on POST /events in about 10 ms, de-duplicates
// by event_id, and exposes how many events it received and how many were
// duplicates, so both lost and double-counted usage are visible from the
// partner's side. Slowness is injected by Toxiproxy in front of it.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type event struct {
	EventID    string `json:"event_id"`
	TenantID   string `json:"tenant_id"`
	DatabaseID string `json:"database_id"`
	Kind       string `json:"kind"`
}

var (
	received   = promauto.NewCounter(prometheus.CounterOpts{Name: "events_received_total", Help: "Distinct usage events received."})
	duplicates = promauto.NewCounter(prometheus.CounterOpts{Name: "events_duplicate_total", Help: "Usage events received more than once."})
	batches    = promauto.NewCounter(prometheus.CounterOpts{Name: "event_batches_total", Help: "POST /events requests."})

	mu       sync.Mutex
	seen     = map[string]struct{}{}
	nRecv    int64
	nDup     int64
	byTenant = map[string]int64{}
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /events", post)
	mux.HandleFunc("GET /stats", stats)
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok\n")) })
	log.Printf("metering listening on :8082")
	log.Fatal(http.ListenAndServe(":8082", mux))
}

func post(w http.ResponseWriter, r *http.Request) {
	batches.Inc()
	var evs []event
	if err := json.NewDecoder(r.Body).Decode(&evs); err != nil {
		http.Error(w, `{"error":{"code":"bad_request","message":"bad batch"}}`, http.StatusBadRequest)
		return
	}
	time.Sleep(10 * time.Millisecond)
	mu.Lock()
	for _, e := range evs {
		if _, ok := seen[e.EventID]; ok {
			nDup++
			duplicates.Inc()
			continue
		}
		seen[e.EventID] = struct{}{}
		nRecv++
		byTenant[e.TenantID]++
		received.Inc()
	}
	mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"accepted":true}` + "\n"))
}

func stats(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"received": nRecv, "duplicates": nDup, "tenants": len(byTenant)})
}
