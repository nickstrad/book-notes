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
			if got := Backoff(tc.kind, tc.base, tc.n); got != time.Duration(tc.want)*time.Second {
				t.Fatalf("got %s", got)
			}
		})
	}
}
func TestToken(t *testing.T) {
	a, b := Token(), Token()
	if len(a) != 32 || a == b {
		t.Fatal("tokens must be distinct 128 bit values")
	}
}
