package main

import (
	"context"
	"flag"
	"log"
	"time"

	"litefleet/internal/registry"
)

// purge keeps the usage ledger at the size of its retention window: every
// pass moves raw rows older than the window into the daily rollup, in
// batches with a pause between them so the purge never competes with the
// critical path for long. Run it as `control purge` (in the lab a sidecar
// every minute; in production, daily).
func purge(reg *registry.Registry, args []string) {
	fs := flag.NewFlagSet("purge", flag.ExitOnError)
	every := fs.Duration("every", time.Minute, "time between passes (0: one pass, then exit)")
	retention := fs.Duration("retention", 25*time.Hour, "raw ledger rows newer than this are kept (the usage report reads 24 h)")
	batch := fs.Int("batch", 5000, "rows per batch")
	pause := fs.Duration("pause", 250*time.Millisecond, "sleep between batches (the purge must not compete with the critical path)")
	fs.Parse(args)
	for {
		start, total := time.Now(), 0
		cutoff := time.Now().Add(-*retention)
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			n, err := reg.PurgeBatch(ctx, cutoff, *batch)
			cancel()
			if err != nil {
				log.Printf("purge: %v", err)
				break
			}
			total += n
			if n < *batch {
				break
			}
			time.Sleep(*pause)
		}
		log.Printf("purge: moved %d ledger rows older than %s into usage_daily in %s", total, cutoff.Format(time.RFC3339), time.Since(start).Round(time.Millisecond))
		if *every == 0 {
			return
		}
		time.Sleep(*every)
	}
}
