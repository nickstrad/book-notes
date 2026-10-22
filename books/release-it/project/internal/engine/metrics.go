package engine

import (
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

// CanaryID is the database the node's health check opens through the same
// handle path tenants use.
const CanaryID = "00000000-0000-4000-8000-000000000000"

// Canary opens (creating if needed) the canary database and reads from it.
func (e *Engine) Canary() error {
	h, err := e.Handle(CanaryID)
	if err == ErrNoSuchDatabase {
		if err := e.Provision(wireCanary); err != nil && err != ErrExists {
			return err
		}
		h, err = e.Handle(CanaryID)
	}
	if err != nil {
		return err
	}
	var n int
	return h.db.QueryRow(`SELECT count(*) FROM _meta`).Scan(&n)
}

var wireCanary = wire.ProvisionRequest{DatabaseID: CanaryID, TenantID: "_node", Epoch: 1}
