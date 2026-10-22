package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"litefleet/internal/wire"
)

func newEngine(t *testing.T) *Engine {
	t.Helper()
	e, err := New(t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := e.Provision(wire.ProvisionRequest{DatabaseID: "db1", TenantID: "t1", Epoch: 1}); err != nil {
		t.Fatal(err)
	}
	return e
}

func exec(t *testing.T, e *Engine, sql string, args ...any) *wire.ExecuteResult {
	t.Helper()
	res, err := e.Execute(context.Background(), "db1", 1, wire.ExecuteRequest{Statements: []wire.Statement{{SQL: sql, Args: args}}})
	if err != nil {
		t.Fatalf("execute %q: %v", sql, err)
	}
	return res
}

// query runs a read and decodes the streamed answer.
func query(e *Engine, id string, epoch int, sql string) (*wire.QueryResult, bool, error) {
	var buf bytes.Buffer
	_, err := e.Query(context.Background(), id, epoch, wire.QueryRequest{SQL: sql}, func() io.Writer { return &buf })
	if err != nil {
		return nil, false, err
	}
	var out struct {
		wire.QueryResult
		Truncated bool `json:"truncated"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		return nil, false, err
	}
	return &out.QueryResult, out.Truncated, nil
}

func TestExecuteAndQuery(t *testing.T) {
	e := newEngine(t)
	exec(t, e, `CREATE TABLE items (id integer PRIMARY KEY, seq integer UNIQUE, body text)`)
	res := exec(t, e, `INSERT INTO items (seq, body) VALUES (?, ?)`, 1, "one")
	if res.RowsAffected != 1 || res.LastInsertRowID != 1 {
		t.Fatalf("execute result = %+v", res)
	}
	q, _, err := query(e, "db1", 1, `SELECT seq, body FROM items`)
	if err != nil {
		t.Fatal(err)
	}
	if len(q.Rows) != 1 || q.Columns[1] != "body" || q.Rows[0][1] != "one" {
		t.Fatalf("query result = %+v", q)
	}
}

func TestQueryRejectsWrites(t *testing.T) {
	e := newEngine(t)
	_, _, err := query(e, "db1", 1, `DELETE FROM _meta`)
	var se *StatementError
	if !errors.As(err, &se) || se.Status != http.StatusBadRequest {
		t.Fatalf("err = %v; want 400", err)
	}
}

func TestStaleEpoch(t *testing.T) {
	e := newEngine(t)
	if _, _, err := query(e, "db1", 2, `SELECT 1`); !errors.Is(err, ErrStaleEpoch) {
		t.Fatalf("err = %v; want ErrStaleEpoch", err)
	}
}

func TestConstraintViolationIsConflict(t *testing.T) {
	e := newEngine(t)
	exec(t, e, `CREATE TABLE items (id integer PRIMARY KEY, seq integer UNIQUE)`)
	exec(t, e, `INSERT INTO items (seq) VALUES (1)`)
	_, err := e.Execute(context.Background(), "db1", 1, wire.ExecuteRequest{Statements: []wire.Statement{{SQL: `INSERT INTO items (seq) VALUES (1)`}}})
	var se *StatementError
	if !errors.As(err, &se) || se.Status != http.StatusConflict {
		t.Fatalf("err = %v; want 409 constraint", err)
	}
}

func TestNoSuchDatabase(t *testing.T) {
	e := newEngine(t)
	if _, _, err := query(e, "nope", 1, `SELECT 1`); !errors.Is(err, ErrNoSuchDatabase) {
		t.Fatalf("err = %v; want ErrNoSuchDatabase", err)
	}
}

func TestReportCountsActivity(t *testing.T) {
	e := newEngine(t)
	exec(t, e, `CREATE TABLE items (id integer PRIMARY KEY)`)
	exec(t, e, `INSERT INTO items DEFAULT VALUES`)
	query(e, "db1", 1, `SELECT * FROM items`)
	r := e.Report()
	if len(r) != 1 || r[0].Writes != 2 || r[0].Reads != 1 || r[0].Bytes == 0 {
		t.Fatalf("report = %+v", r)
	}
	if again := e.Report(); again[0].Writes != 0 {
		t.Fatalf("activity not reset: %+v", again)
	}
}

func TestRunawayStatementIsInterrupted(t *testing.T) {
	e := newEngine(t)
	start := time.Now()
	_, _, err := query(e, "db1", 1, `WITH RECURSIVE c(x) AS (SELECT 1 UNION ALL SELECT x + 1 FROM c WHERE x < 1000000000) SELECT count(*) FROM c`)
	var se *StatementError
	if !errors.As(err, &se) || se.Code != "statement_deadline" {
		t.Fatalf("err = %v; want statement_deadline", err)
	}
	if d := time.Since(start); d < StatementDeadline || d > StatementDeadline+250*time.Millisecond {
		t.Fatalf("interrupted after %s; want just past %s", d, StatementDeadline)
	}
}

func TestFailedWriteReleasesTheWriteLock(t *testing.T) {
	e := newEngine(t)
	exec(t, e, `CREATE TABLE items (id integer PRIMARY KEY, seq integer UNIQUE)`)
	exec(t, e, `INSERT INTO items (seq) VALUES (1)`)
	for i := 0; i < 10; i++ {
		if _, err := e.Execute(context.Background(), "db1", 1, wire.ExecuteRequest{Statements: []wire.Statement{{SQL: `INSERT INTO items (seq) VALUES (1)`}}}); err == nil {
			t.Fatal("duplicate insert succeeded")
		}
	}
	exec(t, e, `INSERT INTO items (seq) VALUES (2)`)
	h, _ := e.Handle("db1")
	if open := h.DB().Stats().OpenConnections; open > maxConnsPerDB {
		t.Fatalf("%d connections open after failed writes; want <= %d", open, maxConnsPerDB)
	}
}

func TestWriterWaitsForTheLock(t *testing.T) {
	e := newEngine(t)
	exec(t, e, `CREATE TABLE notes (id integer PRIMARY KEY, body text)`)
	h, _ := e.Handle("db1")
	lock, err := h.DB().Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lock.ExecContext(context.Background(), "BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(300 * time.Millisecond)
		lock.ExecContext(context.Background(), "COMMIT")
		lock.Close()
	}()
	start := time.Now()
	exec(t, e, `INSERT INTO notes (body) VALUES ('waited')`)
	if d := time.Since(start); d < 250*time.Millisecond {
		t.Fatalf("write finished in %s while the lock was held for 300 ms", d)
	}
}

func TestOneDatabaseCannotTakeTheNode(t *testing.T) {
	e := newEngine(t)
	slow := `WITH RECURSIVE c(x) AS (SELECT 1 UNION ALL SELECT x + 1 FROM c WHERE x < 400000) SELECT count(*) FROM c`
	done := make(chan error, readsPerDB+waitingPerDB)
	for i := 0; i < readsPerDB+waitingPerDB; i++ {
		go func() { _, _, err := query(e, "db1", 1, slow); done <- err }()
	}
	time.Sleep(100 * time.Millisecond)
	start := time.Now()
	_, _, err := query(e, "db1", 1, slow)
	if err != ErrTenantBusy {
		t.Fatalf("err = %v; want ErrTenantBusy for the 13th concurrent reader", err)
	}
	if d := time.Since(start); d > 50*time.Millisecond {
		t.Fatalf("refusal took %s; want immediate", d)
	}
	for i := 0; i < readsPerDB+waitingPerDB; i++ {
		<-done
	}
}

func TestQueryAnswersAreCapped(t *testing.T) {
	e := newEngine(t)
	exec(t, e, `CREATE TABLE big (id integer PRIMARY KEY, body text)`)
	exec(t, e, `WITH RECURSIVE c(x) AS (SELECT 1 UNION ALL SELECT x + 1 FROM c WHERE x < 3000) INSERT INTO big SELECT x, 'row ' || x FROM c`)
	q, truncated, err := query(e, "db1", 1, `SELECT * FROM big`)
	if err != nil || !truncated || len(q.Rows) != MaxRows {
		t.Fatalf("rows = %d, truncated = %v, err = %v; want %d rows, truncated", len(q.Rows), truncated, err, MaxRows)
	}
	exec(t, e, `CREATE TABLE fat (id integer PRIMARY KEY, body text)`)
	exec(t, e, `WITH RECURSIVE c(x) AS (SELECT 1 UNION ALL SELECT x + 1 FROM c WHERE x < 600) INSERT INTO fat SELECT x, hex(randomblob(2000)) FROM c`)
	q, truncated, err = query(e, "db1", 1, `SELECT * FROM fat`)
	if err != nil || !truncated || len(q.Rows) >= 600 || len(q.Rows) < 200 {
		t.Fatalf("rows = %d, truncated = %v, err = %v; want the 1 MB cap", len(q.Rows), truncated, err)
	}
}

func TestCheckpointerTruncatesTheWAL(t *testing.T) {
	e := newEngine(t)
	exec(t, e, `CREATE TABLE notes (id integer PRIMARY KEY, body text)`)
	for i := 0; i < 40; i++ {
		exec(t, e, `INSERT INTO notes (body) VALUES (hex(randomblob(2000)))`)
	}
	h, _ := e.Handle("db1")
	h.Release()
	if fi, err := os.Stat(h.Path + "-wal"); err != nil || fi.Size() < CheckpointBytes {
		t.Fatalf("WAL %v (%v); want it past %d bytes before the checkpoint", fi.Size(), err, CheckpointBytes)
	}
	e.checkpoint()
	if fi, err := os.Stat(h.Path + "-wal"); err == nil && fi.Size() != 0 {
		t.Fatalf("WAL is %d bytes after the checkpoint; want truncated", fi.Size())
	}
}

func TestHandleCacheEvictsButNeverClosesAHandleInUse(t *testing.T) {
	e := newEngine(t)
	held, _ := e.Handle("db1") // in use throughout
	for i := 0; i < HandleCache+5; i++ {
		id := "x" + strconv.Itoa(i)
		if err := e.Provision(wire.ProvisionRequest{DatabaseID: id, TenantID: "t", Epoch: 1}); err != nil {
			t.Fatal(err)
		}
		h, err := e.Handle(id)
		if err != nil {
			t.Fatal(err)
		}
		h.Release()
	}
	if n := len(e.Open()); n != HandleCache {
		t.Fatalf("%d handles cached; want %d", n, HandleCache)
	}
	if err := held.DB().Ping(); err != nil {
		t.Fatalf("an evicted handle in use was closed: %v", err)
	}
	held.Release()
	if err := held.DB().Ping(); err == nil {
		t.Fatal("an evicted handle stayed open after its last user released it")
	}
}
