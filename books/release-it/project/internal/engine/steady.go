package engine

import (
	"context"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Steady state: nothing on a node may grow without bound.
//
//   - WAL files: a WAL holds a full page image for every page a write
//     touches, so it grows tens of times faster than the data. The
//     checkpointer copies it into the database and truncates it, off the
//     tenants' write path, once it passes CheckpointBytes or its database
//     has been idle for CheckpointIdle.
//   - Open handles: at most HandleCache, least recently used closed first,
//     and any handle idle for HandleIdle closed.
const (
	CheckpointBytes = 64 << 10
	CheckpointIdle  = 30 * time.Second
	HandleCache     = 128
	HandleIdle      = 5 * time.Minute
)

var checkpoints = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "sqlite_checkpoints_total", Help: "WAL checkpoints run by the node's checkpointer, by reason and outcome.",
}, []string{"reason", "outcome"})

// Maintain runs the checkpointer every 2 s and closes idle handles every
// 30 s, until the process exits.
func (e *Engine) Maintain() {
	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()
	for n := 0; ; n++ {
		<-tick.C
		e.checkpoint()
		if n%15 == 0 {
			e.closeIdle()
		}
	}
}

func (e *Engine) checkpoint() {
	for _, h := range e.Open() {
		fi, err := os.Stat(h.Path + "-wal")
		if err != nil || fi.Size() == 0 {
			continue
		}
		reason := "size"
		if fi.Size() < CheckpointBytes {
			if time.Since(time.Unix(0, h.lastUsed.Load())) < CheckpointIdle {
				continue
			}
			reason = "idle"
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, err = h.db.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)")
		cancel()
		outcome := "ok"
		if err != nil {
			outcome = "error" // busy, or the handle closed meanwhile; next pass
		}
		checkpoints.WithLabelValues(reason, outcome).Inc()
	}
}

func (e *Engine) closeIdle() {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, id := range e.handles.Keys() {
		if h, ok := e.handles.Peek(id); ok && h.refs.Load() == 0 && time.Since(time.Unix(0, h.lastUsed.Load())) > HandleIdle {
			e.handles.Remove(id)
		}
	}
}
