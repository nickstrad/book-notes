// Package api is the gateway's tenant-facing HTTP API.
package api

import (
	"errors"
	"log"
	"net/http"

	"litefleet/internal/httpx"
	"litefleet/internal/nodeclient"
	"litefleet/internal/routes"
	"litefleet/internal/wire"
)

// MaxSQLBytes is the largest statement the platform accepts.
const MaxSQLBytes = 64 << 10

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
	rt, err := g.Routes.Resolve(key, r.PathValue("name"))
	if err != nil {
		writeUpstreamError(w, r, "control", err)
		return rt, false
	}
	return rt, true
}

func (g *Gateway) query(w http.ResponseWriter, r *http.Request) {
	rt, ok := g.resolve(w, r)
	if !ok {
		return
	}
	var req wire.QueryRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "bad_request", "body: "+err.Error())
		return
	}
	if req.SQL == "" || len(req.SQL) > MaxSQLBytes {
		httpx.WriteError(w, r, http.StatusBadRequest, "bad_request", "sql must be 1..65536 bytes")
		return
	}
	res, err := g.Nodes.Query(rt, req)
	if err != nil {
		writeUpstreamError(w, r, rt.Node, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (g *Gateway) execute(w http.ResponseWriter, r *http.Request) {
	rt, ok := g.resolve(w, r)
	if !ok {
		return
	}
	var req wire.ExecuteRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "bad_request", "body: "+err.Error())
		return
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
	res, err := g.Nodes.Execute(rt, req)
	if err != nil {
		writeUpstreamError(w, r, rt.Node, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
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
	d, err := g.Routes.Create(key, req.Name)
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
	out, err := g.Routes.List(key)
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
	d, err := g.Routes.Stats(key, r.PathValue("name"))
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
	if err := g.Routes.Drop(key, r.PathValue("name")); err != nil {
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
	out, err := g.Routes.Usage(key, window)
	if err != nil {
		writeUpstreamError(w, r, "control", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

// writeUpstreamError passes typed errors from control or a node through to
// the tenant and turns anything else into a 502.
func writeUpstreamError(w http.ResponseWriter, r *http.Request, dep string, err error) {
	var re *httpx.RemoteError
	if errors.As(err, &re) {
		httpx.WriteError(w, r, re.Status, re.Code, re.Message)
		return
	}
	log.Printf("%s: %v", dep, err)
	httpx.WriteError(w, r, http.StatusBadGateway, "upstream_error", dep+" unavailable")
}
