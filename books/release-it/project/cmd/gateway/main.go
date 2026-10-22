// Command gateway is litefleet's stateless tenant-facing tier: it
// authenticates API keys, resolves each database to its node through
// control, and forwards the statement to that node.
package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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
	}))
	mux.Handle("GET /metrics", httpx.MetricsHandler())

	served := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Served-By", instance)
		mux.ServeHTTP(w, r)
	})
	addr := env("LISTEN", ":8080")
	log.Printf("%s listening on %s", instance, addr)
	log.Fatal(httpx.NewServer(addr, httpx.Logged(limiter.Wrap(served))).ListenAndServe())
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
