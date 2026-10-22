// Package breaker is a circuit breaker: it stops calling a dependency that
// keeps failing, fails callers fast instead, and lets one trial call through
// after a cool-down to find out whether the dependency is back.
//
//	closed    --(5 failures in a row, or >= 50 % of >= 20 calls in 10 s)--> open
//	open      --(5 s)-------------------------------------------------------> half-open
//	half-open --(trial call succeeds)--> closed;  (trial fails) --> open
package breaker

import (
	"errors"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ErrOpen is returned without calling the dependency while the breaker is
// open (or a half-open trial is already in flight).
var ErrOpen = errors.New("circuit open")

// State of a breaker; exported as breaker_state (0 closed, 1 half-open, 2 open).
type State int

const (
	Closed State = iota
	HalfOpen
	Open
)

func (s State) String() string {
	return [...]string{"closed", "half-open", "open"}[s]
}

var (
	stateGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "breaker_state", Help: "Circuit breaker state: 0 closed, 1 half-open, 2 open.",
	}, []string{"dep"})
	transitions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "breaker_transitions_total", Help: "Circuit breaker state changes.",
	}, []string{"dep", "from", "to"})
	rejected = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "breaker_rejected_total", Help: "Calls failed fast by an open breaker.",
	}, []string{"dep"})
)

// Breaker guards one dependency.
type Breaker struct {
	dep         string
	consecutive int           // failures in a row that open the breaker
	ratio       float64       // failure ratio over the window that opens it
	minCalls    int           // calls needed in the window before the ratio counts
	window      time.Duration // ratio window
	coolDown    time.Duration // open -> half-open

	mu       sync.Mutex
	state    State
	inRow    int
	calls    int
	failures int
	winStart time.Time
	openedAt time.Time
	trial    bool
}

// New returns a closed breaker for dep with the defaults above.
func New(dep string) *Breaker {
	b := &Breaker{dep: dep, consecutive: 5, ratio: 0.5, minCalls: 20, window: 10 * time.Second, coolDown: 5 * time.Second}
	stateGauge.WithLabelValues(dep).Set(0)
	return b
}

// State is the breaker's current state.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tick(time.Now())
	return b.state
}

// Call runs fn unless the breaker is open. failure decides which errors
// count against the dependency (a 4xx answer does not).
func (b *Breaker) Call(fn func() error, failure func(error) bool) error {
	if !b.allow() {
		rejected.WithLabelValues(b.dep).Inc()
		return ErrOpen
	}
	err := fn()
	b.record(err != nil && failure(err))
	return err
}

func (b *Breaker) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tick(time.Now())
	switch b.state {
	case Open:
		return false
	case HalfOpen:
		if b.trial {
			return false
		}
		b.trial = true
	}
	return true
}

func (b *Breaker) record(failed bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	if b.state == HalfOpen {
		b.trial = false
		if failed {
			b.to(Open, now)
		} else {
			b.to(Closed, now)
		}
		return
	}
	if now.Sub(b.winStart) > b.window {
		b.winStart, b.calls, b.failures = now, 0, 0
	}
	b.calls++
	if failed {
		b.failures++
		b.inRow++
	} else {
		b.inRow = 0
	}
	if b.state == Closed && (b.inRow >= b.consecutive ||
		(b.calls >= b.minCalls && float64(b.failures) >= b.ratio*float64(b.calls))) {
		b.to(Open, now)
	}
}

// tick moves an open breaker to half-open once its cool-down has passed.
func (b *Breaker) tick(now time.Time) {
	if b.state == Open && now.Sub(b.openedAt) >= b.coolDown {
		b.to(HalfOpen, now)
	}
}

func (b *Breaker) to(s State, now time.Time) {
	transitions.WithLabelValues(b.dep, b.state.String(), s.String()).Inc()
	b.state = s
	stateGauge.WithLabelValues(b.dep).Set(float64(s))
	switch s {
	case Open:
		b.openedAt = now
	case Closed:
		b.inRow, b.calls, b.failures, b.winStart = 0, 0, 0, now
	}
}
