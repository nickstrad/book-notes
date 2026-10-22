package metering

import (
	"log"
	"sync"
	"time"

	"litefleet/internal/wire"
)

// Queue buffers usage events between the engine and the metering partner so
// that a statement never waits on billing.
type Queue struct {
	client *Client
	batch  int

	mu      sync.Mutex
	pending []wire.UsageEvent
}

// NewQueue returns a queue that ships to client in batches.
func NewQueue(client *Client) *Queue {
	return &Queue{client: client, batch: 50}
}

// Record enqueues one event.
func (q *Queue) Record(ev wire.UsageEvent) {
	q.mu.Lock()
	q.pending = append(q.pending, ev)
	q.mu.Unlock()
}

// Len is the number of events waiting to ship.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.pending)
}

// Run ships events until the process exits.
func (q *Queue) Run() {
	for {
		q.mu.Lock()
		n := min(len(q.pending), q.batch)
		batch := append([]wire.UsageEvent(nil), q.pending[:n]...)
		q.mu.Unlock()
		if n == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if err := q.client.Send(batch); err != nil {
			log.Printf("metering: %v", err)
			time.Sleep(time.Second)
			continue
		}
		q.mu.Lock()
		q.pending = q.pending[n:]
		q.mu.Unlock()
	}
}
