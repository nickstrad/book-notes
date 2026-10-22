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
	g := &api.Gateway{
		Routes: routes.New(env("CONTROL_URL", "http://control:8090")),
		Nodes:  nodeclient.New(),
	}
	mux := http.NewServeMux()
	g.Register(mux)
	mux.HandleFunc("GET /health", httpx.Health)

	served := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Served-By", instance)
		httpx.RequestID(r)
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
