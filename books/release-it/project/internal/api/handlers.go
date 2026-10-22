// Package api is the gateway's tenant-facing HTTP API.
package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"litefleet/internal/breaker"
	"litefleet/internal/httpx"
	"litefleet/internal/nodeclient"
	"litefleet/internal/routes"
	"litefleet/internal/wire"
)

// MaxSQLBytes is the largest statement the platform accepts.
const MaxSQLBytes = 64 << 10

// RequestDeadline bounds a tenant's query or execute end to end. It sits
// just inside the 3 s our clients wait, so a timeout reaches them as a 504
// from the gateway rather than as silence.
const RequestDeadline = 2900 * time.Millisecond

// withDeadline gives the request the platform's deadline; every call it
// makes (control, node) inherits it.
func withDeadline(r *http.Request) (*http.Request, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestDeadline)
	return r.WithContext(ctx), cancel
}

// Gateway serves the tenant API.
type Gateway struct {
	Routes *routes.Client
	Nodes  *nodeclient.Client
}

// Register installs the API routes on mux.
func (g *Gateway) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/databases", g.create)
	mux.HandleFunc("GET /v1/databases", g.list)
	mux.HandleFunc("DELETE /v1/databases/{name}", g.drop)
	mux.HandleFunc("GET /v1/databases/{name}/stats", g.stats)
	mux.HandleFunc("POST /v1/databases/{name}/query", g.query)
	mux.HandleFunc("POST /v1/databases/{name}/execute", g.execute)
	mux.HandleFunc("GET /v1/admin/usage", g.usage)
}

func (g *Gateway) key(w http.ResponseWriter, r *http.Request) (string, bool) {
	key := httpx.Bearer(r)
	if key == "" {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "missing bearer api key")
		return "", false
	}
	return key, true
}

// resolve authenticates the request and finds the database's node.
func (g *Gateway) resolve(w http.ResponseWriter, r *http.Request) (wire.Route, bool) {
	key, ok := g.key(w, r)
	if !ok {
		return wire.Route{}, false
	}
	rt, err := g.Routes.Resolve(r.Context(), key, r.PathValue("name"))
	if err != nil {
		writeUpstreamError(w, r, "control", err)
		return rt, false
	}
	httpx.Annotate(r.Context(), "tenant_id", rt.TenantID)
	httpx.Annotate(r.Context(), "database_id", rt.DatabaseID)
	httpx.Annotate(r.Context(), "node", rt.Node)
	return rt, true
}

// staleEpoch: the node says the database's placement changed since the
// route was cached.
func staleEpoch(err error) bool {
	var re *httpx.RemoteError
	return errors.As(err, &re) && re.Code == "stale_epoch"
}

// reresolve evicts a stale cached route and resolves it again.
func (g *Gateway) reresolve(w http.ResponseWriter, r *http.Request) (wire.Route, bool) {
	g.Routes.Evict(httpx.Bearer(r), r.PathValue("name"))
	return g.resolve(w, r)
}

func (g *Gateway) query(w http.ResponseWriter, r *http.Request) {
	r, cancel := withDeadline(r)
	defer cancel()
	// Reject what can be rejected before spending a control round trip.
	var req wire.QueryRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "bad_request", "body: "+err.Error())
		return
	}
	if req.SQL == "" || len(req.SQL) > MaxSQLBytes {
		httpx.WriteError(w, r, http.StatusBadRequest, "bad_request", "sql must be 1..65536 bytes")
		return
	}
	rt, ok := g.resolve(w, r)
	if !ok {
		return
	}
	rows, err := g.Nodes.Query(r.Context(), rt, req)
	if staleEpoch(err) {
		if rt, ok = g.reresolve(w, r); !ok {
			return
		}
		rows, err = g.Nodes.Query(r.Context(), rt, req)
	}
	if err != nil {
		writeUpstreamError(w, r, rt.Node, err)
		return
	}
	// Stream the rows through as the node sends them.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := rows.WriteTo(w); err != nil {
		slog.Warn("query answer cut short", "request_id", httpx.RequestID(r), "dep", rt.Node, "error", err.Error())
		panic(http.ErrAbortHandler) // the tenant must see an incomplete answer, not a short one
	}
}

func (g *Gateway) execute(w http.ResponseWriter, r *http.Request) {
	r, cancel := withDeadline(r)
	defer cancel()
	var req wire.ExecuteRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "bad_request", "body: "+err.Error())
		return
	}
	if len(req.Statements) == 0 && req.SQL != "" { // the older single-statement form
		req.Statements = []wire.Statement{{SQL: req.SQL, Args: req.Args}}
		req.SQL, req.Args = "", nil
	}
	if len(req.Statements) == 0 {
		httpx.WriteError(w, r, http.StatusBadRequest, "bad_request", "no statements")
		return
	}
	for _, st := range req.Statements {
		if st.SQL == "" || len(st.SQL) > MaxSQLBytes {
			httpx.WriteError(w, r, http.StatusBadRequest, "bad_request", "sql must be 1..65536 bytes")
			return
		}
	}
	rt, ok := g.resolve(w, r)
	if !ok {
		return
	}
	res, err := g.Nodes.Execute(r.Context(), rt, req)
	if staleEpoch(err) { // the node refused before running anything
		if rt, ok = g.reresolve(w, r); !ok {
			return
		}
		res, err = g.Nodes.Execute(r.Context(), rt, req)
	}
	if err != nil {
		writeUpstreamError(w, r, rt.Node, err)
		return
	}
	// Every field of the first version, plus stats and node_epoch.
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"rows_affected": res.RowsAffected, "last_insert_rowid": res.LastInsertRowID,
		"stats": map[string]int64{"statements": int64(len(req.Statements))}, "node_epoch": rt.Epoch,
	})
}

func (g *Gateway) create(w http.ResponseWriter, r *http.Request) {
	key, ok := g.key(w, r)
	if !ok {
		return
	}
	var req wire.CreateDatabaseRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "bad_request", "body: "+err.Error())
		return
	}
	d, err := g.Routes.Create(r.Context(), key, req.Name)
	if err != nil {
		writeUpstreamError(w, r, "control", err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, d)
}

func (g *Gateway) list(w http.ResponseWriter, r *http.Request) {
	key, ok := g.key(w, r)
	if !ok {
		return
	}
	out, err := g.Routes.List(r.Context(), key)
	if err != nil {
		writeUpstreamError(w, r, "control", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (g *Gateway) stats(w http.ResponseWriter, r *http.Request) {
	key, ok := g.key(w, r)
	if !ok {
		return
	}
	d, err := g.Routes.Stats(r.Context(), key, r.PathValue("name"))
	if err != nil {
		writeUpstreamError(w, r, "control", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, d)
}

func (g *Gateway) drop(w http.ResponseWriter, r *http.Request) {
	key, ok := g.key(w, r)
	if !ok {
		return
	}
	if err := g.Routes.Drop(r.Context(), key, r.PathValue("name")); err != nil {
		writeUpstreamError(w, r, "control", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (g *Gateway) usage(w http.ResponseWriter, r *http.Request) {
	key, ok := g.key(w, r)
	if !ok {
		return
	}
	window := r.URL.Query().Get("window")
	if window == "" {
		window = "24h"
	}
	out, err := g.Routes.Usage(r.Context(), key, window)
	if err != nil {
		writeUpstreamError(w, r, "control", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

// writeUpstreamError passes typed errors from control or a node through to
// the tenant, turns a dependency that ran out of time into a 504 and anything
// else into a 502.
func writeUpstreamError(w http.ResponseWriter, r *http.Request, dep string, err error) {
	var re *httpx.RemoteError
	if errors.As(err, &re) {
		switch {
		case re.Code == "draining":
			w.Header().Set("Retry-After", "2") // the node is restarting
		case re.Status == http.StatusTooManyRequests || re.Status == http.StatusServiceUnavailable:
			w.Header().Set("Retry-After", "1") // refused, not failed: come back shortly
		}
		httpx.WriteError(w, r, re.Status, re.Code, re.Message)
		return
	}
	if errors.Is(err, breaker.ErrOpen) {
		w.Header().Set("Retry-After", "5")
		httpx.WriteError(w, r, http.StatusServiceUnavailable, "control_unavailable", "control plane unavailable; known databases keep working")
		return
	}
	if errors.Is(err, httpx.ErrUnreachable) && dep != "control" {
		// Nobody listening: a node restarting (after its drain) or down.
		w.Header().Set("Retry-After", "1")
		httpx.WriteError(w, r, http.StatusServiceUnavailable, "node_unavailable", dep+" unavailable; retry shortly")
		return
	}
	if httpx.Outcome(0, err) == "timeout" {
		httpx.WriteError(w, r, http.StatusGatewayTimeout, "timeout", dep+" timeout")
		return
	}
	slog.Warn("upstream error", "request_id", httpx.RequestID(r), "dep", dep, "error", err.Error())
	httpx.WriteError(w, r, http.StatusBadGateway, "upstream_error", dep+" unavailable")
}
