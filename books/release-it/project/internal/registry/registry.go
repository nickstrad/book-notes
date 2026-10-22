// Package registry is the control plane's Postgres registry: tenants, API
// keys, nodes, databases and their placement, and the usage ledger.
package registry

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"

	"litefleet/internal/httpx"
	"litefleet/internal/wire"
)

var (
	ErrUnauthorized   = errors.New("unknown api key")
	ErrNoSuchDatabase = errors.New("no such database")
	ErrExists         = errors.New("database already exists")
)

// Registry wraps the control plane's connection pool.
//
// Three pools from one DSN, so one kind of work cannot starve another: oltp
// for the short statements of the critical path (resolve, create, heartbeat),
// reporting for the usage report, and a single connection for health checks.
type Registry struct {
	db        *sql.DB
	reporting *sql.DB
	health    *sql.DB
}

// StatementTimeout is the most any registry statement may run; Postgres
// cancels it server-side even if control has already given up.
const StatementTimeout = 2 * time.Second

// ReportTimeout is the most a usage-report statement may run.
const ReportTimeout = 5 * time.Second

// Open connects to Postgres.
func Open(dsn string) (*Registry, error) {
	oltp, err := openPool(dsn, "oltp", 8, StatementTimeout)
	if err != nil {
		return nil, err
	}
	reporting, err := openPool(dsn, "reporting", 2, ReportTimeout)
	if err != nil {
		return nil, err
	}
	health, err := openPool(dsn, "health", 1, 200*time.Millisecond)
	if err != nil {
		return nil, err
	}
	return &Registry{db: oltp, reporting: reporting, health: health}, nil
}

// openPool opens a bounded pool: Postgres serves 100 connections for
// everyone, and a control plane that opens one per concurrent request turns
// a burst into a Postgres outage. Waits are bounded by the caller's context.
func openPool(dsn, name string, size int, timeout time.Duration) (*sql.DB, error) {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.RuntimeParams["statement_timeout"] = strconv.Itoa(int(timeout.Milliseconds()))
	cfg.RuntimeParams["application_name"] = "control-" + name
	db := stdlib.OpenDB(*cfg)
	db.SetMaxOpenConns(size)
	db.SetMaxIdleConns(size)
	db.SetConnMaxLifetime(30 * time.Minute)
	exportPool(db, name)
	return db, nil
}

// DB exposes the pool (migrations, health).
func (r *Registry) DB() *sql.DB { return r.db }

// Ping checks that Postgres answers, on the health check's own connection,
// so a saturated oltp pool is not mistaken for a dead database.
func (r *Registry) Ping(ctx context.Context) (err error) {
	defer observe(time.Now(), &err)
	return r.health.PingContext(ctx)
}

// observe records one registry call as a call to the postgres dependency.
// Answers such as "unknown key" are successful calls.
func observe(start time.Time, err *error) {
	e := *err
	if errors.Is(e, ErrUnauthorized) || errors.Is(e, ErrNoSuchDatabase) || errors.Is(e, ErrExists) {
		e = nil
	}
	httpx.DependencyMetric("postgres", start, 0, e)
}

// HashKey is how API keys are stored.
func HashKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

// Tenant returns the tenant that owns an API key.
func (r *Registry) Tenant(ctx context.Context, apiKey string) (tenantID string, admin bool, err error) {
	defer observe(time.Now(), &err)
	err = r.db.QueryRowContext(ctx, `SELECT tenant_id, admin FROM api_keys WHERE key_hash = $1`, HashKey(apiKey)).Scan(&tenantID, &admin)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, ErrUnauthorized
	}
	return tenantID, admin, err
}

// Resolve authenticates a key and locates one of its databases in one round
// trip.
func (r *Registry) Resolve(ctx context.Context, apiKey, dbName string) (rt wire.Route, err error) {
	defer observe(time.Now(), &err)
	var dbID, url, node sql.NullString
	var epoch sql.NullInt64
	err = r.db.QueryRowContext(ctx, `
		SELECT k.tenant_id, d.id::text, n.url, n.name, p.epoch
		FROM api_keys k
		LEFT JOIN databases d ON d.tenant_id = k.tenant_id AND d.name = $2 AND d.state = 'active'
		LEFT JOIN placement p ON p.database_id = d.id
		LEFT JOIN nodes n ON n.name = p.owner_node
		WHERE k.key_hash = $1`, HashKey(apiKey), dbName).Scan(&rt.TenantID, &dbID, &url, &node, &epoch)
	if errors.Is(err, sql.ErrNoRows) {
		return rt, ErrUnauthorized
	}
	if err != nil {
		return rt, err
	}
	if !dbID.Valid || !url.Valid {
		return rt, ErrNoSuchDatabase
	}
	rt.DatabaseID, rt.NodeURL, rt.Node, rt.Epoch = dbID.String, url.String, node.String, int(epoch.Int64)
	return rt, nil
}

// NodeLoad returns every node with a recent heartbeat and how many databases
// it holds.
func (r *Registry) NodeLoad(ctx context.Context) (_ map[string]int, err error) {
	defer observe(time.Now(), &err)
	rows, err := r.db.QueryContext(ctx, `
		SELECT n.name, count(p.database_id)
		FROM nodes n LEFT JOIN placement p ON p.owner_node = n.name
		WHERE n.last_heartbeat_at > now() - interval '30 seconds'
		GROUP BY n.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var name string
		var n int
		if err := rows.Scan(&name, &n); err != nil {
			return nil, err
		}
		out[name] = n
	}
	return out, rows.Err()
}

// BeginCreate records a database in state 'provisioning' and returns its id.
func (r *Registry) BeginCreate(ctx context.Context, tenantID, name string) (id string, err error) {
	defer observe(time.Now(), &err)
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO databases (id, tenant_id, name, state)
		VALUES (gen_random_uuid(), $1, $2, 'provisioning')
		RETURNING id::text`, tenantID, name).Scan(&id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return "", ErrExists
	}
	return id, err
}

// FinishCreate publishes the placement once the node has confirmed the file.
func (r *Registry) FinishCreate(ctx context.Context, dbID, node string, epoch int) (err error) {
	defer observe(time.Now(), &err)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO placement (database_id, owner_node, epoch) VALUES ($1, $2, $3)`, dbID, node, epoch); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE databases SET state = 'active' WHERE id = $1`, dbID); err != nil {
		return err
	}
	return tx.Commit()
}

// AbortCreate removes a database whose provisioning failed.
func (r *Registry) AbortCreate(ctx context.Context, dbID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM databases WHERE id = $1 AND state = 'provisioning'`, dbID)
	return err
}

// ListDatabases returns a tenant's active databases.
func (r *Registry) ListDatabases(ctx context.Context, tenantID string) (_ []wire.Database, err error) {
	defer observe(time.Now(), &err)
	rows, err := r.db.QueryContext(ctx, `
		SELECT d.id::text, d.name, p.owner_node, d.state,
		       coalesce(d.bytes_total_v2, d.bytes_total), coalesce(d.rows_total_v2, d.rows_total), d.size_bytes,
		       coalesce(to_char(d.last_backup_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'), '')
		FROM databases d JOIN placement p ON p.database_id = d.id
		WHERE d.tenant_id = $1 AND d.state = 'active'
		ORDER BY d.name`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []wire.Database
	for rows.Next() {
		var d wire.Database
		if err := rows.Scan(&d.ID, &d.Name, &d.Node, &d.State, &d.BytesTotal, &d.RowsTotal, &d.SizeBytes, &d.LastBackup); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// Database returns one of a tenant's databases.
func (r *Registry) Database(ctx context.Context, tenantID, name string) (wire.Database, error) {
	dbs, err := r.ListDatabases(ctx, tenantID)
	if err != nil {
		return wire.Database{}, err
	}
	for _, d := range dbs {
		if d.Name == name {
			return d, nil
		}
	}
	return wire.Database{}, ErrNoSuchDatabase
}

// MarkDropped retires a database after its node has deleted the file.
func (r *Registry) MarkDropped(ctx context.Context, dbID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM placement WHERE database_id = $1`, dbID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE databases SET state = 'dropped', name = name || '.dropped.' || id::text WHERE id = $1`, dbID); err != nil {
		return err
	}
	return tx.Commit()
}

// Heartbeat records a node's liveness, its databases' sizes and activity, and
// appends the activity to the usage ledger.
func (r *Registry) Heartbeat(ctx context.Context, hb wire.Heartbeat) (err error) {
	defer observe(time.Now(), &err)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO nodes (name, url, last_heartbeat_at, fds, disk_free_bytes, wire_version)
		VALUES ($1, $2, now(), $3, $4, $5)
		ON CONFLICT (name) DO UPDATE SET url = $2, last_heartbeat_at = now(), fds = $3, disk_free_bytes = $4, wire_version = $5`,
		hb.Node, hb.URL, hb.FDs, hb.DiskFree, max(hb.WireVersion, 1)); err != nil {
		return fmt.Errorf("upsert node: %w", err)
	}
	for _, d := range hb.Databases {
		var lastBackup any
		if d.LastBackup != "" {
			if t, err := time.Parse(time.RFC3339, d.LastBackup); err == nil {
				lastBackup = t
			}
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE databases SET size_bytes = $2,
			       bytes_total_v2 = coalesce(bytes_total_v2, bytes_total) + $3,
			       rows_total_v2 = coalesce(rows_total_v2, rows_total) + $4,
			       -- expand phase: the 32-bit columns are kept for older readers and saturate
			       bytes_total = least(bytes_total::bigint + $3, 2147483647),
			       rows_total = least(rows_total::bigint + $4, 2147483647),
			       last_backup_at = coalesce($5, last_backup_at)
			WHERE id = $1`,
			d.ID, d.Bytes, d.ReadBytes+d.WriteBytes, d.ReadRows+d.WriteRows, lastBackup); err != nil {
			return fmt.Errorf("update database %s: %w", d.ID, err)
		}
	}
	if err := recordUsage(ctx, tx, hb.Databases); err != nil {
		return fmt.Errorf("record usage: %w", err)
	}
	return tx.Commit()
}
