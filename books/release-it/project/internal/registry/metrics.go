package registry

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// exportPool publishes a connection pool's sql.DBStats as db_pool_*{pool}.
func exportPool(db *sql.DB, pool string) {
	labels := prometheus.Labels{"pool": pool}
	gauge := func(name, help string, f func(sql.DBStats) float64) {
		promauto.NewGaugeFunc(prometheus.GaugeOpts{Name: name, Help: help, ConstLabels: labels}, func() float64 { return f(db.Stats()) })
	}
	counter := func(name, help string, f func(sql.DBStats) float64) {
		promauto.NewCounterFunc(prometheus.CounterOpts{Name: name, Help: help, ConstLabels: labels}, func() float64 { return f(db.Stats()) })
	}
	gauge("db_pool_max_open", "Pool size limit (0 = unbounded).", func(s sql.DBStats) float64 { return float64(s.MaxOpenConnections) })
	gauge("db_pool_open_connections", "Connections open.", func(s sql.DBStats) float64 { return float64(s.OpenConnections) })
	gauge("db_pool_in_use", "Connections checked out.", func(s sql.DBStats) float64 { return float64(s.InUse) })
	gauge("db_pool_idle", "Connections idle in the pool.", func(s sql.DBStats) float64 { return float64(s.Idle) })
	counter("db_pool_wait_count_total", "Checkouts that had to wait for a connection.", func(s sql.DBStats) float64 { return float64(s.WaitCount) })
	counter("db_pool_wait_seconds_total", "Time spent waiting for a connection.", func(s sql.DBStats) float64 { return s.WaitDuration.Seconds() })
}
