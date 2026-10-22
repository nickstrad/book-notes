package main

import (
	"fmt"
	"strconv"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"litefleet/internal/httpx"
)

// A node's health is made of physical facts: can it open a database, how
// close is it to its file-descriptor limit, how full is its volume, how much
// WAL is waiting to be checkpointed.

func (n *node) canaryCheck() (string, string) {
	if err := n.eng.Canary(); err != nil {
		return httpx.Failed, "canary: " + err.Error()
	}
	return httpx.OK, ""
}

func (n *node) fdCheck() (string, string) {
	used, limit := openFDs(), fdLimit()
	if limit > 0 && used*5 >= limit*4 {
		return httpx.Degraded, fmt.Sprintf("fds %d/%d", used, limit)
	}
	return httpx.OK, ""
}

func (n *node) diskCheck() (string, string) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(n.data, &st); err != nil {
		return httpx.Failed, "statfs: " + err.Error()
	}
	if st.Bavail == 0 {
		return httpx.Failed, "data volume full"
	}
	return httpx.OK, ""
}

func (n *node) walCheck() (string, string) {
	limit, _ := strconv.ParseInt(env("WAL_DEGRADED_BYTES", strconv.Itoa(128<<20)), 10, 64)
	if wal := n.eng.WALBytes(); wal > limit {
		return httpx.Degraded, fmt.Sprintf("wal %d MB", wal>>20)
	}
	return httpx.OK, ""
}

func fdLimit() int {
	var rl syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rl); err != nil {
		return 0
	}
	return int(rl.Cur)
}

func (n *node) exportMetrics() {
	n.eng.ExportMetrics()
	promauto.NewGaugeFunc(prometheus.GaugeOpts{Name: "usage_queue_depth", Help: "Usage events waiting to ship to metering."},
		func() float64 { return float64(n.usage.Len()) })
	promauto.NewGaugeFunc(prometheus.GaugeOpts{Name: "usage_queue_oldest_age_seconds", Help: "Age of the oldest waiting usage event."},
		func() float64 { return n.usage.OldestAge().Seconds() })
}
