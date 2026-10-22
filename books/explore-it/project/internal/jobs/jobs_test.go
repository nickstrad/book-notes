package jobs

import (
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	for _, tc := range []struct {
		kind          string
		base, n, want int
	}{{"exponential", 5, 1, 5}, {"exponential", 5, 2, 10}, {"exponential", 5, 3, 20}, {"linear", 5, 1, 5}, {"linear", 5, 2, 10}, {"linear", 5, 3, 15}, {"linear", 0, 2, 0}, {"exponential", 1, 4, 8}} {
		t.Run(tc.kind+time.Duration(tc.n).String(), func(t *testing.T) {
			if got := Backoff(tc.kind, tc.base, tc.n, 0); got != time.Duration(tc.want)*time.Second {
				t.Fatalf("got %s", got)
			}
		})
	}
}
func TestBackoffSaturates(t *testing.T) {
	limit := time.Duration(MaxSeconds) * time.Second
	for _, tc := range []struct {
		kind         string
		base, n, cap int
		want         time.Duration
	}{
		{"exponential", 5, 29, MaxSeconds, 5 << 28 * time.Second},
		{"exponential", 5, 30, MaxSeconds, limit},
		{"exponential", 5, 32, MaxSeconds, limit},
		{"exponential", 5, 65, MaxSeconds, limit},
		{"exponential", 5, 10, 0, 2560 * time.Second},
		{"exponential", 5, 11, 0, time.Hour},
		{"exponential", 5, 65, 0, time.Hour},
		{"linear", 5, 1000, 0, time.Hour},
		{"exponential", 7200, 3, 0, 2 * time.Hour},
		{"exponential", MaxSeconds, 4, 0, limit},
		{"linear", MaxSeconds, 5, 0, limit},
		{"exponential", 10, 1, 1, time.Second},
		{"exponential", 5, 20, 3600, time.Hour},
		{"linear", 5, 2, 60, 10 * time.Second},
	} {
		if got := Backoff(tc.kind, tc.base, tc.n, tc.cap); got != tc.want {
			t.Errorf("%s base %d attempt %d cap %d: got %s want %s", tc.kind, tc.base, tc.n, tc.cap, got, tc.want)
		}
	}
	for n := 1; n < 100; n++ {
		if Backoff("exponential", 5, n+1, MaxSeconds) < Backoff("exponential", 5, n, MaxSeconds) {
			t.Fatalf("backoff decreases after attempt %d", n)
		}
	}
}
func TestToken(t *testing.T) {
	a, b := Token(), Token()
	if len(a) != 32 || a == b {
		t.Fatal("tokens must be distinct 128 bit values")
	}
}
