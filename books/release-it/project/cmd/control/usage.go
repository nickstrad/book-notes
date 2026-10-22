package main

import (
	"net/http"
	"sort"
	"time"

	"litefleet/internal/httpx"
	"litefleet/internal/wire"
)

// usage serves the fleet-wide per-tenant usage report the ops dashboard
// polls: statements, rows and bytes over a window, plus current storage.
func (c *control) usage(w http.ResponseWriter, r *http.Request) {
	_, admin, err := c.reg.Tenant(httpx.Bearer(r))
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

	ledger, err := c.reg.LedgerSince(time.Now().Add(-d))
	if err != nil {
		writeRegistryError(w, r, err)
		return
	}
	storage, err := c.reg.StorageByTenant()
	if err != nil {
		writeRegistryError(w, r, err)
		return
	}
	byTenant := map[string]*wire.TenantUsage{}
	for _, l := range ledger {
		u := byTenant[l.TenantID]
		if u == nil {
			u = &wire.TenantUsage{TenantID: l.TenantID}
			byTenant[l.TenantID] = u
		}
		u.Statements += l.Statements
		u.Rows += l.Rows
		u.Bytes += l.Bytes
	}
	report := wire.UsageReport{Window: window, Tenants: []wire.TenantUsage{}}
	for t, u := range byTenant {
		u.StorageBytes = storage[t]
		report.Tenants = append(report.Tenants, *u)
	}
	sort.Slice(report.Tenants, func(i, j int) bool { return report.Tenants[i].Bytes > report.Tenants[j].Bytes })
	if len(report.Tenants) > 50 {
		report.Tenants = report.Tenants[:50]
	}
	httpx.WriteJSON(w, http.StatusOK, report)
}
