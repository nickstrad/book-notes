// Command gateway is litefleet's stateless tenant-facing tier: it
// authenticates API keys, resolves each database to its node through
// control, and forwards the statement to that node.
package main

import (
	"log"
	"net/http"
	"os"

	"litefleet/internal/api"
	"litefleet/internal/httpx"
	"litefleet/internal/nodeclient"
	"litefleet/internal/routes"
)

func main() {
	instance := env("INSTANCE", "gateway")
	httpx.SetupLogging(instance)
	controlURL := env("CONTROL_URL", "http://control:8090")
	g := &api.Gateway{
		Routes: routes.New(controlURL),
		Nodes:  nodeclient.New(),
	}
	health := newProber(controlURL, env("NODES", "node-a=http://node-a:8070,node-b=http://node-b:8070"))
	go health.run()

	mux := http.NewServeMux()
	g.Register(mux)
	mux.HandleFunc("GET /health", httpx.HealthHandler(map[string]httpx.Check{
		"control": health.controlCheck,
		"nodes":   health.nodesCheck,
	}))
	mux.Handle("GET /metrics", httpx.MetricsHandler())

	served := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Served-By", instance)
		mux.ServeHTTP(w, r)
	})
	addr := env("LISTEN", ":8080")
	log.Printf("%s listening on %s", instance, addr)
	log.Fatal(httpx.NewServer(addr, httpx.Logged(served)).ListenAndServe())
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
