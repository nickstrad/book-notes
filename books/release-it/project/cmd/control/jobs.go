package main

import (
	"context"
	"flag"
	"log"
	"time"

	"litefleet/internal/registry"
	"litefleet/migrations"
)

// migrate is `control migrate`: a one-shot job run before a rollout, never
// by the serving process. It applies the main migrations (each in a
// transaction) and the online ones (outside one). With -contract it then
// applies the guarded contract migrations, which refuse to run until nothing
// depends on the old shape.
func migrate(reg *registry.Registry, args []string) {
	fs := flag.NewFlagSet("migrate", flag.ExitOnError)
	contract := fs.Bool("contract", false, "also apply contract/ migrations (after the rollout and the backfill)")
	fs.Parse(args)
	steps := []struct {
		dir  string
		inTx bool
	}{{".", true}, {"online", false}}
	if *contract {
		steps = append(steps, struct {
			dir  string
			inTx bool
		}{"contract", true})
	}
	for _, s := range steps {
		if err := reg.Migrate(migrations.FS, s.dir, s.inTx); err != nil {
			log.Fatalf("migrate %s: %v", s.dir, err)
		}
	}
	log.Printf("migrations up to date")
}

// backfill is `control backfill`: copies the 32-bit counters into their
// 64-bit columns for rows written before the expand, in id ranges of 5 000
// with a pause, so it never holds many rows or competes for long.
func backfill(reg *registry.Registry, args []string) {
	fs := flag.NewFlagSet("backfill", flag.ExitOnError)
	batch := fs.Int("batch", 5000, "rows per batch")
	pause := fs.Duration("pause", 100*time.Millisecond, "sleep between batches")
	fs.Parse(args)
	start, total := time.Now(), 0
	for from := int64(0); ; from += int64(*batch) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		n, more, err := reg.BackfillBatch(ctx, from, int64(*batch))
		cancel()
		if err != nil {
			log.Fatalf("backfill: %v", err)
		}
		total += n
		if !more {
			break
		}
		time.Sleep(*pause)
	}
	log.Printf("backfill: %d rows in %s", total, time.Since(start).Round(time.Millisecond))
}
