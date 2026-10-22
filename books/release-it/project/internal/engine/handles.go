// Package engine runs tenant SQL against the SQLite files a node owns. Each
// database is one file in WAL mode under the node's data directory.
package engine

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"

	"litefleet/internal/wire"
)

var (
	ErrNoSuchDatabase = errors.New("no such database")
	ErrExists         = errors.New("database already exists")
)

// Recorder receives one usage event per statement.
type Recorder interface {
	Record(wire.UsageEvent)
}

// Engine owns the node's open databases.
type Engine struct {
	dir   string
	usage Recorder

	mu      sync.Mutex
	handles map[string]*Handle
}

// Handle is an open tenant database.
type Handle struct {
	ID       string
	TenantID string
	Epoch    int
	Path     string
	db       *sql.DB

	mu         sync.Mutex
	activity   wire.DatabaseReport // since the last heartbeat
	dirty      bool                // written since the last backup
	generation int
	lastBackup time.Time
}

// New returns an engine for the tenant files under dir.
func New(dir string, usage Recorder) (*Engine, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Engine{dir: dir, usage: usage, handles: map[string]*Handle{}}, nil
}

// Dir is the directory holding the tenant files.
func (e *Engine) Dir() string { return e.dir }

func (e *Engine) path(id string) string { return filepath.Join(e.dir, id+".db") }

// dsn opens tenant files in WAL mode. Automatic checkpoints are off because
// they landed on tenants' writes as latency spikes; checkpoints belong in the
// node's maintenance loop instead (TODO: schedule them).
func dsn(path string) string {
	return "file:" + path + "?_pragma=journal_mode(wal)&_pragma=synchronous(normal)&_pragma=wal_autocheckpoint(0)&_pragma=foreign_keys(on)"
}

// Handle returns the open database for id, opening it on first use.
func (e *Engine) Handle(id string) (*Handle, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if h, ok := e.handles[id]; ok {
		return h, nil
	}
	path := e.path(id)
	if _, err := os.Stat(path); err != nil {
		return nil, ErrNoSuchDatabase
	}
	db, err := sql.Open("sqlite3", dsn(path))
	if err != nil {
		return nil, err
	}
	h := &Handle{ID: id, Path: path, db: db}
	if err := h.loadMeta(); err != nil {
		db.Close()
		return nil, fmt.Errorf("open %s: %w", id, err)
	}
	e.handles[id] = h
	return h, nil
}

// DB is the handle's connection pool.
func (h *Handle) DB() *sql.DB { return h.db }

func (h *Handle) loadMeta() error {
	rows, err := h.db.Query(`SELECT key, value FROM _meta`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return err
		}
		switch k {
		case "tenant_id":
			h.TenantID = v
		case "epoch":
			h.Epoch, _ = strconv.Atoi(v)
		case "backup_generation":
			h.generation, _ = strconv.Atoi(v)
		}
	}
	return rows.Err()
}

// Provision creates a new, empty tenant database.
func (e *Engine) Provision(req wire.ProvisionRequest) error {
	path := e.path(req.DatabaseID)
	if _, err := os.Stat(path); err == nil {
		return ErrExists
	}
	db, err := sql.Open("sqlite3", dsn(path))
	if err != nil {
		return err
	}
	defer db.Close()
	return InitMeta(db, req.DatabaseID, req.TenantID, req.Epoch)
}

// InitMeta writes the platform's _meta table into a fresh database file.
func InitMeta(db *sql.DB, dbID, tenantID string, epoch int) error {
	if _, err := db.Exec(`CREATE TABLE _meta (key text PRIMARY KEY, value text NOT NULL)`); err != nil {
		return err
	}
	_, err := db.Exec(`INSERT INTO _meta (key, value) VALUES
		('database_id', ?), ('tenant_id', ?), ('schema_version', '1'), ('epoch', ?), ('backup_generation', '0')`,
		dbID, tenantID, strconv.Itoa(epoch))
	return err
}

// Drop closes and deletes a tenant database.
func (e *Engine) Drop(id string) error {
	e.mu.Lock()
	h, ok := e.handles[id]
	delete(e.handles, id)
	e.mu.Unlock()
	if ok {
		h.db.Close()
	}
	path := e.path(id)
	if _, err := os.Stat(path); err != nil {
		return ErrNoSuchDatabase
	}
	for _, suffix := range []string{"", "-wal", "-shm"} {
		if err := os.Remove(path + suffix); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// Open returns every open handle.
func (e *Engine) Open() []*Handle {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]*Handle, 0, len(e.handles))
	for _, h := range e.handles {
		out = append(out, h)
	}
	return out
}

// Report returns each open database's size and its activity since the last
// call, for the heartbeat.
func (e *Engine) Report() []wire.DatabaseReport {
	var out []wire.DatabaseReport
	for _, h := range e.Open() {
		h.mu.Lock()
		r := h.activity
		h.activity = wire.DatabaseReport{}
		if !h.lastBackup.IsZero() {
			r.LastBackup = h.lastBackup.UTC().Format(time.RFC3339)
		}
		h.mu.Unlock()
		r.ID = h.ID
		if fi, err := os.Stat(h.Path); err == nil {
			r.Bytes = fi.Size()
		}
		if fi, err := os.Stat(h.Path + "-wal"); err == nil {
			r.WALBytes = fi.Size()
			r.Bytes += fi.Size()
		}
		out = append(out, r)
	}
	return out
}

// Dirty reports whether the database was written since its last backup.
func (h *Handle) Dirty() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.dirty
}

// Generation is the last uploaded backup generation.
func (h *Handle) Generation() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.generation
}

// BackedUp records a completed backup.
func (h *Handle) BackedUp(gen int, at time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.generation, h.lastBackup, h.dirty = gen, at, false
}
