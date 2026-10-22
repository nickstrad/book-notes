package engine

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ncruces/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"litefleet/internal/wire"
)

// StatementError is a tenant-visible failure of a statement.
type StatementError struct {
	Status  int
	Code    string
	Message string
}

func (e *StatementError) Error() string { return e.Code + ": " + e.Message }

// StatementDeadline is the longest a tenant statement may run. The gateway
// gives the whole request 2.9 s, so a statement that hits this deadline still
// gets its 504 back to the tenant in time.
const StatementDeadline = 2 * time.Second

// PoolWait bounds how long a statement waits for one of its database's
// connections. Past it the node answers 503 at once rather than queueing
// the request behind the ones holding them.
const PoolWait = 250 * time.Millisecond

// ErrPoolTimeout means every connection to the database is in use.
var ErrPoolTimeout = &StatementError{http.StatusServiceUnavailable, "pool_timeout", "database busy: no connection free"}

var poolTimeouts = promauto.NewCounter(prometheus.CounterOpts{
	Name: "sqlite_pool_timeouts_total", Help: "Statements refused because no connection to their database came free within PoolWait.",
})

// conn checks out a connection, waiting at most PoolWait. The caller must
// Close it.
func (h *Handle) conn(ctx context.Context) (*sql.Conn, error) {
	wait, cancel := context.WithTimeout(ctx, PoolWait)
	defer cancel()
	c, err := h.db.Conn(wait)
	if err != nil {
		if ctx.Err() == nil && wait.Err() != nil {
			poolTimeouts.Inc()
			return nil, ErrPoolTimeout
		}
		return nil, classify(ctx, err)
	}
	return c, nil
}

var sqliteBusy = promauto.NewCounter(prometheus.CounterOpts{
	Name: "sqlite_busy_total", Help: "Statements that gave up waiting for a database's write lock.",
})

var statementsCancelled = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "statements_cancelled_total", Help: "Statements interrupted because their deadline passed or their caller went away.",
}, []string{"reason"})

// classify turns a SQLite error into the typed error tenants see. A
// statement whose context ended was interrupted by the driver
// (sqlite3_interrupt) wherever it was.
func classify(ctx context.Context, err error) error {
	switch ctx.Err() {
	case context.DeadlineExceeded:
		statementsCancelled.WithLabelValues("deadline").Inc()
		return &StatementError{http.StatusGatewayTimeout, "statement_deadline", "statement exceeded the " + StatementDeadline.String() + " deadline"}
	case context.Canceled:
		statementsCancelled.WithLabelValues("caller_gone").Inc()
		return &StatementError{499, "canceled", "caller went away"}
	}
	var se *sqlite3.Error
	if !errors.As(err, &se) {
		return err
	}
	switch se.Code() {
	case sqlite3.BUSY:
		sqliteBusy.Inc()
		return &StatementError{http.StatusServiceUnavailable, "busy", "database busy: the write lock stayed taken past busy_timeout"}
	case sqlite3.CONSTRAINT:
		return &StatementError{http.StatusConflict, "constraint", se.Error()}
	case sqlite3.ERROR, sqlite3.READONLY, sqlite3.RANGE, sqlite3.MISMATCH, sqlite3.TOOBIG:
		return &StatementError{http.StatusBadRequest, "bad_request", se.Error()}
	}
	return err
}

// isRead reports whether sql is a single read-only statement.
func isRead(q string) bool {
	q = strings.TrimSpace(q)
	q = strings.TrimSuffix(q, ";")
	if strings.Contains(q, ";") {
		return false
	}
	first, _, _ := strings.Cut(strings.ToUpper(q), " ")
	switch strings.TrimSpace(first) {
	case "SELECT", "WITH", "VALUES", "EXPLAIN":
		return true
	}
	return false
}

// A query's answer is capped at MaxRows rows or MaxBytes of JSON, whichever
// comes first; a capped answer says "truncated": true. Tenants page with
// keyset queries (WHERE id > ? ORDER BY id LIMIT n). The cap is the
// platform's bulkhead for the node's memory, not a suggestion.
const (
	MaxRows  = 1000
	MaxBytes = 1 << 20
)

// Query runs one read statement on a query_only connection and streams its
// rows to the writer start returns, as they are scanned:
// {"columns":[...],"rows":[[...],...],"truncated":bool}. Nothing is written
// unless the statement starts; started reports whether the answer had begun
// when err happened (the caller can then only abandon it).
func (e *Engine) Query(ctx context.Context, id string, epoch int, req wire.QueryRequest, start func() io.Writer) (started bool, err error) {
	defer observeStatement("read", time.Now())
	ctx, cancel := context.WithTimeout(ctx, StatementDeadline)
	defer cancel()
	if !isRead(req.SQL) {
		return false, &StatementError{http.StatusBadRequest, "bad_request", "query takes a single read statement; use execute for writes"}
	}
	h, err := e.Handle(id)
	if err != nil {
		return false, err
	}
	defer h.Release()
	if err := h.admit(epoch); err != nil {
		return false, err
	}
	release, err := e.acquire(ctx, h, false)
	if err != nil {
		return false, err
	}
	defer release()

	c, err := h.conn(ctx)
	if err != nil {
		return false, err
	}
	defer c.Close()
	tx, err := c.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return false, classify(ctx, err)
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, req.SQL, req.Args...)
	if err != nil {
		return false, classify(ctx, err)
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return false, err
	}
	head, _ := json.Marshal(cols)
	w := start()
	n, _ := fmt.Fprintf(w, `{"columns":%s,`, head)
	size := int64(n)
	if e.WireVersion >= 2 { // v2 adds column types; every v1 field stays
		types := make([]string, len(cols))
		if cts, err := rows.ColumnTypes(); err == nil {
			for i, ct := range cts {
				types[i] = ct.DatabaseTypeName()
			}
		}
		b, _ := json.Marshal(types)
		n, _ = fmt.Fprintf(w, `"column_types":%s,`, b)
		size += int64(n)
	}
	n, _ = io.WriteString(w, `"rows":[`)
	size += int64(n)
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	var count int64
	truncated := false
	for rows.Next() {
		if count == MaxRows {
			truncated = true
			break
		}
		if err := rows.Scan(ptrs...); err != nil {
			return true, classify(ctx, err)
		}
		row, err := json.Marshal(vals)
		if err != nil {
			return true, err
		}
		if size+int64(len(row)) > MaxBytes {
			truncated = true
			break
		}
		if count > 0 {
			w.Write([]byte{','})
			size++
		}
		w.Write(row)
		size += int64(len(row))
		count++
	}
	if err := rows.Err(); err != nil {
		return true, classify(ctx, err)
	}
	if e.WireVersion >= 2 {
		n, _ = fmt.Fprintf(w, `],"truncated":%t,"stats":{"rows_read":%d,"bytes":%d}}`, truncated, count, size)
	} else {
		n, _ = fmt.Fprintf(w, `],"truncated":%t}`, truncated)
	}
	size += int64(n)
	h.recordRead(count, size)
	e.emit(h, "read", count, size)
	return true, nil
}

// Execute runs every statement in one write transaction.
func (e *Engine) Execute(ctx context.Context, id string, epoch int, req wire.ExecuteRequest) (*wire.ExecuteResult, error) {
	defer observeStatement("write", time.Now())
	ctx, cancel := context.WithTimeout(ctx, StatementDeadline)
	defer cancel()
	if len(req.Statements) == 0 {
		return nil, &StatementError{http.StatusBadRequest, "bad_request", "no statements"}
	}
	h, err := e.Handle(id)
	if err != nil {
		return nil, err
	}
	defer h.Release()
	if err := h.admit(epoch); err != nil {
		return nil, err
	}
	release, err := e.acquire(ctx, h, true)
	if err != nil {
		return nil, err
	}
	defer release()

	c, err := h.conn(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	tx, err := c.BeginTx(ctx, nil)
	if err != nil {
		return nil, classify(ctx, err)
	}
	// Every return path below rolls back unless Commit succeeded, which
	// releases the file's write lock and returns the connection.
	defer tx.Rollback()
	var res wire.ExecuteResult
	var inBytes int64
	affected := make([]int64, 0, len(req.Statements))
	for _, st := range req.Statements {
		r, err := tx.ExecContext(ctx, st.SQL, st.Args...)
		if err != nil {
			return nil, classify(ctx, err)
		}
		n, _ := r.RowsAffected()
		res.RowsAffected += n
		affected = append(affected, n)
		if id, err := r.LastInsertId(); err == nil {
			res.LastInsertRowID = id
		}
		inBytes += int64(len(st.SQL))
	}
	if err := tx.Commit(); err != nil {
		return nil, classify(ctx, err)
	}
	h.recordWrite(res.RowsAffected, inBytes)
	for i, n := range affected {
		e.emit(h, "write", n, int64(len(req.Statements[i].SQL)))
	}
	return &res, nil
}

func (h *Handle) recordRead(rows, bytes int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.activity.Reads++
	h.activity.ReadRows += rows
	h.activity.ReadBytes += bytes
}

func (h *Handle) recordWrite(rows, bytes int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.activity.Writes++
	h.activity.WriteRows += rows
	h.activity.WriteBytes += bytes
	h.dirty = true
}

func (e *Engine) emit(h *Handle, kind string, rows, bytes int64) {
	if e.usage == nil {
		return
	}
	e.usage.Record(wire.UsageEvent{
		EventID: eventID(), TenantID: h.TenantID, DatabaseID: h.ID,
		Kind: kind, Rows: rows, Bytes: bytes, At: time.Now().UTC().Format(time.RFC3339Nano),
	})
}

func eventID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
