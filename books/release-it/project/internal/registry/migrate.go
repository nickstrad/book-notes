package registry

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"path"
	"sort"
)

// Migrate applies every migration under dir in fsys that has not been
// applied yet, in file-name order. inTx runs each in a transaction; online
// migrations (CREATE INDEX CONCURRENTLY) cannot be.
//
// It runs on one connection with no statement timeout (an online index
// build takes minutes) but a 3 s lock timeout: a migration that cannot get
// its lock fails, to be retried, instead of queueing every other query on
// the table behind it while it waits.
func (r *Registry) Migrate(fsys fs.FS, dir string, inTx bool) error {
	ctx := context.Background()
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `SET statement_timeout = 0`); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, `SET lock_timeout = '3s'`); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		return err
	}
	names, err := fs.Glob(fsys, path.Join(dir, "*.sql"))
	if err != nil {
		return err
	}
	sort.Strings(names)
	for _, name := range names {
		var done bool
		if err := conn.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, name).Scan(&done); err != nil {
			return err
		}
		if done {
			continue
		}
		body, err := fs.ReadFile(fsys, name)
		if err != nil {
			return err
		}
		if inTx {
			tx, err := conn.BeginTx(ctx, nil)
			if err != nil {
				return err
			}
			if _, err := tx.Exec(string(body)); err != nil {
				tx.Rollback()
				return fmt.Errorf("migration %s: %w", name, err)
			}
			if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
				tx.Rollback()
				return err
			}
			if err := tx.Commit(); err != nil {
				return err
			}
		} else {
			if _, err := conn.ExecContext(ctx, string(body)); err != nil {
				return fmt.Errorf("migration %s: %w", name, err)
			}
			if _, err := conn.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
				return err
			}
		}
		log.Printf("applied migration %s", name)
	}
	return nil
}
