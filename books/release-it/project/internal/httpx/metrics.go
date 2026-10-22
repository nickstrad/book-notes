package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics never carry a tenant or database label: one series per tenant
// would be thousands of series per instance. Per-tenant questions go to the
// usage ledger and the logs.
var (
	requests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total", Help: "Requests served, by route pattern, method and status.",
	}, []string{"route", "method", "status"})
	duration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds", Help: "Time to serve a request, by route pattern.",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 14),
	}, []string{"route"})
	inflight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "http_inflight_requests", Help: "Requests being served right now.",
	})
	depRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "dependency_requests_total", Help: "Calls to a dependency, by outcome.",
	}, []string{"dep", "outcome"})
	depDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "dependency_duration_seconds", Help: "Time a dependency took to answer.",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 14),
	}, []string{"dep"})
)

// MetricsHandler serves /metrics (Go runtime and process collectors included:
// process_open_fds, go_goroutines).
func MetricsHandler() http.Handler { return promhttp.Handler() }

// Outcome classifies a dependency call for metrics and logs.
func Outcome(status int, err error) string {
	var ne net.Error
	switch {
	case err == nil && status < 400:
		return "ok"
	case errors.Is(err, context.DeadlineExceeded), errors.As(err, &ne) && ne.Timeout():
		return "timeout"
	case err != nil && status == 0:
		return "error"
	case status >= 500:
		return "http_5xx"
	default:
		return "http_4xx"
	}
}

// Dependency records one call to dep: a metric sample and a log line with the
// request id, so a slow request can be attributed to the dependency that made
// it slow.
func Dependency(ctx context.Context, dep string, start time.Time, status int, err error) {
	d := time.Since(start)
	outcome := DependencyMetric(dep, start, status, err)
	attrs := []any{"dep", dep, "dur_ms", ms(d), "outcome", outcome}
	if id := RequestIDFrom(ctx); id != "" {
		attrs = append([]any{"request_id", id}, attrs...)
	}
	if status != 0 {
		attrs = append(attrs, "status", status)
	}
	if err != nil {
		attrs = append(attrs, "error", err.Error())
	}
	slog.Info("dependency", attrs...)
}

// DependencyMetric records a dependency call in metrics only (for calls too
// frequent to log one line each, such as registry queries) and returns its
// outcome; failed and slow calls are still logged.
func DependencyMetric(dep string, start time.Time, status int, err error) string {
	d := time.Since(start)
	outcome := Outcome(status, err)
	depRequests.WithLabelValues(dep, outcome).Inc()
	depDuration.WithLabelValues(dep).Observe(d.Seconds())
	if dep == "postgres" && (outcome != "ok" || d > 100*time.Millisecond) {
		attrs := []any{"dep", dep, "dur_ms", ms(d), "outcome", outcome}
		if err != nil {
			attrs = append(attrs, "error", err.Error())
		}
		slog.Warn("slow or failed dependency call", attrs...)
	}
	return outcome
}

// Propagate copies the request id from ctx onto an outgoing request.
func Propagate(ctx context.Context, req *http.Request) {
	if id := RequestIDFrom(ctx); id != "" {
		req.Header.Set("X-Request-Id", id)
	}
}
