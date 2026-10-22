// Package shed refuses work an instance cannot finish in time: past a limit
// of requests in flight, it answers 503 with Retry-After in microseconds
// instead of queueing the request behind the ones it is already late for.
//
// Under overload a queue does not make an instance serve more; it makes
// every request wait longer until they all outlive their callers and the
// instance does work nobody receives. Refusing the excess keeps the requests
// it accepts fast. Critical routes may use the whole limit; the rest (the
// usage report, listings) share a quarter of it, so a burst of them cannot
// crowd out queries and writes.
package shed

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"litefleet/internal/httpx"
)

var shedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "shed_total", Help: "Requests refused because the instance was at its in-flight limit.",
}, []string{"route"})

// Limiter caps the requests an instance has in flight.
type Limiter struct {
	limit    int64
	critical func(*http.Request) bool

	all, other atomic.Int64
	lastShed   atomic.Int64 // unix nanoseconds
}

// New returns a limiter of limit requests; critical says which requests may
// use all of it.
func New(limit int, critical func(*http.Request) bool) *Limiter {
	return &Limiter{limit: int64(limit), critical: critical}
}

// Wrap sheds requests over the limit. /health and /metrics are never shed.
func (l *Limiter) Wrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || r.URL.Path == "/metrics" {
			h.ServeHTTP(w, r)
			return
		}
		crit := l.critical(r)
		n := l.all.Add(1)
		defer l.all.Add(-1)
		over := n > l.limit
		if !crit {
			m := l.other.Add(1)
			defer l.other.Add(-1)
			over = over || m > l.limit/4
		}
		if over {
			l.lastShed.Store(time.Now().UnixNano())
			route := "critical"
			if !crit {
				route = "other"
			}
			shedTotal.WithLabelValues(route).Inc()
			w.Header().Set("Retry-After", "1")
			httpx.WriteError(w, r, http.StatusServiceUnavailable, "overloaded", "instance at its in-flight limit; retry shortly")
			return
		}
		h.ServeHTTP(w, r)
	})
}

// Check reports degraded while the limiter has shed in the last 5 s.
func (l *Limiter) Check() (string, string) {
	if time.Since(time.Unix(0, l.lastShed.Load())) < 5*time.Second {
		return httpx.Degraded, "shedding load"
	}
	return httpx.OK, ""
}
