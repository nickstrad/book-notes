package main

import (
	"net/http"
	"time"

	"litefleet/internal/httpx"
	"litefleet/internal/wire"
)

// reports admits two usage reports at a time; the dashboard's refreshes
// beyond that get 503 at once instead of queueing for the reporting pool.
var reports = make(chan struct{}, 2)

// usage serves the fleet-wide per-tenant usage report the ops dashboard
// polls: statements, rows and bytes over a window, plus current storage.
func (c *control) usage(w http.ResponseWriter, r *http.Request) {
	select {
	case reports <- struct{}{}:
		defer func() { <-reports }()
	default:
		w.Header().Set("Retry-After", "5")
		httpx.WriteError(w, r, http.StatusServiceUnavailable, "report_busy", "two usage reports are already running")
		return
	}
	_, admin, err := c.reg.Tenant(r.Context(), httpx.Bearer(r))
	if err != nil || !admin {
		httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "admin key required")
		return
	}
	window := r.URL.Query().Get("window")
	if window == "" {
		window = "24h"
	}
	d, err := time.ParseDuration(window)
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "bad_request", "bad window")
		return
	}

	tenants, err := c.reg.UsageReport(r.Context(), time.Now().Add(-d))
	if err != nil {
		writeRegistryError(w, r, err)
		return
	}
	report := wire.UsageReport{Window: window, Tenants: tenants}
	httpx.WriteJSON(w, http.StatusOK, report)
}
