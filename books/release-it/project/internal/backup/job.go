// Package backup snapshots tenant databases to the blob store.
package backup

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"litefleet/internal/blob"
	"litefleet/internal/engine"
)

var (
	backupDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "backup_duration_seconds", Help: "Time to snapshot and upload one database.",
		Buckets: prometheus.ExponentialBuckets(0.005, 2, 16),
	})
	writerLockHold = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "writer_lock_hold_seconds", Help: "Time a backup held a database's write lock.",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 16),
	})
)

// Job uploads a snapshot of every database written since its last backup.
type Job struct {
	eng      *engine.Engine
	blob     *blob.Client
	interval time.Duration
	tmp      string
}

// New returns a backup job.
func New(eng *engine.Engine, b *blob.Client, interval time.Duration, tmp string) *Job {
	return &Job{eng: eng, blob: b, interval: interval, tmp: tmp}
}

// Run backs up dirty databases every interval, one at a time, spread over
// the interval so the snapshots do not land together.
func (j *Job) Run() {
	for range time.Tick(j.interval) {
		dirty := j.eng.TakeEvictedDirty() // closed by the handle cache before their backup
		for _, h := range j.eng.Open() {
			if h.Dirty() {
				dirty = append(dirty, h.ID)
			}
		}
		gap := j.interval / time.Duration(len(dirty)+1)
		for _, id := range dirty {
			if _, err := j.Backup(id); err != nil {
				log.Printf("backup %s: %v", id, err)
			}
			time.Sleep(gap)
		}
	}
}

// Backup snapshots one database and uploads it as the next generation.
//
// The write lock covers only taking the snapshot (VACUUM INTO writes a
// consistent copy in milliseconds). The upload runs after the lock is
// released, so a slow blob store delays the backup, not the tenant's writes.
// The generation is recorded in _meta once the upload has succeeded.
func (j *Job) Backup(id string) (int, error) {
	defer func(start time.Time) { backupDuration.Observe(time.Since(start).Seconds()) }(time.Now())
	h, err := j.eng.Handle(id)
	if err != nil {
		return 0, err
	}
	defer h.Release()
	ctx := context.Background()
	gen := h.Generation() + 1
	tmp := filepath.Join(j.tmp, fmt.Sprintf("%s.%d.db", id, gen))
	os.Remove(tmp)
	if err := j.snapshot(ctx, h, tmp); err != nil {
		return 0, err
	}
	defer os.Remove(tmp)
	if _, err := j.blob.Put(id, gen, tmp); err != nil {
		return 0, fmt.Errorf("upload: %w", err)
	}
	if _, err := h.DB().ExecContext(ctx, "UPDATE _meta SET value = ? WHERE key = 'backup_generation'", strconv.Itoa(gen)); err != nil {
		return 0, err
	}
	h.BackedUp(gen, time.Now())
	return gen, nil
}

// snapshot copies the database to path while holding its write lock.
func (j *Job) snapshot(ctx context.Context, h *engine.Handle, path string) error {
	lock, err := h.DB().Conn(ctx)
	if err != nil {
		return err
	}
	defer lock.Close()
	if _, err := lock.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return fmt.Errorf("lock: %w", err)
	}
	held := time.Now()
	defer func() {
		lock.ExecContext(ctx, "ROLLBACK")
		writerLockHold.Observe(time.Since(held).Seconds())
	}()
	if _, err := h.DB().ExecContext(ctx, "VACUUM INTO ?", path); err != nil {
		return fmt.Errorf("snapshot: %w", err)
	}
	return nil
}
