// Command gateway is litefleet's stateless tenant-facing tier: it
// authenticates API keys, resolves each database to its node through
// control, and forwards the statement to that node.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"litefleet/internal/api"
	"litefleet/internal/httpx"
	"litefleet/internal/nodeclient"
	"litefleet/internal/routes"
	"litefleet/internal/shed"
)

func main() {
	instance := env("INSTANCE", "gateway")
	httpx.SetupLogging(instance)
	controlURL := env("CONTROL_URL", "http://control:8090")
	g := &api.Gateway{
		Routes: routes.New(controlURL),
		Nodes:  nodeclient.New(),
	}
	if ttl, err := time.ParseDuration(env("ROUTE_CACHE_TTL", "")); err == nil {
		g.Routes.SetCacheTTL(ttl) // a config override for experiments (lesson 12's Godot run)
	}
	health := newProber(controlURL, env("NODES", "node-a=http://node-a:8070,node-b=http://node-b:8070"), g.Routes)
	go health.run()
	// About 100 ms of work at this instance's calibrated capacity; past it
	// a request would wait longer than it is worth.
	limit, _ := strconv.Atoi(env("SHED_LIMIT", "24"))
	limiter := shed.New(limit, func(r *http.Request) bool {
		return strings.HasSuffix(r.URL.Path, "/query") || strings.HasSuffix(r.URL.Path, "/execute") ||
			(r.Method == http.MethodPost && r.URL.Path == "/v1/databases")
	})

	mux := http.NewServeMux()
	g.Register(mux)
	mux.HandleFunc("GET /health", httpx.HealthHandler(map[string]httpx.Check{
		"control": health.controlCheck,
		"nodes":   health.nodesCheck,
		"load":    limiter.Check,
		"drain":   health.drainCheck,
	}))
	mux.Handle("GET /metrics", httpx.MetricsHandler())

	served := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Served-By", instance)
		mux.ServeHTTP(w, r)
	})
	addr := env("LISTEN", ":8080")
	srv := httpx.NewServer(addr, httpx.Logged(limiter.Wrap(served)))
	go drainOnSignal(srv, health)
	log.Printf("%s listening on %s", instance, addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
	select {} // drainOnSignal exits the process
}

// drainOnSignal takes the gateway out of rotation before it stops: on
// SIGTERM its health check fails, HAProxy stops sending it requests within
// two checks, and only then does it stop listening, after the requests it
// has finish.
func drainOnSignal(srv *http.Server, p *prober) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig
	p.draining.Store(true)
	time.Sleep(2500 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Printf("drained")
	os.Exit(0)
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
