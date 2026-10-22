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

	"litefleet/internal/blob"
	"litefleet/internal/engine"
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
		var dirty []string
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
// The write lock is held across the snapshot, the upload and recording the
// generation in _meta, so a generation recorded in the file always names an
// uploaded snapshot of exactly that state.
func (j *Job) Backup(id string) (int, error) {
	h, err := j.eng.Handle(id)
	if err != nil {
		return 0, err
	}
	ctx := context.Background()
	lock, err := h.DB().Conn(ctx)
	if err != nil {
		return 0, err
	}
	defer lock.Close()
	if _, err := lock.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return 0, fmt.Errorf("lock: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			lock.ExecContext(ctx, "ROLLBACK")
		}
	}()

	gen := h.Generation() + 1
	tmp := filepath.Join(j.tmp, fmt.Sprintf("%s.%d.db", id, gen))
	os.Remove(tmp)
	if _, err := h.DB().ExecContext(ctx, "VACUUM INTO ?", tmp); err != nil {
		return 0, fmt.Errorf("snapshot: %w", err)
	}
	defer os.Remove(tmp)
	if _, err := j.blob.Put(id, gen, tmp); err != nil {
		return 0, fmt.Errorf("upload: %w", err)
	}
	if _, err := lock.ExecContext(ctx, "UPDATE _meta SET value = ? WHERE key = 'backup_generation'", strconv.Itoa(gen)); err != nil {
		return 0, err
	}
	if _, err := lock.ExecContext(ctx, "COMMIT"); err != nil {
		return 0, err
	}
	committed = true
	h.BackedUp(gen, time.Now())
	return gen, nil
}
