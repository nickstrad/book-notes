package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

var errBoom = errors.New("boom")

func always(error) bool { return true }

func TestRetriesUpToMax(t *testing.T) {
	p := Policy{Dep: "t", Base: time.Millisecond, Cap: 2 * time.Millisecond, MaxRetries: 2, Budget: NewBudget(1, 10)}
	calls := 0
	err := p.Do(context.Background(), always, func(context.Context) error { calls++; return errBoom })
	if err != errBoom || calls != 3 {
		t.Fatalf("calls = %d, err = %v; want 3 calls", calls, err)
	}
}

func TestBudgetCapsRetries(t *testing.T) {
	b := NewBudget(0.1, 1)
	p := Policy{Dep: "t", Base: time.Millisecond, Cap: time.Millisecond, MaxRetries: 2, Budget: b}
	retries := 0
	for i := 0; i < 100; i++ {
		first := true
		p.Do(context.Background(), always, func(context.Context) error {
			if !first {
				retries++
			}
			first = false
			return errBoom
		})
	}
	if retries > 12 {
		t.Fatalf("%d retries for 100 calls; budget allows about 10", retries)
	}
}

func TestNoRetryPastDeadline(t *testing.T) {
	p := Policy{Dep: "t", Base: 50 * time.Millisecond, Cap: 400 * time.Millisecond, MaxRetries: 2, MinAttempt: 100 * time.Millisecond, Budget: NewBudget(1, 10)}
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	calls := 0
	p.Do(ctx, always, func(context.Context) error { calls++; return errBoom })
	if calls != 1 {
		t.Fatalf("calls = %d; want 1 (no time left for a retry)", calls)
	}
}

func TestNotRetryable(t *testing.T) {
	p := Policy{Dep: "t", Base: time.Millisecond, Cap: time.Millisecond, MaxRetries: 2, Budget: NewBudget(1, 10)}
	calls := 0
	p.Do(context.Background(), func(error) bool { return false }, func(context.Context) error { calls++; return errBoom })
	if calls != 1 {
		t.Fatalf("calls = %d; want 1", calls)
	}
}
