// Command seedgen loads the lab's data set: 200 tenants with 240 databases
// spread over the two nodes (SQLite files written straight onto the node
// volumes and registered in Postgres in the same run), one whale tenant with a
// 60 MB events table (15 000 rows of 4 KB), a 5 MB "bulky" database, and a usage ledger with
// millions of historical rows. It writes results/manifest.json, which the load
// generator reads to know the tenants' keys and where their databases live.
package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/ncruces/go-sqlite3/driver"

	"litefleet/internal/engine"
)

type dbEntry struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Node string `json:"node"`
}

type tenant struct {
	ID    string    `json:"id"`
	Key   string    `json:"key"`
	Class string    `json:"class"`
	DBs   []dbEntry `json:"dbs"`
}

type manifest struct {
	AdminKey string   `json:"admin_key"`
	Tenants  []tenant `json:"tenants"`
}

var nodes = []string{"node-a", "node-b"}

func main() {
	dsn := flag.String("dsn", "postgres://litefleet:litefleet@postgres:5432/litefleet", "registry DSN")
	nodeDir := flag.String("nodes", "/nodes", "directory holding one volume per node")
	out := flag.String("manifest", "/results/manifest.json", "manifest path")
	nTenants := flag.Int("tenants", 200, "tenants")
	whale := flag.Bool("whale", true, "seed the whale's 60 MB events table")
	ledgerRows := flag.Int("ledger", 5_000_000, "historical usage_ledger rows")
	flag.Parse()
	start := time.Now()

	pg, err := sql.Open("pgx", *dsn)
	if err != nil {
		log.Fatal(err)
	}
	waitForSchema(pg)
	must(exec(pg, `TRUNCATE usage_ledger, placement, databases, api_keys, tenants RESTART IDENTITY CASCADE`))
	for i, n := range nodes {
		must(exec(pg, `INSERT INTO nodes (name, url, last_heartbeat_at) VALUES ($1, $2, now())
			ON CONFLICT (name) DO NOTHING`, n, fmt.Sprintf("http://toxiproxy:%d", 9071+i)))
	}
	for _, n := range nodes {
		dir := filepath.Join(*nodeDir, n, "tenants")
		os.RemoveAll(dir)
		must(os.MkdirAll(dir, 0o755))
	}

	m := manifest{AdminKey: "lfk_admin_ops"}
	must(exec(pg, `INSERT INTO tenants (id, name, plan) VALUES ('ops', 'operations', 'internal')`))
	must(exec(pg, `INSERT INTO api_keys (key_hash, tenant_id, admin) VALUES ($1, 'ops', true)`, hash(m.AdminKey)))

	dbIndex := 0
	for i := 1; i <= *nTenants; i++ {
		t := tenant{ID: fmt.Sprintf("t%04d", i), Class: "normal"}
		t.Key = "lfk_" + t.ID + "_" + hash("seed:" + t.ID)[:16]
		switch i {
		case 1:
			t.Class = "whale"
		case 2:
			t.Class = "noisy"
		case 3:
			t.Class = "bulky"
		}
		must(exec(pg, `INSERT INTO tenants (id, name) VALUES ($1, $2)`, t.ID, "tenant "+t.ID))
		must(exec(pg, `INSERT INTO api_keys (key_hash, tenant_id) VALUES ($1, $2)`, hash(t.Key), t.ID))
		names := []string{"default"}
		if i%10 == 5 {
			names = append(names, "analytics", "staging")
		}
		for _, name := range names {
			node := nodes[dbIndex%2]
			if t.Class != "normal" {
				node = "node-a" // the special tenants all live on node-a
			}
			dbIndex++
			id := uuid(t.ID + "/" + name)
			path := filepath.Join(*nodeDir, node, "tenants", id+".db")
			must(writeTenantFile(path, id, t.ID, t.Class, name, *whale))
			fi, _ := os.Stat(path)
			must(exec(pg, `INSERT INTO databases (id, tenant_id, name, state, size_bytes) VALUES ($1, $2, $3, 'active', $4)`, id, t.ID, name, fi.Size()))
			must(exec(pg, `INSERT INTO placement (database_id, owner_node, epoch) VALUES ($1, $2, 1)`, id, node))
			t.DBs = append(t.DBs, dbEntry{Name: name, ID: id, Node: node})
		}
		m.Tenants = append(m.Tenants, t)
	}
	log.Printf("seeded %d tenants, %d databases in %s", len(m.Tenants), dbIndex, time.Since(start).Round(time.Millisecond))

	if *ledgerRows > 0 {
		must(exec(pg, `
			INSERT INTO usage_ledger (at, tenant_id, database_id, kind, statements, rows, bytes)
			SELECT now() - interval '30 days' * (g::float8 / $1),
			       d.tenant_id, d.id, CASE WHEN g % 4 = 0 THEN 'write' ELSE 'read' END,
			       1 + g % 20, 1 + g % 50, 100 + g % 4000
			FROM generate_series(1, $1) AS g
			JOIN (SELECT row_number() OVER (ORDER BY id) - 1 AS rn, id, tenant_id FROM databases) d
			  ON d.rn = g % $2`, *ledgerRows, dbIndex))
		must(exec(pg, `VACUUM (ANALYZE) usage_ledger`))
		log.Printf("seeded %d ledger rows in %s", *ledgerRows, time.Since(start).Round(time.Millisecond))
	}

	f, err := os.Create(*out)
	must(err)
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	must(enc.Encode(m))
	must(f.Close())
	log.Printf("wrote %s", *out)
}

func writeTenantFile(path, id, tenantID, class, name string, whale bool) error {
	db, err := sql.Open("sqlite3", "file:"+path+"?_pragma=journal_mode(wal)&_pragma=synchronous(off)")
	if err != nil {
		return err
	}
	defer db.Close()
	if err := engine.InitMeta(db, id, tenantID, 1); err != nil {
		return err
	}
	stmts := []string{
		`CREATE TABLE items (id integer PRIMARY KEY, seq integer UNIQUE, body text, created_at text)`,
		`CREATE TABLE notes (id integer PRIMARY KEY, seq integer, body text, created_at text)`,
		`WITH RECURSIVE c(x) AS (SELECT 1 UNION ALL SELECT x + 1 FROM c WHERE x < 1000)
		 INSERT INTO items (id, seq, body, created_at)
		 SELECT x, x, 'item ' || x || ' ' || hex(randomblob(16)), datetime('now', '-' || x || ' minutes') FROM c`,
	}
	if class == "whale" && name == "default" && whale {
		stmts = append(stmts,
			`CREATE TABLE events (id integer PRIMARY KEY, kind text, actor text, payload text, created_at text)`,
			`WITH RECURSIVE c(x) AS (SELECT 1 UNION ALL SELECT x + 1 FROM c WHERE x < 15000)
			 INSERT INTO events (id, kind, actor, payload, created_at)
			 SELECT x, 'kind' || (x % 7), 'user' || (x % 997), hex(randomblob(2000)), datetime('now', '-' || x || ' seconds') FROM c`)
	}
	if class == "bulky" && name == "default" {
		stmts = append(stmts,
			`CREATE TABLE attachments (id integer PRIMARY KEY, body blob)`,
			`WITH RECURSIVE c(x) AS (SELECT 1 UNION ALL SELECT x + 1 FROM c WHERE x < 1250)
			 INSERT INTO attachments (id, body) SELECT x, randomblob(4000) FROM c`)
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}
	_, err = db.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`)
	return err
}

func waitForSchema(pg *sql.DB) {
	for i := 0; ; i++ {
		var ok bool
		err := pg.QueryRow(`SELECT to_regclass('public.usage_ledger') IS NOT NULL`).Scan(&ok)
		if err == nil && ok {
			return
		}
		if i == 120 {
			log.Fatalf("registry schema not ready: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func exec(pg *sql.DB, q string, args ...any) error {
	_, err := pg.Exec(q, args...)
	return err
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func hash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// uuid derives a stable version-4-shaped UUID from a name so reseeding gives
// the same database ids.
func uuid(name string) string {
	h := hash("db:" + name)
	return h[0:8] + "-" + h[8:12] + "-4" + h[13:16] + "-a" + h[17:20] + "-" + h[20:32]
}
