package metering

import (
	"database/sql"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"

	"litefleet/internal/wire"
)

func newTestQueue(t *testing.T) *Queue {
	t.Helper()
	db, err := sql.Open("sqlite3", "file:"+filepath.Join(t.TempDir(), "node.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	ob, err := NewOutbox(db)
	if err != nil {
		t.Fatal(err)
	}
	return NewQueue(nil, ob) // no shippers running: nothing drains
}

func event(i int) wire.UsageEvent {
	return wire.UsageEvent{EventID: strconv.Itoa(i), At: time.Now().UTC().Format(time.RFC3339Nano)}
}

func TestFullQueueOverflowsToOutbox(t *testing.T) {
	q := newTestQueue(t)
	for i := 0; i < Capacity; i++ {
		q.Record(event(i))
	}
	start := time.Now()
	q.Record(event(Capacity))
	if d := time.Since(start); d < EnqueueBudget || d > EnqueueBudget+50*time.Millisecond {
		t.Fatalf("a full queue made the statement wait %s; want about %s", d, EnqueueBudget)
	}
	if q.Len() != Capacity || q.OutboxLen() != 1 {
		t.Fatalf("queue %d, outbox %d; want %d and 1", q.Len(), q.OutboxLen(), Capacity)
	}
}

func TestSpillKeepsEverything(t *testing.T) {
	q := newTestQueue(t)
	for i := 0; i < 10; i++ {
		q.Record(event(i))
	}
	q.Spill(0)
	if q.OutboxLen() != 10 {
		t.Fatalf("outbox %d after spill; want 10", q.OutboxLen())
	}
}
