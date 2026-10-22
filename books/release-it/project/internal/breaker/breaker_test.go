package breaker

import (
	"errors"
	"testing"
	"time"
)

var errDown = errors.New("down")

func all(error) bool { return true }

func TestOpensAfterConsecutiveFailures(t *testing.T) {
	b := New("t1")
	for i := 0; i < 5; i++ {
		b.Call(func() error { return errDown }, all)
	}
	called := false
	if err := b.Call(func() error { called = true; return nil }, all); err != ErrOpen || called {
		t.Fatalf("err = %v, called = %v; want ErrOpen without calling", err, called)
	}
}

func TestHalfOpenTrialCloses(t *testing.T) {
	b := New("t2")
	b.coolDown = 20 * time.Millisecond
	for i := 0; i < 5; i++ {
		b.Call(func() error { return errDown }, all)
	}
	time.Sleep(30 * time.Millisecond)
	if b.State() != HalfOpen {
		t.Fatalf("state = %v; want half-open", b.State())
	}
	if err := b.Call(func() error { return nil }, all); err != nil {
		t.Fatal(err)
	}
	if b.State() != Closed {
		t.Fatalf("state = %v; want closed after a good trial", b.State())
	}
}

func TestHalfOpenTrialFailureReopens(t *testing.T) {
	b := New("t3")
	b.coolDown = 20 * time.Millisecond
	for i := 0; i < 5; i++ {
		b.Call(func() error { return errDown }, all)
	}
	time.Sleep(30 * time.Millisecond)
	b.Call(func() error { return errDown }, all)
	if b.State() != Open {
		t.Fatalf("state = %v; want open after a failed trial", b.State())
	}
}

func TestAnswersDoNotCount(t *testing.T) {
	b := New("t4")
	for i := 0; i < 10; i++ {
		b.Call(func() error { return errDown }, func(error) bool { return false })
	}
	if b.State() != Closed {
		t.Fatalf("state = %v; 4xx-style answers must not open the breaker", b.State())
	}
}
