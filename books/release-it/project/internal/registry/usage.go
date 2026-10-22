package registry

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"litefleet/internal/wire"
)

// recordUsage appends one ledger row per database and kind with activity in
// this heartbeat.
func recordUsage(tx *sql.Tx, dbs []wire.DatabaseReport) error {
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
	_, err := tx.Exec(`
		INSERT INTO usage_ledger (tenant_id, database_id, kind, statements, rows, bytes)
		SELECT d.tenant_id, v.id, v.kind, v.statements, v.rows, v.bytes
		FROM (VALUES `+strings.Join(vals, ", ")+`) AS v(id, kind, statements, rows, bytes)
		JOIN databases d ON d.id = v.id`, args...)
	return err
}

// LedgerRow is one row of the usage ledger.
type LedgerRow struct {
	At         time.Time
	TenantID   string
	DatabaseID string
	Kind       string
	Statements int64
	Rows       int64
	Bytes      int64
}

// LedgerSince returns every ledger row recorded after since.
func (r *Registry) LedgerSince(since time.Time) ([]LedgerRow, error) {
	rows, err := r.db.Query(`
		SELECT at, tenant_id, database_id::text, kind, statements, rows, bytes
		FROM usage_ledger WHERE at > $1`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LedgerRow
	for rows.Next() {
		var l LedgerRow
		if err := rows.Scan(&l.At, &l.TenantID, &l.DatabaseID, &l.Kind, &l.Statements, &l.Rows, &l.Bytes); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// StorageByTenant returns the current storage per tenant.
func (r *Registry) StorageByTenant() (map[string]int64, error) {
	rows, err := r.db.Query(`SELECT tenant_id, sum(size_bytes) FROM databases WHERE state = 'active' GROUP BY tenant_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int64{}
	for rows.Next() {
		var t string
		var n int64
		if err := rows.Scan(&t, &n); err != nil {
			return nil, err
		}
		out[t] = n
	}
	return out, rows.Err()
}
