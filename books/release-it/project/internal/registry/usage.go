package registry

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"litefleet/internal/wire"
)

// recordUsage appends one ledger row per database and kind with activity in
// this heartbeat.
func recordUsage(ctx context.Context, tx *sql.Tx, dbs []wire.DatabaseReport) error {
	var vals []string
	var args []any
	add := func(id, kind string, statements, rows, bytes int64) {
		if statements == 0 {
			return
		}
		n := len(args)
		vals = append(vals, fmt.Sprintf("($%d::uuid, $%d, $%d::bigint, $%d::bigint, $%d::bigint)", n+1, n+2, n+3, n+4, n+5))
		args = append(args, id, kind, statements, rows, bytes)
	}
	for _, d := range dbs {
		add(d.ID, "read", d.Reads, d.ReadRows, d.ReadBytes)
		add(d.ID, "write", d.Writes, d.WriteRows, d.WriteBytes)
	}
	if len(vals) == 0 {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO usage_ledger (tenant_id, database_id, kind, statements, rows, bytes, rows_v2, bytes_v2)
		SELECT d.tenant_id, v.id, v.kind, v.statements,
		       least(v.rows, 2147483647), least(v.bytes, 2147483647), v.rows, v.bytes
		FROM (VALUES `+strings.Join(vals, ", ")+`) AS v(id, kind, statements, rows, bytes)
		JOIN databases d ON d.id = v.id`, args...)
	return err
}

// UsageReport aggregates the ledger in Postgres: the window's statements,
// rows and bytes per tenant, plus current storage, the 50 heaviest tenants
// first. Only 50 rows cross the wire, however large the ledger.
func (r *Registry) UsageReport(ctx context.Context, since time.Time) (_ []wire.TenantUsage, err error) {
	defer observe(time.Now(), &err)
	rows, err := r.reporting.QueryContext(ctx, `
		SELECT l.tenant_id, sum(l.statements), sum(coalesce(l.rows_v2, l.rows)), sum(coalesce(l.bytes_v2, l.bytes)),
		       coalesce(max(s.storage), 0)
		FROM usage_ledger l
		LEFT JOIN (SELECT tenant_id, sum(size_bytes) AS storage FROM databases
		           WHERE state = 'active' GROUP BY tenant_id) s ON s.tenant_id = l.tenant_id
		WHERE l.at > $1
		GROUP BY l.tenant_id
		ORDER BY 4 DESC
		LIMIT 50`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []wire.TenantUsage{}
	for rows.Next() {
		var u wire.TenantUsage
		if err := rows.Scan(&u.TenantID, &u.Statements, &u.Rows, &u.Bytes, &u.StorageBytes); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// PurgeBatch rolls up to n ledger rows older than cutoff into usage_daily and
// deletes them, in one statement, and returns how many it moved.
func (r *Registry) PurgeBatch(ctx context.Context, cutoff time.Time, n int) (moved int, err error) {
	defer observe(time.Now(), &err)
	err = r.db.QueryRowContext(ctx, `
		WITH old AS (
			DELETE FROM usage_ledger
			WHERE id IN (SELECT id FROM usage_ledger WHERE at < $1 ORDER BY at LIMIT $2)
			RETURNING tenant_id, at, statements, coalesce(rows_v2, rows) AS rows, coalesce(bytes_v2, bytes) AS bytes
		), rolled AS (
			INSERT INTO usage_daily AS d (tenant_id, day, statements, rows, bytes)
			SELECT tenant_id, at::date, sum(statements), sum(rows), sum(bytes) FROM old GROUP BY 1, 2
			ON CONFLICT (tenant_id, day) DO UPDATE SET
				statements = d.statements + excluded.statements,
				rows = d.rows + excluded.rows,
				bytes = d.bytes + excluded.bytes
			RETURNING 1
		)
		SELECT count(*) FROM old`, cutoff, n).Scan(&moved)
	return moved, err
}

// BackfillBatch copies the 32-bit counters into the 64-bit columns for ledger
// rows with ids in [from, from+n) (and, on the first batch, every database),
// and reports whether ledger rows beyond the range remain.
func (r *Registry) BackfillBatch(ctx context.Context, from, n int64) (filled int, more bool, err error) {
	defer observe(time.Now(), &err)
	if from == 0 {
		res, err := r.db.ExecContext(ctx, `UPDATE databases SET bytes_total_v2 = bytes_total, rows_total_v2 = rows_total
			WHERE bytes_total_v2 IS NULL OR rows_total_v2 IS NULL`)
		if err != nil {
			return 0, false, err
		}
		k, _ := res.RowsAffected()
		filled += int(k)
	}
	res, err := r.db.ExecContext(ctx, `UPDATE usage_ledger SET bytes_v2 = bytes, rows_v2 = rows
		WHERE id >= $1 AND id < $2 AND (bytes_v2 IS NULL OR rows_v2 IS NULL)`, from, from+n)
	if err != nil {
		return filled, false, err
	}
	k, _ := res.RowsAffected()
	err = r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM usage_ledger WHERE id >= $1)`, from+n).Scan(&more)
	return filled + int(k), more, err
}
