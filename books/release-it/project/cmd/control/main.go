// Command control is litefleet's control plane: it owns the Postgres
// registry, authenticates API keys, resolves databases to nodes, places new
// databases, receives node heartbeats and serves the fleet usage report.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"litefleet/internal/httpx"
	"litefleet/internal/placement"
	"litefleet/internal/registry"
	"litefleet/internal/wire"
	"litefleet/migrations"
)

type control struct {
	reg   *registry.Registry
	nodes *nodeCommands
}

func main() {
	httpx.SetupLogging("control")
	reg, err := registry.Open(env("DATABASE_URL", "postgres://litefleet:litefleet@postgres:5432/litefleet"))
	if err != nil {
		log.Fatal(err)
	}
	// Bring the schema up to date before serving.
	for {
		err := reg.Migrate(migrations.FS)
		if err == nil {
			break
		}
		log.Printf("migrate: %v; retrying", err)
		time.Sleep(time.Second)
	}

	c := &control{reg: reg, nodes: newNodeCommands(env("NODES", "node-a=http://node-a:8070,node-b=http://node-b:8070"))}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /internal/resolve", budget(450*time.Millisecond, c.resolve))
	mux.HandleFunc("POST /internal/databases", budget(2*time.Second, c.createDatabase))
	mux.HandleFunc("GET /internal/databases", budget(450*time.Millisecond, c.listDatabases))
	mux.HandleFunc("GET /internal/databases/{name}", budget(450*time.Millisecond, c.getDatabase))
	mux.HandleFunc("DELETE /internal/databases/{name}", budget(2*time.Second, c.dropDatabase))
	mux.HandleFunc("POST /internal/heartbeat", budget(2*time.Second, c.heartbeat))
	mux.HandleFunc("GET /internal/usage", budget(8*time.Second, c.usage))
	mux.HandleFunc("GET /health", httpx.HealthHandler(map[string]httpx.Check{"postgres": c.postgresCheck}))
	mux.Handle("GET /metrics", httpx.MetricsHandler())

	addr := env("LISTEN", ":8090")
	log.Printf("control listening on %s", addr)
	log.Fatal(httpx.NewServer(addr, httpx.Logged(mux)).ListenAndServe())
}

// budget gives a handler's registry calls a deadline inside its caller's, so
// a request waiting for a pool connection fails fast instead of hanging.
func budget(d time.Duration, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), d)
		defer cancel()
		h(w, r.WithContext(ctx))
	}
}

// postgresCheck: control cannot answer anything without its registry.
func (c *control) postgresCheck() (string, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := c.reg.Ping(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
			return httpx.Failed, "postgres timeout"
		}
		return httpx.Failed, "postgres: " + err.Error()
	}
	return httpx.OK, ""
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

var dbName = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)

func (c *control) resolve(w http.ResponseWriter, r *http.Request) {
	var req wire.ResolveRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	rt, err := c.reg.Resolve(r.Context(), req.APIKey, req.DBName)
	if err != nil {
		writeRegistryError(w, r, err)
		return
	}
	httpx.Annotate(r.Context(), "tenant_id", rt.TenantID)
	httpx.Annotate(r.Context(), "database_id", rt.DatabaseID)
	httpx.WriteJSON(w, http.StatusOK, rt)
}

// tenant authenticates the bearer key forwarded by the gateway.
func (c *control) tenant(w http.ResponseWriter, r *http.Request) (string, bool) {
	t, _, err := c.reg.Tenant(r.Context(), httpx.Bearer(r))
	if err != nil {
		writeRegistryError(w, r, err)
		return "", false
	}
	return t, true
}

func (c *control) createDatabase(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := c.tenant(w, r)
	if !ok {
		return
	}
	var req wire.CreateDatabaseRequest
	if err := httpx.DecodeJSON(r, &req); err != nil || !dbName.MatchString(req.Name) {
		httpx.WriteError(w, r, http.StatusBadRequest, "bad_request", "name must match "+dbName.String())
		return
	}
	id, err := c.reg.BeginCreate(r.Context(), tenantID, req.Name)
	if err != nil {
		writeRegistryError(w, r, err)
		return
	}
	load, err := c.reg.NodeLoad(r.Context())
	if err != nil {
		c.reg.AbortCreate(r.Context(), id)
		writeRegistryError(w, r, err)
		return
	}
	node, err := placement.Pick(load)
	if err != nil {
		c.reg.AbortCreate(r.Context(), id)
		httpx.WriteError(w, r, http.StatusServiceUnavailable, "no_capacity", err.Error())
		return
	}
	const epoch = 1
	if err := c.nodes.provision(node, wire.ProvisionRequest{DatabaseID: id, TenantID: tenantID, Epoch: epoch}); err != nil {
		c.reg.AbortCreate(r.Context(), id)
		httpx.WriteError(w, r, http.StatusBadGateway, "provision_failed", err.Error())
		return
	}
	if err := c.reg.FinishCreate(r.Context(), id, node, epoch); err != nil {
		writeRegistryError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, wire.Database{ID: id, Name: req.Name, Node: node})
}

func (c *control) listDatabases(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := c.tenant(w, r)
	if !ok {
		return
	}
	dbs, err := c.reg.ListDatabases(r.Context(), tenantID)
	if err != nil {
		writeRegistryError(w, r, err)
		return
	}
	if dbs == nil {
		dbs = []wire.Database{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"databases": dbs})
}

func (c *control) getDatabase(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := c.tenant(w, r)
	if !ok {
		return
	}
	d, err := c.reg.Database(r.Context(), tenantID, r.PathValue("name"))
	if err != nil {
		writeRegistryError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, d)
}

func (c *control) dropDatabase(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := c.tenant(w, r)
	if !ok {
		return
	}
	d, err := c.reg.Database(r.Context(), tenantID, r.PathValue("name"))
	if err != nil {
		writeRegistryError(w, r, err)
		return
	}
	if err := c.nodes.drop(d.Node, d.ID); err != nil {
		httpx.WriteError(w, r, http.StatusBadGateway, "drop_failed", err.Error())
		return
	}
	if err := c.reg.MarkDropped(r.Context(), d.ID); err != nil {
		writeRegistryError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *control) heartbeat(w http.ResponseWriter, r *http.Request) {
	var hb wire.Heartbeat
	if err := httpx.DecodeJSON(r, &hb); err != nil || hb.Node == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "bad_request", "bad heartbeat")
		return
	}
	if err := c.reg.Heartbeat(r.Context(), hb); err != nil {
		log.Printf("heartbeat from %s: %v", hb.Node, err)
		httpx.WriteError(w, r, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeRegistryError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, registry.ErrUnauthorized):
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", err.Error())
	case errors.Is(err, registry.ErrNoSuchDatabase):
		httpx.WriteError(w, r, http.StatusNotFound, "no_such_database", err.Error())
	case errors.Is(err, registry.ErrExists):
		httpx.WriteError(w, r, http.StatusConflict, "exists", err.Error())
	case errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled):
		httpx.WriteError(w, r, http.StatusServiceUnavailable, "pool_timeout", "registry busy")
	default:
		log.Printf("registry: %v", err)
		httpx.WriteError(w, r, http.StatusInternalServerError, "internal", strings.TrimSpace(err.Error()))
	}
}
