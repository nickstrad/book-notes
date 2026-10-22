// Package retry repeats idempotent calls with exponential backoff and full
// jitter, under a per-client budget, and never past the caller's deadline.
//
// A retry is extra load sent to a dependency that has just failed. Unbounded
// and synchronised, retries turn one slow node into a saturated one and keep
// it saturated after the original fault is gone. Hence: idempotent calls only,
// at most MaxRetries, random spacing, and a budget that caps retries at a
// fraction of real traffic.
package retry

import (
	"context"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	attempts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "retry_attempts_total", Help: "Retries made, by dependency and outcome.",
	}, []string{"dep", "outcome"})
	exhausted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "retry_budget_exhausted_total", Help: "Retries not made because the retry budget was empty.",
	}, []string{"dep"})
	skipped = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "retry_skipped_deadline_total", Help: "Retries not made because the caller's deadline was too close.",
	}, []string{"dep"})
)

// Budget is a token bucket shared by every call of one client: each call
// deposits Ratio tokens (up to Max) and each retry withdraws one, so retries
// stay below Ratio of recent calls however bad things get.
type Budget struct {
	mu     sync.Mutex
	tokens float64
	ratio  float64
	max    float64
}

// NewBudget allows retries up to ratio of calls, with a burst of max.
func NewBudget(ratio, max float64) *Budget {
	return &Budget{tokens: max, ratio: ratio, max: max}
}

func (b *Budget) deposit() {
	b.mu.Lock()
	b.tokens = min(b.max, b.tokens+b.ratio)
	b.mu.Unlock()
}

func (b *Budget) withdraw() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Policy is how one client retries.
type Policy struct {
	Dep        string
	Base, Cap  time.Duration // backoff before retry n is uniform in [0, min(Cap, Base·2^n)]
	MaxRetries int
	MinAttempt time.Duration // do not start an attempt with less time than this left
	Budget     *Budget
}

// Do calls fn, then retries it while retryable(err) holds, the budget allows
// and the context leaves time for the backoff plus an attempt.
func (p Policy) Do(ctx context.Context, retryable func(error) bool, fn func(context.Context) error) error {
	p.Budget.deposit()
	err := fn(ctx)
	for n := 1; err != nil && n <= p.MaxRetries && retryable(err); n++ {
		wait := time.Duration(rand.Int64N(int64(min(p.Cap, p.Base<<n)) + 1))
		if dl, ok := ctx.Deadline(); ok && time.Until(dl) < wait+p.MinAttempt {
			skipped.WithLabelValues(p.Dep).Inc()
			return err
		}
		if !p.Budget.withdraw() {
			exhausted.WithLabelValues(p.Dep).Inc()
			return err
		}
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return err
		}
		err = fn(ctx)
		outcome := "ok"
		if err != nil {
			outcome = "error"
		}
		attempts.WithLabelValues(p.Dep, outcome).Inc()
	}
	return err
}
