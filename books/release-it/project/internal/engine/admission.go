package engine

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ErrStaleEpoch means the caller's route predates the database's current
// placement; it must re-resolve.
var ErrStaleEpoch = errors.New("stale epoch")

// admit decides whether a request may run on this handle. The node only
// serves requests routed with the database's current epoch.
func (h *Handle) admit(epoch int) error {
	if epoch != h.Epoch {
		return ErrStaleEpoch
	}
	return nil
}

// Bulkheads, three deep. Each database gets its own lanes: 2 statements
// reading at once, 1 writing, and at most 8 waiting for a lane; past that its
// tenant gets 429 at once. Each tenant may run at most 4 statements on the
// node across all its databases, and the node runs at most 16, so one tenant
// can hold at most a quarter of the node however hard it pushes. (The lanes
// bound concurrency, not CPU: an expensive query still takes its share of
// the processor while it runs.)
const (
	readsPerDB       = 2
	writersPerDB     = 1
	waitingPerDB     = 8
	statementsTenant = 4
	NodeStatements   = 16
	nodeWait         = 250 * time.Millisecond
)

var (
	ErrTenantBusy = &StatementError{http.StatusTooManyRequests, "tenant_busy", "too many concurrent statements on this database; retry after a second"}
	ErrNodeBusy   = &StatementError{http.StatusServiceUnavailable, "node_busy", "node at its statement limit"}

	admissionRejected = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admission_rejected_total", Help: "Statements refused by the bulkheads.",
	}, []string{"reason"})
	statementsRunning = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "statements_running", Help: "Statements holding a node slot.",
	})
)

// lanes are one database's bulkhead.
type lanes struct {
	read, write chan struct{}
	waiting     atomic.Int32
}

func newLanes() *lanes {
	return &lanes{read: make(chan struct{}, readsPerDB), write: make(chan struct{}, writersPerDB)}
}

// acquire takes a lane on the database and a slot on the node, or refuses.
// The returned release must be called when the statement ends.
func (e *Engine) acquire(ctx context.Context, h *Handle, write bool) (func(), error) {
	lane := h.lanes.read
	if write {
		lane = h.lanes.write
	}
	select {
	case lane <- struct{}{}:
	default:
		if h.lanes.waiting.Add(1) > waitingPerDB {
			h.lanes.waiting.Add(-1)
			admissionRejected.WithLabelValues("tenant_busy").Inc()
			return nil, ErrTenantBusy
		}
		select {
		case lane <- struct{}{}:
			h.lanes.waiting.Add(-1)
		case <-ctx.Done():
			h.lanes.waiting.Add(-1)
			return nil, classify(ctx, ctx.Err())
		}
	}
	wait, cancel := context.WithTimeout(ctx, nodeWait)
	defer cancel()
	tenant := e.tenantSlots(h.TenantID)
	select {
	case tenant <- struct{}{}:
	case <-wait.Done():
		<-lane
		admissionRejected.WithLabelValues("tenant_busy").Inc()
		return nil, ErrTenantBusy
	}
	select {
	case e.slots <- struct{}{}:
	case <-wait.Done():
		<-tenant
		<-lane
		admissionRejected.WithLabelValues("node_busy").Inc()
		return nil, ErrNodeBusy
	}
	statementsRunning.Inc()
	return func() {
		statementsRunning.Dec()
		<-e.slots
		<-tenant
		<-lane
	}, nil
}

// tenantSlots is the tenant's share of the node.
func (e *Engine) tenantSlots(tenantID string) chan struct{} {
	e.admitMu.Lock()
	defer e.admitMu.Unlock()
	t, ok := e.tenants[tenantID]
	if !ok {
		t = make(chan struct{}, statementsTenant)
		e.tenants[tenantID] = t
	}
	return t
}
