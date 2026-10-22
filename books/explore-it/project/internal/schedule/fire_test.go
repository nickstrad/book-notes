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
func TestNextAcrossDST(t *testing.T) {
	at := func(s string) time.Time { v, _ := time.Parse(time.RFC3339, s); return v }
	for _, tc := range []struct{ cron, after, want string }{
		// Spring forward 2027-03-14: 02:xx does not exist in New York; run at 03:00 EDT once.
		{"30 2 * * *", "2027-03-13T12:00:00Z", "2027-03-14T07:00:00Z"},
		{"0 2 * * *", "2027-03-13T12:00:00Z", "2027-03-14T07:00:00Z"},
		{"30 2 * * *", "2027-03-14T07:00:00Z", "2027-03-15T06:30:00Z"},
		{"* 2 * * *", "2027-03-14T07:00:00Z", "2027-03-15T06:00:00Z"},
		// Fall back 2027-11-07: 01:xx occurs twice; run at the first occurrence only.
		{"30 1 * * *", "2027-11-07T04:00:00Z", "2027-11-07T05:30:00Z"},
		{"30 1 * * *", "2027-11-07T05:30:00Z", "2027-11-08T06:30:00Z"},
		{"0 1 * * *", "2027-11-07T05:00:00Z", "2027-11-08T06:00:00Z"},
		// Outside the transition hour the offset change is honoured.
		{"0 9 * * *", "2027-03-13T14:00:00Z", "2027-03-14T13:00:00Z"},
		{"0 2 * * *", "2027-03-15T06:00:00Z", "2027-03-16T06:00:00Z"},
		// Wildcard hours follow elapsed time in both directions.
		{"0 * * * *", "2027-03-14T06:00:00Z", "2027-03-14T07:00:00Z"},
		{"0 * * * *", "2027-11-07T05:00:00Z", "2027-11-07T06:00:00Z"},
	} {
		got, e := Next(tc.cron, "America/New_York", 0, at(tc.after))
		if e != nil || !got.Equal(at(tc.want)) {
			t.Errorf("%s after %s = %v %v, want %s", tc.cron, tc.after, got, e, tc.want)
		}
	}
}
func TestNextRejectsWhatTheSpecDoesNotName(t *testing.T) {
	now := time.Date(2026, 10, 22, 21, 0, 0, 0, time.UTC)
	for _, tc := range []struct{ cron, zone string }{
		{"0 0 30 2 *", "UTC"},                   // never matches (fixed hour)
		{"* * 30 2 *", "UTC"},                   // never matches (wildcard hour)
		{"0 0 31 4 *", "America/New_York"},      // never matches
		{"CRON_TZ=Asia/Tokyo 0 9 * * *", "UTC"}, // zone prefix
		{"TZ=Asia/Tokyo 0 9 * * *", "UTC"},      // zone prefix
		{"0 9 * * *", "Local"},                  // not an IANA identifier
		{"0 9 * * *", ""},                       // not an IANA identifier
		{"0 9 * * * *", "UTC"},                  // six fields
	} {
		if next, e := Next(tc.cron, tc.zone, 0, now); e == nil {
			t.Errorf("%q in %q accepted, next %v", tc.cron, tc.zone, next)
		}
	}
	if _, e := Offset("Local", now); e == nil {
		t.Error("Local accepted as a zone")
	}
}
