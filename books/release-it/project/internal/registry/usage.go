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
		vals = append(vals, fmt.Sprintf("($%d::uuid, $%d, $%d::int, $%d::int, $%d::int)", n+1, n+2, n+3, n+4, n+5))
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
		INSERT INTO usage_ledger (tenant_id, database_id, kind, statements, rows, bytes)
		SELECT d.tenant_id, v.id, v.kind, v.statements, v.rows, v.bytes
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
		SELECT l.tenant_id, sum(l.statements), sum(l.rows), sum(l.bytes), coalesce(max(s.storage), 0)
		FROM usage_ledger l
		LEFT JOIN (SELECT tenant_id, sum(size_bytes) AS storage FROM databases
		           WHERE state = 'active' GROUP BY tenant_id) s ON s.tenant_id = l.tenant_id
		WHERE l.at > $1
		GROUP BY l.tenant_id
		ORDER BY sum(l.bytes) DESC
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
