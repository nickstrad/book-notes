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
	"sync/atomic"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	_ "github.com/ncruces/go-sqlite3/driver"

	"litefleet/internal/wire"
)

var (
	ErrNoSuchDatabase = errors.New("no such database")
	ErrExists         = errors.New("database already exists")
	ErrNodeFull       = errors.New("node has its maximum of open databases")
)

// Per-database pool: up to 4 statements at once, 2 connections kept warm,
// idle ones closed after a minute. Each WAL connection costs 3 fds.
const (
	maxConnsPerDB = 4
	maxIdlePerDB  = 2
	connIdleTime  = time.Minute
	fdsPerConn    = 3
)

// MaxOpenDatabases is the node-wide budget of open handles that fits the
// process's file-descriptor limit with every handle at its idle size, keeping
// headroom for sockets and the node's own files.
func MaxOpenDatabases(fdLimit int) int {
	const reserve = 124
	if fdLimit <= reserve {
		return 1
	}
	return (fdLimit - reserve) / (maxIdlePerDB * fdsPerConn)
}

// Recorder receives one usage event per statement.
type Recorder interface {
	Record(wire.UsageEvent)
}

// Engine owns the node's open databases.
type Engine struct {
	dir   string
	usage Recorder
	// MaxOpen is the node-wide budget of open handles; 0 means unbounded.
	MaxOpen int
	slots   chan struct{} // node-wide statement limit (admission.go)
	admitMu sync.Mutex
	tenants map[string]chan struct{} // each tenant's share of slots

	canaryOnce sync.Once
	canary     *sql.DB

	mu           sync.Mutex
	handles      *lru.Cache[string, *Handle] // at most HandleCache, least recently used evicted
	open         atomic.Int64                // handles not yet closed, evicted ones included
	evictedDirty map[string]bool             // evicted before their backup
}

// Handle is an open tenant database.
type Handle struct {
	ID       string
	TenantID string
	Epoch    int
	Path     string
	db       *sql.DB
	lanes    *lanes

	eng       *Engine
	refs      atomic.Int32 // callers using the handle (Handle ... Release)
	lastUsed  atomic.Int64 // unix nanoseconds
	evicted   atomic.Bool
	dropped   bool
	closeOnce sync.Once

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
	e := &Engine{dir: dir, usage: usage, slots: make(chan struct{}, NodeStatements), tenants: map[string]chan struct{}{},
		evictedDirty: map[string]bool{}}
	e.handles, _ = lru.NewWithEvict(HandleCache, e.evict)
	return e, nil
}

// Dir is the directory holding the tenant files.
func (e *Engine) Dir() string { return e.dir }

func (e *Engine) path(id string) string { return filepath.Join(e.dir, id+".db") }

// dsn opens tenant files in WAL mode. SQLite checkpoints a WAL on commit
// once it passes 1 000 pages (its default), and the node's checkpointer
// truncates WALs in the background (steady.go). A writer that finds the file
// locked waits up to 2 s with SQLite's own backing-off busy handler (it must
// be the first pragma), bounded further by the statement deadline.
func dsn(path string) string {
	return "file:" + path + "?_pragma=busy_timeout(2000)&_pragma=journal_mode(wal)&_pragma=synchronous(normal)&_pragma=foreign_keys(on)"
}

// Handle returns the open database for id, opening it on first use. The
// caller must Release it.
func (e *Engine) Handle(id string) (*Handle, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if h, ok := e.handles.Get(id); ok {
		h.use()
		return h, nil
	}
	path := e.path(id)
	if _, err := os.Stat(path); err != nil {
		return nil, ErrNoSuchDatabase
	}
	if e.MaxOpen > 0 && e.open.Load() >= int64(e.MaxOpen) {
		return nil, ErrNodeFull
	}
	db, err := sql.Open("sqlite3", dsn(path))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxConnsPerDB)
	db.SetMaxIdleConns(maxIdlePerDB)
	db.SetConnMaxIdleTime(connIdleTime)
	h := &Handle{ID: id, Path: path, db: db, lanes: newLanes(), eng: e}
	if err := h.loadMeta(); err != nil {
		db.Close()
		return nil, fmt.Errorf("open %s: %w", id, err)
	}
	e.open.Add(1)
	h.use()
	e.handles.Add(id, h)
	return h, nil
}

func (h *Handle) use() {
	h.refs.Add(1)
	h.lastUsed.Store(time.Now().UnixNano())
}

// Release ends a use of the handle; an evicted handle closes with its last
// user.
func (h *Handle) Release() {
	if h.refs.Add(-1) == 0 && h.evicted.Load() {
		h.eng.close(h)
	}
}

// evict runs (under e.mu) when the cache drops a handle: it closes now if
// nobody is using it, else when its last user releases it. A database
// written since its last backup is remembered for the backup job.
func (e *Engine) evict(id string, h *Handle) {
	if !h.dropped && h.Dirty() {
		e.evictedDirty[id] = true
	}
	h.evicted.Store(true)
	if h.refs.Load() == 0 {
		e.close(h)
	}
}

func (e *Engine) close(h *Handle) {
	h.closeOnce.Do(func() {
		h.db.Close()
		e.open.Add(-1)
	})
}

// TakeEvictedDirty returns the databases evicted before their backup.
func (e *Engine) TakeEvictedDirty() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	ids := make([]string, 0, len(e.evictedDirty))
	for id := range e.evictedDirty {
		ids = append(ids, id)
	}
	clear(e.evictedDirty)
	return ids
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
	if h, ok := e.handles.Peek(id); ok {
		h.dropped = true
		e.handles.Remove(id)
	}
	e.mu.Unlock()
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
	return e.handles.Values()
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
