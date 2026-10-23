package jobs

import (
	"slices"
	"strings"
	"testing"
)

// The expectations are the public contract (SPEC and the E07 state matrix),
// written out by hand rather than derived from the table.
func TestTransitionContract(t *testing.T) {
	accepts := map[Event][]string{
		EvPromote:   {Scheduled, RetryWait},
		EvLease:     {Queued},
		EvHeartbeat: {Leased},
		EvAck:       {Leased},
		EvNack:      {Leased},
		EvExpire:    {Leased},
		EvRelease:   {Leased},
		EvCancel:    {Scheduled, Queued, Leased, RetryWait},
		EvRetry:     {Dead},
		EvPurge:     {Scheduled, Queued, Leased, RetryWait},
		EvDelete:    {Succeeded, Dead, Cancelled},
		EvRetain:    {Succeeded, Cancelled},
	}
	if got := Events(); len(got) != len(accepts) {
		t.Fatalf("table has events %v", got)
	}
	for ev, want := range accepts {
		for _, s := range States {
			if got := Targets(ev, s) != nil; got != slices.Contains(want, s) {
				t.Errorf("%s from %s accepted=%v, contract says %v", ev, s, got, !got)
			}
		}
	}
	for _, s := range []string{Succeeded, Dead, Cancelled} {
		for _, ev := range Events() {
			for _, to := range Targets(ev, s) {
				if to != Gone && !(ev == EvRetry && s == Dead) {
					t.Errorf("terminal %s leaves to %s on %s", s, to, ev)
				}
			}
		}
	}
	if !slices.Equal(Targets(EvPurge, Leased), []string{Leased}) || !slices.Equal(Targets(EvPurge, Queued), []string{Gone}) {
		t.Error("purge must mark leased jobs and delete pending ones")
	}
}

func TestGuardText(t *testing.T) {
	for _, tc := range []struct {
		got, want string
	}{
		{Guard("j.state", EvAck), "j.state IN ('leased')"},
		{Guard("state", EvPurge, Gone), "state IN ('scheduled','queued','retry_wait')"},
		{Guard("state", EvPurge, Leased), "state IN ('leased')"},
		{Guard("state", EvCancel), "state IN ('scheduled','queued','leased','retry_wait')"},
	} {
		if tc.got != tc.want {
			t.Errorf("got %q want %q", tc.got, tc.want)
		}
	}
	for _, ev := range Events() {
		if g := Guard("state", ev); strings.Contains(g, Gone) || !strings.HasPrefix(g, "state IN ('") {
			t.Errorf("%s guard %q", ev, g)
		}
	}
	defer func() {
		if recover() == nil {
			t.Error("a guard for a target the event never reaches must panic")
		}
	}()
	Guard("state", EvAck, Dead)
}
