package engine

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"time"

	"litefleet/internal/wire"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var statementDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "statement_duration_seconds", Help: "Time to run a tenant statement on this node.",
	Buckets: prometheus.ExponentialBuckets(0.0005, 2, 15),
}, []string{"kind"})

func observeStatement(kind string, start time.Time) {
	statementDuration.WithLabelValues(kind).Observe(time.Since(start).Seconds())
}

// ExportMetrics publishes the engine's gauges: open databases, SQLite
// connections, and WAL bytes on disk.
func (e *Engine) ExportMetrics() {
	promauto.NewGaugeFunc(prometheus.GaugeOpts{Name: "sqlite_open_databases", Help: "Tenant databases with an open handle."},
		func() float64 { return float64(len(e.Open())) })
	promauto.NewGaugeFunc(prometheus.GaugeOpts{Name: "sqlite_open_connections", Help: "SQLite connections open across all handles."},
		func() float64 {
			n := 0
			for _, h := range e.Open() {
				n += h.db.Stats().OpenConnections
			}
			return float64(n)
		})
	promauto.NewGaugeFunc(prometheus.GaugeOpts{Name: "sqlite_wal_bytes_total", Help: "Bytes in the WAL files of open databases."},
		func() float64 { return float64(e.WALBytes()) })
}

// WALBytes is the total size of the open databases' WAL files.
func (e *Engine) WALBytes() int64 {
	var n int64
	for _, h := range e.Open() {
		if fi, err := os.Stat(h.Path + "-wal"); err == nil {
			n += fi.Size()
		}
	}
	return n
}

// CanaryID is the database the node's health check opens.
const CanaryID = "00000000-0000-4000-8000-000000000000"

// Canary reads from the node's canary database on a dedicated handle of one
// connection, outside the tenant handle cache and its budgets, so a busy
// tenant path cannot make a healthy node look dead (or the other way round).
func (e *Engine) Canary() error {
	e.canaryOnce.Do(func() {
		if _, err := os.Stat(e.path(CanaryID)); err != nil {
			e.Provision(wireCanary)
		}
		db, err := sql.Open("sqlite3", dsn(e.path(CanaryID)))
		if err == nil {
			db.SetMaxOpenConns(1)
			e.canary = db
		}
	})
	if e.canary == nil {
		return errors.New("canary database unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	var n int
	return e.canary.QueryRowContext(ctx, `SELECT count(*) FROM _meta`).Scan(&n)
}

var wireCanary = wire.ProvisionRequest{DatabaseID: CanaryID, TenantID: "_node", Epoch: 1}
