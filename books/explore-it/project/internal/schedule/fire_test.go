package schedule

import (
	"testing"
	"time"
)

func TestNext(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		cron string
		want time.Duration
	}{{"* * * * *", time.Minute}, {"*/5 * * * *", 5 * time.Minute}, {"0 * * * *", time.Hour}} {
		t.Run(tc.cron, func(t *testing.T) {
			n, e := Next(tc.cron, "UTC", 0, base)
			if e != nil || n.Sub(base) != tc.want {
				t.Fatalf("%v %v", n, e)
			}
		})
	}
}
func TestInvalidCron(t *testing.T) {
	if _, e := Next("wrong", "UTC", 0, time.Now()); e == nil {
		t.Fatal("invalid cron accepted")
	}
}
