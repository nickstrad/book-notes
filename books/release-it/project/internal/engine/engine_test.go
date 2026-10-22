package engine

import (
	"context"
	"errors"
	"net/http"
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

func TestExecuteAndQuery(t *testing.T) {
	e := newEngine(t)
	exec(t, e, `CREATE TABLE items (id integer PRIMARY KEY, seq integer UNIQUE, body text)`)
	res := exec(t, e, `INSERT INTO items (seq, body) VALUES (?, ?)`, 1, "one")
	if res.RowsAffected != 1 || res.LastInsertRowID != 1 {
		t.Fatalf("execute result = %+v", res)
	}
	q, _, err := e.Query(context.Background(), "db1", 1, wire.QueryRequest{SQL: `SELECT seq, body FROM items`})
	if err != nil {
		t.Fatal(err)
	}
	if len(q.Rows) != 1 || q.Columns[1] != "body" || q.Rows[0][1] != "one" {
		t.Fatalf("query result = %+v", q)
	}
}

func TestQueryRejectsWrites(t *testing.T) {
	e := newEngine(t)
	_, _, err := e.Query(context.Background(), "db1", 1, wire.QueryRequest{SQL: `DELETE FROM _meta`})
	var se *StatementError
	if !errors.As(err, &se) || se.Status != http.StatusBadRequest {
		t.Fatalf("err = %v; want 400", err)
	}
}

func TestStaleEpoch(t *testing.T) {
	e := newEngine(t)
	if _, _, err := e.Query(context.Background(), "db1", 2, wire.QueryRequest{SQL: `SELECT 1`}); !errors.Is(err, ErrStaleEpoch) {
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
	if _, _, err := e.Query(context.Background(), "nope", 1, wire.QueryRequest{SQL: `SELECT 1`}); !errors.Is(err, ErrNoSuchDatabase) {
		t.Fatalf("err = %v; want ErrNoSuchDatabase", err)
	}
}

func TestReportCountsActivity(t *testing.T) {
	e := newEngine(t)
	exec(t, e, `CREATE TABLE items (id integer PRIMARY KEY)`)
	exec(t, e, `INSERT INTO items DEFAULT VALUES`)
	e.Query(context.Background(), "db1", 1, wire.QueryRequest{SQL: `SELECT * FROM items`})
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
	_, _, err := e.Query(context.Background(), "db1", 1, wire.QueryRequest{
		SQL: `WITH RECURSIVE c(x) AS (SELECT 1 UNION ALL SELECT x + 1 FROM c WHERE x < 1000000000) SELECT count(*) FROM c`,
	})
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
