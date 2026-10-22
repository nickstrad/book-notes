package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"litefleet/internal/engine"
	"litefleet/internal/httpx"
	"litefleet/internal/shed"
)

// A node restart is a deploy step, not a crash. On SIGTERM the node drains:
// it stops admitting statements (503 draining with Retry-After, and a failed
// health check), lets the ones running finish within their deadline,
// checkpoints and closes every database, keeps its queued usage events, and
// only then exits. Nothing in flight is cut off.
func (n *node) drainOnSignal(srv *http.Server, limiter *shed.Limiter) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig
	start := time.Now()
	n.draining.Store(true)
	deadline := start.Add(engine.StatementDeadline + time.Second)
	for limiter.InFlight() > 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	left := limiter.InFlight()
	n.eng.Close() // checkpoints every WAL, closes every handle
	n.usage.Spill(3 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	srv.Shutdown(ctx)
	cancel()
	log.Printf("drained in %s (%d requests still running at the deadline)", time.Since(start).Round(time.Millisecond), left)
	os.Exit(0)
}

// admitting answers 503 draining once the node has started to drain.
func (n *node) admitting(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n.draining.Load() && r.URL.Path != "/health" && r.URL.Path != "/metrics" {
			w.Header().Set("Retry-After", "2")
			w.Header().Set("Connection", "close") // no kept-alive connection outlives the process
			httpx.WriteError(w, r, http.StatusServiceUnavailable, "draining", "node is restarting; retry shortly")
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (n *node) drainCheck() (string, string) {
	if n.draining.Load() {
		return httpx.Failed, "draining"
	}
	return httpx.OK, ""
}
