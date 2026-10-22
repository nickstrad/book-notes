package metering

import (
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"litefleet/internal/wire"
)

// Back pressure for usage events. The queue holds at most Capacity events.
// A statement that finds it full waits up to EnqueueBudget for room, then
// writes its event to the outbox instead, so a slow metering partner costs
// a statement at most a few milliseconds and never memory or lost events.
// Shippers send batches of up to Batch; a batch that fails goes to the
// outbox; the sweeper drains the outbox at a bounded rate.
const (
	Capacity      = 1024
	EnqueueBudget = 20 * time.Millisecond
	Shippers      = 4
	Batch         = 100
	SweepEvery    = 5 * time.Second
	SweepLimit    = 500
)

var (
	enqueueTimeouts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "usage_enqueue_timeouts_total", Help: "Usage events written to the outbox because the queue stayed full.",
	})
	sweeperShipped = promauto.NewCounter(prometheus.CounterOpts{
		Name: "usage_sweeper_shipped_total", Help: "Usage events delivered from the outbox.",
	})
)

// Queue buffers usage events between the engine and the metering partner so
// that a statement never waits on billing.
type Queue struct {
	client *Client
	outbox *Outbox
	space  chan struct{} // one token per free slot
	ready  chan struct{}

	mu       sync.Mutex
	pending  []wire.UsageEvent
	inflight map[int][]wire.UsageEvent
}

// NewQueue returns a bounded queue that ships to client and overflows to
// outbox.
func NewQueue(client *Client, outbox *Outbox) *Queue {
	return &Queue{client: client, outbox: outbox, space: make(chan struct{}, Capacity),
		ready: make(chan struct{}, 1), inflight: map[int][]wire.UsageEvent{}}
}

// Record enqueues one event, or stores it in the outbox if the queue stays
// full for EnqueueBudget.
func (q *Queue) Record(ev wire.UsageEvent) {
	select {
	case q.space <- struct{}{}:
	default:
		t := time.NewTimer(EnqueueBudget)
		defer t.Stop()
		select {
		case q.space <- struct{}{}:
		case <-t.C:
			enqueueTimeouts.Inc()
			if err := q.outbox.Add([]wire.UsageEvent{ev}); err != nil {
				log.Printf("usage outbox: %v", err)
			}
			return
		}
	}
	q.mu.Lock()
	q.pending = append(q.pending, ev)
	q.mu.Unlock()
	select {
	case q.ready <- struct{}{}:
	default:
	}
}

func (q *Queue) take(n int) []wire.UsageEvent {
	q.mu.Lock()
	k := min(n, len(q.pending))
	batch := append([]wire.UsageEvent(nil), q.pending[:k]...)
	q.pending = q.pending[k:]
	q.mu.Unlock()
	for range k {
		<-q.space
	}
	return batch
}

// Len is the number of events waiting in memory.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.pending)
}

// OldestAge is how long the oldest waiting event has waited.
func (q *Queue) OldestAge() time.Duration {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.pending) == 0 {
		return 0
	}
	at, err := time.Parse(time.RFC3339Nano, q.pending[0].At)
	if err != nil {
		return 0
	}
	return time.Since(at)
}

// OutboxLen is the number of events waiting in the outbox.
func (q *Queue) OutboxLen() int { return q.outbox.Len() }

// Run starts the shippers and the sweeper.
func (q *Queue) Run() {
	for i := range Shippers {
		go q.ship(i)
	}
	go q.sweep()
}

func (q *Queue) ship(id int) {
	for {
		batch := q.take(Batch)
		if len(batch) == 0 {
			select {
			case <-q.ready:
			case <-time.After(100 * time.Millisecond):
			}
			continue
		}
		q.mu.Lock()
		q.inflight[id] = batch
		q.mu.Unlock()
		err := q.client.Send(batch)
		q.mu.Lock()
		delete(q.inflight, id)
		q.mu.Unlock()
		if err != nil {
			log.Printf("metering: %v; %d events to the outbox", err, len(batch))
			if err := q.outbox.Add(batch); err != nil {
				log.Printf("usage outbox: %v", err)
			}
		}
	}
}

func (q *Queue) sweep() {
	for range time.Tick(SweepEvery) {
		events, err := q.outbox.Take(SweepLimit)
		if err != nil {
			log.Printf("usage outbox: %v", err)
			continue
		}
		for len(events) > 0 {
			batch := events[:min(Batch, len(events))]
			events = events[len(batch):]
			if err := q.client.Send(batch); err != nil {
				break // metering still unhappy; next sweep
			}
			q.outbox.Delete(batch)
			sweeperShipped.Add(float64(len(batch)))
		}
	}
}

// Spill keeps every event across a shutdown: what is queued goes to the
// outbox at once; batches being sent get up to wait to finish (a failed one
// goes to the outbox by itself); any still in flight after that are written
// too, and if one was delivered after all, metering drops the duplicate by
// event id. The node calls it on SIGTERM.
func (q *Queue) Spill(wait time.Duration) {
	q.spillPending()
	defer q.spillPending() // statements finishing during the wait queue events too
	for deadline := time.Now().Add(wait); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		q.mu.Lock()
		n := len(q.inflight)
		q.mu.Unlock()
		if n == 0 {
			return
		}
	}
	q.mu.Lock()
	var stuck []wire.UsageEvent
	for _, b := range q.inflight {
		stuck = append(stuck, b...)
	}
	q.mu.Unlock()
	if err := q.outbox.Add(stuck); err != nil {
		log.Printf("usage outbox: %v", err)
	}
}

func (q *Queue) spillPending() {
	q.mu.Lock()
	events := q.pending
	q.pending = nil
	q.mu.Unlock()
	for range events {
		<-q.space
	}
	if err := q.outbox.Add(events); err != nil {
		log.Printf("usage outbox: %v", err)
	}
	log.Printf("usage queue: spilled %d events to the outbox", len(events))
}
